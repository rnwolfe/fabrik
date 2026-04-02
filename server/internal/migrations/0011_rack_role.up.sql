ALTER TABLE racks ADD COLUMN role TEXT NOT NULL DEFAULT 'compute'
    CHECK (role IN ('compute', 'base'));
