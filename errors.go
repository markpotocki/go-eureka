package eureka

import "errors"

var ErrInstanceIDNotFound = errors.New("instance ID not found")
var ErrStatusUpdateFailed = errors.New("status update failed")
