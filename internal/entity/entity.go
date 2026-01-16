package entity

import "time"

type Incident struct {
	ID           int       `json:"id"`
	Title        string    `json:"title"`
	Description  string    `json:"description"`
	Latitude     float64   `json:"latitude"`
	Longitude    float64   `json:"longitude"`
	RadiusMeters int       `json:"radius_meters"`
	CreatedAt    time.Time `json:"created_at"`
}

type LocationCheck struct {
	ID        int       `json:"id"`
	UserID    string    `json:"user_id"`
	Latitude  float64   `json:"latitude"`
	Longitude float64   `json:"longitude"`
	CheckedAt time.Time `json:"checked_at"`
}

type WebhookEvent struct {
	Event                string  `json:"event"`
	UserID               string  `json:"user_id"`
	IncidentID           int     `json:"incident_id"`
	IncidentLatitude     float64 `json:"incident_latitude"`
	IncidentLongitude    float64 `json:"incident_longitude"`
	IncidentRadiusMeters int     `json:"incident_radius_meters"`
	DetectedAt           string  `json:"detected_at"`
}
