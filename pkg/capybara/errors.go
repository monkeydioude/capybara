package capybara

import "errors"

// common
var ErrNilPointer = errors.New("ponter was nil")

// service
var ErrServiceHandleHTTP = errors.New("service could not handle HTTP request")
