package handler

import (
	"errors"
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"
	"github.com/lib/pq"
)

// Postgres SQLSTATE codes. Only invalid_text_representation is mapped onto a
// client-facing status: a malformed identifier reaching a uuid/inet column is
// unambiguously the caller's mistake. Every other code deliberately stays a
// 500 — mislabelling one of our own bugs as the caller's fault is worse than a
// vague 500. Extending the map later is a one-line change. (#298)
const sqlStateInvalidTextRepresentation = "22P02"

// ErrMsgInvalidIdentifier is returned when a path/body identifier could not be
// parsed by Postgres (e.g. "not-a-uuid" reaching a uuid column).
const ErrMsgInvalidIdentifier = "Invalid identifier format"

// isDriverError reports whether a Postgres driver error sits anywhere in the
// error chain. lib/pq renders "pq: <message> (<SQLSTATE>)" and, when the driver
// kept the statement, appends " at position <line>:<col>" — so the rendered
// string hands any authenticated caller the SQLSTATE, the column type and the
// geometry of our SQL. errors.As is the primary test (repositories wrap with
// %w); the strings.Contains fallback catches a chain that was flattened with
// %v somewhere and can no longer be unwrapped.
func isDriverError(err error) bool {
	if err == nil {
		return false
	}
	var pqErr *pq.Error
	if errors.As(err, &pqErr) {
		return true
	}
	return strings.Contains(err.Error(), "pq: ")
}

// SafeErrorDetail is what goes into ErrorResponse.Details. It returns "" when a
// driver error is present so that `json:"details,omitempty"` drops the field
// entirely. The caller has already logged the full error; this only decides
// what crosses the wire. Application errors (nginx -t output, validation,
// ACME failures) are returned verbatim.
func SafeErrorDetail(err error) string {
	if err == nil || isDriverError(err) {
		return ""
	}
	return err.Error()
}

// SafeErrorMessage is what goes into a bare {"error": ...} field, where an
// empty string would leave the client with nothing to show.
func SafeErrorMessage(err error) string {
	if err == nil {
		return ErrMsgInternalError
	}
	if isDriverError(err) {
		return ErrMsgDatabaseError
	}
	return err.Error()
}

// ClientStatusForDBError maps a driver error onto a caller-facing status when
// the SQLSTATE proves the request itself was malformed. ok=false means "leave
// it a 500". Only errors.As is consulted here — never the string fallback —
// so a status is never changed on the strength of a substring match.
func ClientStatusForDBError(err error) (int, string, bool) {
	var pqErr *pq.Error
	if errors.As(err, &pqErr) && string(pqErr.Code) == sqlStateInvalidTextRepresentation {
		return http.StatusBadRequest, ErrMsgInvalidIdentifier, true
	}
	return 0, "", false
}

// directInternalError is the gate for the handlers that build their own 500
// body instead of calling internalError/databaseError. It keeps their existing
// bare {"error": ...} shape — so no client contract changes — while adding the
// same two properties the shared helpers now have: a malformed identifier is
// answered with 400, and driver text never reaches the body.
func directInternalError(c echo.Context, err error) error {
	if status, msg, ok := ClientStatusForDBError(err); ok {
		return c.JSON(status, map[string]string{"error": msg})
	}
	return c.JSON(http.StatusInternalServerError, map[string]string{"error": SafeErrorMessage(err)})
}
