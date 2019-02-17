package graph

import "strings"

// ErrorSlice returns a slice of errors.
type ErrorSlice map[int]error

// Get works around the issue of interface having a concrete nil value is not nil
func (es *ErrorSlice) Get() error {
	if es == nil || len(*es) == 0 {
		return nil
	}
	return es
}

func (es *ErrorSlice) Error() string {
	var sb strings.Builder
	for index, err := range *es {
		sb.WriteString(string(index))
		sb.WriteString(":")
		sb.WriteString(err.Error())
		sb.WriteString(";")
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
