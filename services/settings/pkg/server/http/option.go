package http

import (
	"context"

	"github.com/opencloud-eu/opencloud/pkg/log"
	"github.com/opencloud-eu/opencloud/services/settings/pkg/config"
	"github.com/opencloud-eu/opencloud/services/settings/pkg/metrics"
	"github.com/opencloud-eu/opencloud/services/settings/pkg/settings"
	"github.com/urfave/cli/v2"
	"go.opentelemetry.io/otel/trace"
)

// Option defines a single option function.
type Option func(o *Options)

// Options defines the available options for this package.
type Options struct {
	Name           string
	Logger         log.Logger
	Context        context.Context
	Config         *config.Config
	Metrics        *metrics.Metrics
	ServiceHandler settings.ServiceHandler
	Flags          []cli.Flag
	TraceProvider  trace.TracerProvider
}

// newOptions initializes the available default options.
func newOptions(opts ...Option) Options {
	opt := Options{}

	for _, o := range opts {
		o(&opt)
	}

	return opt
}

// Logger provides a function to set the logger option.
func Logger(val log.Logger) Option {
	return func(o *Options) {
		o.Logger = val
	}
}

// Name provides a name for the service.
func Name(val string) Option {
	return func(o *Options) {
		o.Name = val
	}
}

// Context provides a function to set the context option.
func Context(val context.Context) Option {
	return func(o *Options) {
		o.Context = val
	}
}

// Config provides a function to set the config option.
func Config(val *config.Config) Option {
	return func(o *Options) {
		o.Config = val
	}
}

// Metrics provides a function to set the metrics option.
func Metrics(val *metrics.Metrics) Option {
	return func(o *Options) {
		o.Metrics = val
	}
}

// Flags provides a function to set the flags option.
func Flags(val []cli.Flag) Option {
	return func(o *Options) {
		o.Flags = append(o.Flags, val...)
	}
}

// ServiceHandler provides a function to set the ServiceHandler option
func ServiceHandler(val settings.ServiceHandler) Option {
	return func(o *Options) {
		o.ServiceHandler = val
	}
}

// TraceProvider provides a function to set the TraceProvider option
func TraceProvider(val trace.TracerProvider) Option {
	return func(o *Options) {
		o.TraceProvider = val
	}
}
