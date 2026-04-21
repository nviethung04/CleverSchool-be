package websocket

import "errors"

// WebSocket errors
var (
	ErrClientBufferFull = errors.New("client send buffer is full")
	ErrInvalidToken     = errors.New("invalid or missing token")
	ErrInvalidCourseID  = errors.New("invalid or missing course_id")
	ErrUserNotFound     = errors.New("user not found")
	ErrUnauthorized     = errors.New("unauthorized")
	ErrMessageTooLong   = errors.New("message content too long")
	ErrEmptyMessage     = errors.New("message content is empty")
)

