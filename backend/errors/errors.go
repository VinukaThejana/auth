// Package errors is used to handle errors
package errors

import (
	errs "errors"
	"fmt"

	"github.com/jackc/pgx/v5/pgconn"
)

var (
	//revive:disable
	ErrInternalServerError    = fmt.Errorf("internal_server_error")
	ErrUnauthorized           = fmt.Errorf("unauthorized")
	ErrAccessTokenNotProvided = fmt.Errorf("access_token_not_provided")
	ErrBadRequest             = fmt.Errorf("bad_request")
	ErrIncorrectCredentials   = fmt.Errorf("incorrect_credentials")
	Okay                      = "okay"

//revive:enable
)

// CheckDBError is a struc that is used to identify the database errors
type CheckDBError struct{}

// DuplicateKey is a function that is used to find wether the the returned postgres error
// is due to a duplicate key entry (A unique key constraint)
func (CheckDBError) DuplicateKey(err error) bool {
	var pgErr *pgconn.PgError
	if errs.As(err, &pgErr) {
		if pgErr.Code == "23505" {
			return true
		}
	}

	return false
}
