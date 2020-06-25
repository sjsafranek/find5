
-- Enable pgcrypto for passwords
CREATE EXTENSION IF NOT EXISTS pgcrypto;

-- Enable PostGIS (includes raster)
CREATE EXTENSION IF NOT EXISTS postgis;

-- Enable Topology
CREATE EXTENSION IF NOT EXISTS postgis_topology;
