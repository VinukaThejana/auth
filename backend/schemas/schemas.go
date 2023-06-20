// Package schemas is used to defiend custom schemas that are used throughout the application
package schemas

import "github.com/VinukaThejana/go-utils/logger"

var log logger.Logger

// Response is a the default response that is being sent to the client
type Response struct {
	Status  string `json:"status"`
	Message string `json:"message"`
}
