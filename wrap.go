package graph

import "strings"

// ErrorMap is a collection of errors. Integer offsets in the map are assumed to be
// meaningful in the context of the function signature.
type ErrorMap map[int]error

// Get is required to work around the issue of interface having a concrete underlying
// value nil is not a nil interface.
func (es *ErrorMap) Get() error {
	if es == nil || len(*es) == 0 {
		return nil
	}
	return es
}

// Error satisfies the error interface.
func (es *ErrorMap) Error() string {
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
func (es *ErrorMap) Size() int {
	if es == nil {
		return 0
	}
	return len(*es)
}
