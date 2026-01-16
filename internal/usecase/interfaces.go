package usecase

import (
	"context"

	"github.com/paincake00/geocore/internal/entity"
)

// IncidentRepository интерфейс для работы с хранилищем инцидентов (PostgreSQL).
type IncidentRepository interface {
	Create(ctx context.Context, incident *entity.Incident) error
	GetByID(ctx context.Context, id int) (*entity.Incident, error)
	GetAll(ctx context.Context, limit, offset int) ([]*entity.Incident, error)
	GetAllActive(ctx context.Context) ([]*entity.Incident, error) // Для кеширования
	Update(ctx context.Context, incident *entity.Incident) error
	Delete(ctx context.Context, id int) error
	GetStats(ctx context.Context, windowMinutes int) (map[int]int, error) // incident_id -> количество пользователей
}

// LocationCheckRepository интерфейс для сохранения проверок местоположения.
type LocationCheckRepository interface {
	CreateLocationCheck(ctx context.Context, check *entity.LocationCheck) error
	RecordIncidentMatch(ctx context.Context, checkID, incidentID int) error
}

// QueueRepository интерфейс для работы с очередью задач (Redis).
type QueueRepository interface {
	Enqueue(ctx context.Context, task string, payload interface{}) error
	Dequeue(ctx context.Context, task string) (string, error) // Возвращает JSON полезной нагрузки
}

// IncidentCache интерфейс для кеширования инцидентов (Redis).
type IncidentCache interface {
	SetIncidents(ctx context.Context, incidents []*entity.Incident) error
	GetIncidents(ctx context.Context) ([]*entity.Incident, error)
}
