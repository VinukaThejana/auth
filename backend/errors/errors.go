// Package errors is used to handle errors
package errors

import "fmt"

var (
	//revive:disable
	ErrInternalServerError    = fmt.Errorf("internal_server_error")
	ErrUnauthorized           = fmt.Errorf("unauthorized")
	ErrAccessTokenNotProvided = fmt.Errorf("access_token_not_provided")
	ErrBadRequest             = fmt.Errorf("bad_request")
	Okay                      = "okay"

//revive:enable
)
