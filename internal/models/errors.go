package models

import "errors"

var (
	ErrInvalidURL       = errors.New("invalid URL")
	ErrNotFound         = errors.New("short code not found")
	ErrCollision        = errors.New("short code collision")
	ErrFailedToGenerate = errors.New("failed to generate unique short code")
)
