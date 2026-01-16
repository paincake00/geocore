package usecase

import (
	"context"

	"github.com/paincake00/geocore/internal/entity"
)

type IncidentRepository interface {
	Create(ctx context.Context, incident *entity.Incident) error
	GetByID(ctx context.Context, id int) (*entity.Incident, error)
	GetAll(ctx context.Context, limit, offset int) ([]*entity.Incident, error)
	GetAllActive(ctx context.Context) ([]*entity.Incident, error) // For caching
	Update(ctx context.Context, incident *entity.Incident) error
	Delete(ctx context.Context, id int) error
	GetStats(ctx context.Context, windowMinutes int) (map[int]int, error) // incident_id -> user_count
}

type LocationCheckRepository interface {
	CreateLocationCheck(ctx context.Context, check *entity.LocationCheck) error
	RecordIncidentMatch(ctx context.Context, checkID, incidentID int) error
}

type QueueRepository interface {
	Enqueue(ctx context.Context, task string, payload interface{}) error
	Dequeue(ctx context.Context, task string) (string, error) // Returns payload JSON
}

type IncidentCache interface {
	SetIncidents(ctx context.Context, incidents []*entity.Incident) error
	GetIncidents(ctx context.Context) ([]*entity.Incident, error)
}
