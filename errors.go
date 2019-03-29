package graph

import "fmt"

type ValidationError struct {
	errorStr string
}

func (v ValidationError) Error() string {
	return fmt.Sprintf("validation error: %s", v.errorStr)
}
