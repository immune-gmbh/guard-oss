CREATE TYPE v2.organizations_feature AS ENUM ();

CREATE TABLE v2.organizations (
  id BIGSERIAL PRIMARY KEY,
  external TEXT NOT NULL CHECK (external <> ''),

  devices INT NOT NULL CHECK (devices >= 0),
  features v2.organizations_feature[],

  updated_at TIMESTAMP NOT NULL
);

ALTER SEQUENCE v2.organizations_id_seq RESTART WITH 1000;

CREATE UNIQUE INDEX organizations_external_index ON v2.organizations (external);

GRANT SELECT, UPDATE, INSERT, DELETE ON TABLE v2.organizations TO apisrv;
GRANT USAGE, SELECT ON ALL SEQUENCES IN SCHEMA v2 TO apisrv;
