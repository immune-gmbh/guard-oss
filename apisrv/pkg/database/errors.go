package database

import (
	"context"
	"errors"
	"sort"

	"github.com/georgysavva/scany/pgxscan"
	"github.com/jackc/pgconn"
	"github.com/jackc/pgerrcode"
)

var (
	ErrNotFound               = errors.New("not found")
	ErrConnection             = errors.New("db unreachable")
	ErrTooManyRetries         = errors.New("too many retries")
	ErrSerialization          = errors.New("serialization failure")
	ErrInvalidTransactionMode = errors.New("invalid transaction mode") // this includes the case when DB is in RO mode

	pkeyConstraints = []string{
		"organizations_pkey",
		"devices_pkey",
		"keys_pkey",
		"policies_pkey",
		"constraints_pkey",
		"appraisals_pkey",
		"gue_jobs_pkey",
		"changes_pkey",
		"revokations_pkey",
		"expired_appraisals_pkey",
	}
	fkeyConstraints = []string{
		"devices_organization_id_fkey",
		"devices_replaced_by_fkey",
		"constraints_policy_id_fkey",
		"keys_device_id_fkey",
		"policies_replaced_by_fkey",
		"policies_organization_id_fkey",
		"devices_policies_device_id_fkey",
		"devices_policies_policy_id_fkey",
		"appraisals_key_id_fkey",
		"changes_organization_id_fkey",
		"changes_device_id_fkey",
		"changes_policy_id_fkey",
		"changes_constraint_id_fkey",
		"changes_key_id_fkey",
	}
	checkConstraints = []string{
		"yes_or_no_check",
		"organizations_external_check",
		"organizations_devices_check",
		"devices_name_check",
		"devices_attributes_check",
		"devices_hwid_check",
		"devices_fpr_check",
		"devices_check",
		"keys_public_check",
		"keys_name_check",
		"keys_fpr_check",
		"keys_credential_check",
		"policies_name_check",
		"policies_check",
		"policies_pcr_template_check",
		"policies_fw_template_check",
		"policies_check1",
		"constraints_pcr_values_check",
		"appraisals_evidence_check",
		"appraisals_report_check",
		"changes_actor_check",
		"changes_comment_check",
		"changes_check",
		"verdict_typed",
		"revokations_kid_check",
		"revokations_tid_check",
		"revokations_check",
		"appraisals_check",
		"received_before_appraised",
	}
)

func init() {
	sort.Strings(pkeyConstraints)
	sort.Strings(fkeyConstraints)
	sort.Strings(checkConstraints)
}

type InputErr struct {
	pgerr *pgconn.PgError
}

func (err InputErr) Error() string {
	return err.pgerr.Message
}

func (err InputErr) Type() string {
	return err.pgerr.TableName
}

func (err InputErr) Field() string {
	return err.pgerr.ColumnName
}

func (err InputErr) IsCheck() bool {
	return sort.SearchStrings(checkConstraints, err.pgerr.ConstraintName) < len(checkConstraints)
}

func (err InputErr) IsPrimaryKey() bool {
	return sort.SearchStrings(pkeyConstraints, err.pgerr.ConstraintName) < len(pkeyConstraints)
}

func (err InputErr) IsForeignKey() bool {
	return sort.SearchStrings(fkeyConstraints, err.pgerr.ConstraintName) < len(fkeyConstraints)
}

func Error(err error) error {
	var pgerr *pgconn.PgError

	if pgconn.Timeout(err) {
		// if this is a canceled context it's likely a client closed the request prematurely
		if errors.Is(err, context.Canceled) {
			return context.Canceled
		}
		return ErrConnection
	} else if errors.Is(err, context.DeadlineExceeded) {
		return ErrConnection
	} else if pgxscan.NotFound(err) {
		return ErrNotFound
	} else if errors.As(err, &pgerr) {
		switch {
		case pgerrcode.IsInvalidTransactionState(pgerr.Code):
			return ErrInvalidTransactionMode
		case pgerrcode.IsIntegrityConstraintViolation(pgerr.Code):
			return InputErr{pgerr}
		case pgerrcode.IsConnectionException(pgerr.Code):
			return ErrConnection
		case pgerrcode.IsOperatorIntervention(pgerr.Code):
			return ErrConnection
		case pgerrcode.IsInsufficientResources(pgerr.Code):
			return ErrConnection
		case pgerrcode.SerializationFailure == pgerr.Code:
			return ErrSerialization
		}
	}

	return err
}
