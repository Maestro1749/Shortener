package models

import "errors"

var (
	ErrInternalServer   = errors.New("Internal server error")
	ErrInvalidDataInput = errors.New("Invalid input data")
	ErrNotFound         = errors.New("Not found link")
)
