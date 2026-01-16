CREATE TABLE incidents (
    id SERIAL PRIMARY KEY,
    title TEXT NOT NULL,
    description TEXT,
    latitude DOUBLE PRECISION NOT NULL,
    longitude DOUBLE PRECISION NOT NULL,
    radius_meters INTEGER NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_incidents_lat_lon ON incidents (latitude, longitude);

CREATE TABLE location_checks (
    id SERIAL PRIMARY KEY,
    user_id TEXT NOT NULL,
    latitude DOUBLE PRECISION NOT NULL,
    longitude DOUBLE PRECISION NOT NULL,
    checked_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_location_checks_user_id ON location_checks (user_id);
CREATE INDEX idx_location_checks_created_at ON location_checks (checked_at);

CREATE TABLE location_check_incidents (
    location_check_id INTEGER NOT NULL REFERENCES location_checks(id) ON DELETE CASCADE,
    incident_id INTEGER NOT NULL REFERENCES incidents(id) ON DELETE CASCADE,
    PRIMARY KEY (location_check_id, incident_id)
);
