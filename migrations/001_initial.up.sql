CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS "postgis";

-- Инциденты (опасные зоны)
CREATE TABLE incidents (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    title VARCHAR(255) NOT NULL,
    description TEXT,
    latitude DECIMAL(10, 8) NOT NULL,
    longitude DECIMAL(11, 8) NOT NULL,
    radius DECIMAL(10, 2) NOT NULL DEFAULT 100.0, -- радиус в метрах
    is_active BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_incidents_location ON incidents USING GIST (
    ST_MakePoint(longitude, latitude)
);
CREATE INDEX idx_incidents_active ON incidents(is_active) WHERE is_active = true;

-- Проверки координат
CREATE TABLE location_checks (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id VARCHAR(255) NOT NULL,
    latitude DECIMAL(10, 8) NOT NULL,
    longitude DECIMAL(11, 8) NOT NULL,
    checked_at TIMESTAMP NOT NULL DEFAULT NOW(),
    webhook_sent BOOLEAN NOT NULL DEFAULT false
);

CREATE INDEX idx_location_checks_user_time ON location_checks(user_id, checked_at);
CREATE INDEX idx_location_checks_time ON location_checks(checked_at);

-- Связь проверок с инцидентами (многие ко многим)
CREATE TABLE location_check_incidents (
    location_check_id UUID REFERENCES location_checks(id) ON DELETE CASCADE,
    incident_id UUID REFERENCES incidents(id) ON DELETE CASCADE,
    PRIMARY KEY (location_check_id, incident_id)
);

CREATE INDEX idx_lci_location_check ON location_check_incidents(location_check_id);
CREATE INDEX idx_lci_incident ON location_check_incidents(incident_id);