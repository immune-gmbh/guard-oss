ALTER TABLE v2.devices
ADD COLUMN attributes HSTORE NOT NULL CHECK (NOT attributes ? '') DEFAULT hstore('');