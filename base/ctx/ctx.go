package ctx

import (
	"context"

	"github.com/sirupsen/logrus"
)

// CTX extends Golang's context to support logging methods
type CTX struct {
	context.Context
	logrus.FieldLogger
}

// Background returns an empty context
func Background() CTX {
	return CTX{
		Context:     context.Background(),
		FieldLogger: logrus.StandardLogger(),
	}
}
