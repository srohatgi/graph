package graph

var logger = func(args ...interface{}) {}

// WithLogger allows library users to customize logging
func WithLogger(customLogger func(...interface{})) {
	logger = customLogger
}
