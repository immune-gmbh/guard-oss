CREATE TABLE v2.issues_appraisals (
    appraisal_id BIGINT REFERENCES v2.appraisals NOT NULL,
    issue_type TEXT NOT NULL,
    incident boolean NOT NULL,
    payload jsonb NOT NULL
);


CREATE UNIQUE INDEX issues_appraisals_id_type ON v2.issues_appraisals USING btree (appraisal_id, issue_type);


GRANT SELECT,
    INSERT ON TABLE v2.issues_appraisals TO apisrv;