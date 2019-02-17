package graph

import "strings"

// ErrorSlice returns a slice of errors.
type ErrorSlice []error

// Get works around the issue of interface having a concrete nil value is not nil
func (es *ErrorSlice) Get() error {
	if es == nil || len(*es) == 0 {
		return nil
	}
	return es
}

func (es *ErrorSlice) Error() string {
	var sb strings.Builder
	for _, e := range *es {
		sb.WriteString(e.Error())
	}
	return sb.String()
}

// Size returns the numbers of errors contained in the slice.
func (es *ErrorSlice) Size() int {
	if es == nil {
		return 0
	}
	return len(*es)
}
