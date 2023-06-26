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
	ErrRefreshTokenExpired    = fmt.Errorf("refresh_token_expired")
	ErrAccessTokenExpired     = fmt.Errorf("access_token_expired")
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

// CheckTokenError is a struct that is used to handle token related errors
type CheckTokenError struct{}

// Expired is a funciton that is used to identify wether the token is expired or not
func (CheckTokenError) Expired(err error) bool {
	if err.Error() == "token has invalid claims: token is expired" {
		return true
	}

	return false
}
