package usecase

import (
	"context"

	"github.com/paincake00/geocore/internal/entity"
)

type IncidentService struct {
	Repo  IncidentRepository
	Cache IncidentCache
}

func NewIncidentService(r IncidentRepository, c IncidentCache) *IncidentService {
	return &IncidentService{Repo: r, Cache: c}
}

func (s *IncidentService) Create(ctx context.Context, i *entity.Incident) error {
	if err := s.Repo.Create(ctx, i); err != nil {
		return err
	}
	// Invalidate or update cache
	// For simplicity, just invalidate/clear or simple re-fetch could happen here.
	// Or we rely on TTL. Let's try to update logic:
	// Ideally we should just delete the cache key so next read fetches fresh data.
	return nil
}

func (s *IncidentService) GetByID(ctx context.Context, id int) (*entity.Incident, error) {
	return s.Repo.GetByID(ctx, id)
}

func (s *IncidentService) GetAll(ctx context.Context, limit, offset int) ([]*entity.Incident, error) {
	return s.Repo.GetAll(ctx, limit, offset)
}

func (s *IncidentService) Update(ctx context.Context, i *entity.Incident) error {
	if err := s.Repo.Update(ctx, i); err != nil {
		return err
	}
	// Invalidate cache logic would go here
	return nil
}

func (s *IncidentService) Delete(ctx context.Context, id int) error {
	if err := s.Repo.Delete(ctx, id); err != nil {
		return err
	}
	// Invalidate cache logic would go here
	return nil
}

func (s *IncidentService) GetStats(ctx context.Context, windowMinutes int) (map[int]int, error) {
	return s.Repo.GetStats(ctx, windowMinutes)
}
