package graph

var logger = func(args ...interface{}) {}

// WithLogger allows library users to customize logging. By
// default, the library does not produce any logs.
func WithLogger(customLogger func(...interface{})) {
	logger = customLogger
}
