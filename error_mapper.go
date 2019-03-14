package graph

import "strings"

// ErrorMapper enables query into a map of errors
type ErrorMapper interface {
	ErrorMap() map[string]error
}

type errorMap map[string]error

func (es errorMap) Error() string {
	var sb strings.Builder
	for index, err := range es {
		sb.WriteString(string(index))
		sb.WriteString(":")
		sb.WriteString(err.Error())
		sb.WriteString(";")
	}
	return sb.String()
}

func (es errorMap) ErrorMap() map[string]error {
	return es
}
