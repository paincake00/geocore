package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/paincake00/geocore/internal/entity"
)

type PostgresRepo struct {
	Pool *pgxpool.Pool
}

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

func (r *PostgresRepo) Close() {
	r.Pool.Close()
}

func (r *PostgresRepo) Ping(ctx context.Context) error {
	return r.Pool.Ping(ctx)
}

// Incident Repository

func (r *PostgresRepo) Create(ctx context.Context, i *entity.Incident) error {
	sql := `INSERT INTO incidents (title, description, latitude, longitude, radius_meters, created_at)
			VALUES ($1, $2, $3, $4, $5, NOW()) RETURNING id, created_at`
	return r.Pool.QueryRow(ctx, sql, i.Title, i.Description, i.Latitude, i.Longitude, i.RadiusMeters).Scan(&i.ID, &i.CreatedAt)
}

func (r *PostgresRepo) GetByID(ctx context.Context, id int) (*entity.Incident, error) {
	sql := `SELECT id, title, description, latitude, longitude, radius_meters, created_at FROM incidents WHERE id = $1`
	var i entity.Incident
	err := r.Pool.QueryRow(ctx, sql, id).Scan(&i.ID, &i.Title, &i.Description, &i.Latitude, &i.Longitude, &i.RadiusMeters, &i.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &i, nil
}

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

// GetAllActive returns all incidents. In a real system you might filter by "active" status or expiry.
// For this task, "active" is simply all incidents in the table implies "current" danger zones.
func (r *PostgresRepo) GetAllActive(ctx context.Context) ([]*entity.Incident, error) {
	// Assuming max safe limit or simply fetching all for syncing cache.
	// If table is huge, this is bad, but for "Nerve" core logic it implies fetching zones to check.
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

func (r *PostgresRepo) GetStats(ctx context.Context, windowMinutes int) (map[int]int, error) {
	// Return for each zone quantity of unique users for the last N minutes
	startTime := time.Now().Add(-time.Duration(windowMinutes) * time.Minute)

	// Join location_check_incidents with location_checks to filter by time
	// distinct user_id per incident
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

func (r *PostgresRepo) CreateCheck(ctx context.Context, check *entity.LocationCheck) error {
	sql := `INSERT INTO location_checks (user_id, latitude, longitude, checked_at)
            VALUES ($1, $2, $3, NOW()) RETURNING id, checked_at`
	return r.Pool.QueryRow(ctx, sql, check.UserID, check.Latitude, check.Longitude).Scan(&check.ID, &check.CheckedAt)
}

func (r *PostgresRepo) RecordIncidentMatch(ctx context.Context, checkID, incidentID int) error {
	sql := `INSERT INTO location_check_incidents (location_check_id, incident_id) VALUES ($1, $2)`
	_, err := r.Pool.Exec(ctx, sql, checkID, incidentID)
	return err
}

// Adapter methods to satisfy interfaces if signatures differ slightly (they match here)
func (r *PostgresRepo) CreateLocationCheck(ctx context.Context, check *entity.LocationCheck) error {
	return r.CreateCheck(ctx, check)
}
