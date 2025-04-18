package module

import "go.opentelemetry.io/otel"

var (
	tracer = otel.Tracer("github.com/datachainlab/ibc-parlia-relay/module")
)
