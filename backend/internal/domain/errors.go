package domain

import (
	"errors"
	"fmt"
)

var ErrNotFound = errors.New("not found")

type ValidationError struct {
	Message string
}

func (e ValidationError) Error() string {
	return fmt.Sprintf("validation error: %s", e.Message)
}
