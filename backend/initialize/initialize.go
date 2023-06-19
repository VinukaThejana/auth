// Package initialize is used to initialize connections to third party services
package initialize

import "github.com/VinukaThejana/go-utils/logger"

var log logger.Logger

// H contains all the connections to third party services
type H struct {
	DB *DB
}
