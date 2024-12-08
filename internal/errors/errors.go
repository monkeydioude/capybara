package errors

import "errors"

// common
var ErrNilPointer = errors.New("pointer was nil")

// service
var ErrServiceHandleHTTP = errors.New("service could not handle HTTP request")

// groc
var ErrUnableToConnectToBackend = errors.New("unable to connect to backend")
var ErrUnableToCreateClientStream = errors.New("unable to create client stream")
