package customerrors

import "errors"

var (
	//ErrTooManyRequests is returned when a worker controller is currently paused
	ErrTooManyRequests error = errors.New("the worker controller is currently paused")
	ErrNilRedisValue error  = errors.New("this key doesn't exist in the redis cache")
)