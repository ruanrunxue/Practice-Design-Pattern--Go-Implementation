package network

import "errors"

var (
	ErrEndpointAlreadyListened = errors.New("endpoint already listened")
	ErrConnectionRefuse        = errors.New("connection refuse")
)
