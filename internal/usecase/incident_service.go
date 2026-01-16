package usecase

import (
	"context"

	"github.com/paincake00/geocore/internal/entity"
)

// IncidentService отвечает за бизнес-логику управления инцидентами.
type IncidentService struct {
	Repo  IncidentRepository
	Cache IncidentCache
}

// NewIncidentService создает новый экземпляр сервиса инцидентов.
func NewIncidentService(r IncidentRepository, c IncidentCache) *IncidentService {
	return &IncidentService{Repo: r, Cache: c}
}

// Create создает новый инцидент.
func (s *IncidentService) Create(ctx context.Context, i *entity.Incident) error {
	if err := s.Repo.Create(ctx, i); err != nil {
		return err
	}
	// Инвалидация или обновление кеша
	// Для простоты, здесь можно было бы сбросить кеш, чтобы при следующем чтении данные обновились.
	return nil
}

// GetByID возвращает инцидент по его ID.
func (s *IncidentService) GetByID(ctx context.Context, id int) (*entity.Incident, error) {
	return s.Repo.GetByID(ctx, id)
}

// GetAll возвращает список инцидентов с пагинацией.
func (s *IncidentService) GetAll(ctx context.Context, limit, offset int) ([]*entity.Incident, error) {
	return s.Repo.GetAll(ctx, limit, offset)
}

// Update обновляет существующий инцидент.
func (s *IncidentService) Update(ctx context.Context, i *entity.Incident) error {
	if err := s.Repo.Update(ctx, i); err != nil {
		return err
	}
	// Логика инвалидации кеша должна быть здесь
	return nil
}

// Delete удаляет инцидент по ID (или помечает удаленным).
func (s *IncidentService) Delete(ctx context.Context, id int) error {
	if err := s.Repo.Delete(ctx, id); err != nil {
		return err
	}
	// Логика инвалидации кеша должна быть здесь
	return nil
}

// GetStats возвращает статистику: сколько пользователей попало в опасные зоны за последние N минут.
func (s *IncidentService) GetStats(ctx context.Context, windowMinutes int) (map[int]int, error) {
	return s.Repo.GetStats(ctx, windowMinutes)
}
