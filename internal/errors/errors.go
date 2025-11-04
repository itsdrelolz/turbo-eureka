package errors

import (
	"errors"
)

// indicates an unrecoverable error
var ErrPermanentFailure = errors.New("permanent failure, do not retry")
