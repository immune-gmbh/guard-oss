CREATE TABLE v2.expired_appraisals (
  id BIGINT PRIMARY KEY
);

GRANT SELECT, INSERT ON TABLE v2.expired_appraisals TO apisrv;
