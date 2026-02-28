package models

import "errors"

var (
	ErrCollision        = errors.New("short code collision")
	ErrFailedToGenerate = errors.New("failed to generate unique short code")
	ErrRecordNotFound   = errors.New("record not found")
)
