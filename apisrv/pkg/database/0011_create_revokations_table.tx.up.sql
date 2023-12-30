CREATE TABLE v2.revokations (
  id BIGSERIAL PRIMARY KEY,

  kid TEXT NULL CHECK (kid <> ''),

  tid TEXT NULL CHECK (tid <> ''),

  expiry TIMESTAMP NOT NULL,

  CHECK (kid IS NOT NULL OR tid IS NOT NULL)
);

-- Start device primary keys at 1000
ALTER SEQUENCE v2.revokations_id_seq RESTART WITH 1000;

--
-- Index by expiry, tid, kid
--
CREATE INDEX revokations_kid_index ON v2.revokations (kid);
CREATE INDEX revokations_tid_index ON v2.revokations (tid);
CREATE INDEX revokations_expiry_index ON v2.revokations (expiry);

GRANT SELECT, INSERT, DELETE ON TABLE v2.revokations TO apisrv;
GRANT USAGE, SELECT ON ALL SEQUENCES IN SCHEMA v2 TO apisrv;
