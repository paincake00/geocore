package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/paincake00/geocore/internal/entity"
)

// PostgresRepo реализация репозитория на основе PostgreSQL.
type PostgresRepo struct {
	Pool *pgxpool.Pool
}

// New создает новое подключение к PostgreSQL.
func New(dsn string) (*PostgresRepo, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	config, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, fmt.Errorf("unable to parse dsn: %w", err)
	}

	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		return nil, fmt.Errorf("unable to connect to database: %w", err)
	}

	if err := pool.Ping(ctx); err != nil {
		return nil, fmt.Errorf("ping failed: %w", err)
	}

	return &PostgresRepo{Pool: pool}, nil
}

// Close закрывает пул соединений.
func (r *PostgresRepo) Close() {
	r.Pool.Close()
}

// Ping проверяет соединение с БД.
func (r *PostgresRepo) Ping(ctx context.Context) error {
	return r.Pool.Ping(ctx)
}

// Incident Repository

// Create сохраняет новый инцидент в БД.
func (r *PostgresRepo) Create(ctx context.Context, i *entity.Incident) error {
	sql := `INSERT INTO incidents (title, description, latitude, longitude, radius_meters, created_at)
			VALUES ($1, $2, $3, $4, $5, NOW()) RETURNING id, created_at`
	return r.Pool.QueryRow(ctx, sql, i.Title, i.Description, i.Latitude, i.Longitude, i.RadiusMeters).Scan(&i.ID, &i.CreatedAt)
}

// GetByID получает инцидент по ID.
func (r *PostgresRepo) GetByID(ctx context.Context, id int) (*entity.Incident, error) {
	sql := `SELECT id, title, description, latitude, longitude, radius_meters, created_at FROM incidents WHERE id = $1`
	var i entity.Incident
	err := r.Pool.QueryRow(ctx, sql, id).Scan(&i.ID, &i.Title, &i.Description, &i.Latitude, &i.Longitude, &i.RadiusMeters, &i.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &i, nil
}

// GetAll получает список инцидентов с пагинацией.
func (r *PostgresRepo) GetAll(ctx context.Context, limit, offset int) ([]*entity.Incident, error) {
	sql := `SELECT id, title, description, latitude, longitude, radius_meters, created_at FROM incidents ORDER BY created_at DESC LIMIT $1 OFFSET $2`
	rows, err := r.Pool.Query(ctx, sql, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var incidents []*entity.Incident
	for rows.Next() {
		var i entity.Incident
		if err := rows.Scan(&i.ID, &i.Title, &i.Description, &i.Latitude, &i.Longitude, &i.RadiusMeters, &i.CreatedAt); err != nil {
			return nil, err
		}
		incidents = append(incidents, &i)
	}
	return incidents, nil
}

// GetAllActive возвращает все инциденты.
// В реальной системе стоит фильтровать по статусу "active" или времени истечения.
// В рамках задачи считаем все записи в таблице активными.
func (r *PostgresRepo) GetAllActive(ctx context.Context) ([]*entity.Incident, error) {
	sql := `SELECT id, title, description, latitude, longitude, radius_meters, created_at FROM incidents`
	rows, err := r.Pool.Query(ctx, sql)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var incidents []*entity.Incident
	for rows.Next() {
		var i entity.Incident
		if err := rows.Scan(&i.ID, &i.Title, &i.Description, &i.Latitude, &i.Longitude, &i.RadiusMeters, &i.CreatedAt); err != nil {
			return nil, err
		}
		incidents = append(incidents, &i)
	}
	return incidents, nil
}

// Update обновляет данные инцидента.
func (r *PostgresRepo) Update(ctx context.Context, i *entity.Incident) error {
	sql := `UPDATE incidents SET title=$1, description=$2, latitude=$3, longitude=$4, radius_meters=$5 WHERE id=$6`
	ct, err := r.Pool.Exec(ctx, sql, i.Title, i.Description, i.Latitude, i.Longitude, i.RadiusMeters, i.ID)
	if err != nil {
		return err
	}
	if ct.RowsAffected() == 0 {
		return errors.New("not found")
	}
	return nil
}

// Delete удаляет инцидент.
func (r *PostgresRepo) Delete(ctx context.Context, id int) error {
	sql := `DELETE FROM incidents WHERE id=$1`
	ct, err := r.Pool.Exec(ctx, sql, id)
	if err != nil {
		return err
	}
	if ct.RowsAffected() == 0 {
		return errors.New("not found")
	}
	return nil
}

// GetStats возвращает статистику: уникальные пользователи на инцидент за период.
func (r *PostgresRepo) GetStats(ctx context.Context, windowMinutes int) (map[int]int, error) {
	// Возвращаем количество уникальных пользователей на каждый инцидент за последние N минут
	startTime := time.Now().Add(-time.Duration(windowMinutes) * time.Minute)

	sql := `
    SELECT lci.incident_id, COUNT(DISTINCT lc.user_id)
    FROM location_check_incidents lci
    JOIN location_checks lc ON lci.location_check_id = lc.id
    WHERE lc.checked_at >= $1
    GROUP BY lci.incident_id
    `

	rows, err := r.Pool.Query(ctx, sql, startTime)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	stats := make(map[int]int)
	for rows.Next() {
		var incidentID, count int
		if err := rows.Scan(&incidentID, &count); err != nil {
			return nil, err
		}
		stats[incidentID] = count
	}
	return stats, nil
}

// LocationCheck Repository

// CreateCheck создает запись о проверке местоположения.
func (r *PostgresRepo) CreateCheck(ctx context.Context, check *entity.LocationCheck) error {
	sql := `INSERT INTO location_checks (user_id, latitude, longitude, checked_at)
            VALUES ($1, $2, $3, NOW()) RETURNING id, checked_at`
	return r.Pool.QueryRow(ctx, sql, check.UserID, check.Latitude, check.Longitude).Scan(&check.ID, &check.CheckedAt)
}

// RecordIncidentMatch фиксирует факт попадания проверки в инцидент.
func (r *PostgresRepo) RecordIncidentMatch(ctx context.Context, checkID, incidentID int) error {
	sql := `INSERT INTO location_check_incidents (location_check_id, incident_id) VALUES ($1, $2)`
	_, err := r.Pool.Exec(ctx, sql, checkID, incidentID)
	return err
}

// CreateLocationCheck адаптер для интерфейса.
func (r *PostgresRepo) CreateLocationCheck(ctx context.Context, check *entity.LocationCheck) error {
	return r.CreateCheck(ctx, check)
}
