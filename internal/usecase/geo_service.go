package usecase

import (
	"context"
	"log"
	"math"
	"time"

	"github.com/paincake00/geocore/internal/entity"
)

type GeoService struct {
	IncidentRepo IncidentRepository
	LocationRepo LocationCheckRepository
	Queue        QueueRepository
	Cache        IncidentCache
	QueueName    string
}

func NewGeoService(ir IncidentRepository, lr LocationCheckRepository, q QueueRepository, c IncidentCache) *GeoService {
	return &GeoService{
		IncidentRepo: ir,
		LocationRepo: lr,
		Queue:        q,
		Cache:        c,
		QueueName:    "webhook_tasks", // define queue name here
	}
}

// Haversine formula to calculate distance in meters
func distanceMeters(lat1, lon1, lat2, lon2 float64) float64 {
	const R = 6371000 // Earth radius in meters
	phi1 := lat1 * math.Pi / 180
	phi2 := lat2 * math.Pi / 180
	deltaPhi := (lat2 - lat1) * math.Pi / 180
	deltaLambda := (lon2 - lon1) * math.Pi / 180

	a := math.Sin(deltaPhi/2)*math.Sin(deltaPhi/2) +
		math.Cos(phi1)*math.Cos(phi2)*
			math.Sin(deltaLambda/2)*math.Sin(deltaLambda/2)
	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))

	return R * c
}

func (s *GeoService) CheckLocation(ctx context.Context, userID string, lat, lon float64) ([]*entity.Incident, error) {
	// 1. Get Active Incidents (Try Cache, then DB)
	var incidents []*entity.Incident
	var err error

	incidents, err = s.Cache.GetIncidents(ctx)
	if err != nil || incidents == nil {
		// Cache miss or error
		incidents, err = s.IncidentRepo.GetAllActive(ctx)
		if err != nil {
			return nil, err
		}
		// Populate cache
		_ = s.Cache.SetIncidents(ctx, incidents)
	}

	// 2. Filter incidents by distance
	var matches []*entity.Incident
	for _, i := range incidents {
		dist := distanceMeters(lat, lon, i.Latitude, i.Longitude)
		if dist <= float64(i.RadiusMeters) {
			matches = append(matches, i)
		}
	}

	// 3. Async processing (DB log + Queue)
	// We detach from the incoming context to ensure async operations complete
	// independent of the HTTP request cancellation.
	go func(uID string, latitude, longitude float64, found []*entity.Incident) {
		asyncCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		check := &entity.LocationCheck{
			UserID:    uID,
			Latitude:  latitude,
			Longitude: longitude,
		}

		if err := s.LocationRepo.CreateLocationCheck(asyncCtx, check); err != nil {
			log.Printf("Failed to create location check: %v", err)
			return
		}

		for _, incident := range found {
			// Record match
			if err := s.LocationRepo.RecordIncidentMatch(asyncCtx, check.ID, incident.ID); err != nil {
				log.Printf("Failed to record incident match: %v", err)
				continue
			}

			// Enqueue Task
			payload := entity.WebhookEvent{
				Event:                "danger_zone_detected",
				UserID:               uID,
				IncidentID:           incident.ID,
				IncidentLatitude:     incident.Latitude,
				IncidentLongitude:    incident.Longitude,
				IncidentRadiusMeters: incident.RadiusMeters,
				DetectedAt:           time.Now().Format(time.RFC3339),
			}

			if err := s.Queue.Enqueue(asyncCtx, s.QueueName, payload); err != nil {
				log.Printf("Failed to enqueue webhook task: %v", err)
			}
		}
	}(userID, lat, lon, matches)

	return matches, nil
}
