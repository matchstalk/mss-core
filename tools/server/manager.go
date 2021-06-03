/*
 * @Author: lwnmengjing
 * @Date: 2021/6/2 3:01 下午
 * @Last Modified by: lwnmengjing
 * @Last Modified time: 2021/6/2 3:01 下午
 */

package server

import (
	"context"
	log "github.com/matchstalk/mss-core/logger"
	"net/http"
)

// Manager initializes shared dependencies such as Caches and Clients, and provides them to Runnable.
type Manager interface {
	// Add will set requested dependencies on the component, and cause the component to be
	// started when Start is called.
	Add(Runnable) error
	// AddMetricsExtraHandler adds an handler served on path to the http server that serves metrics.
	AddMetricsExtraHandler(path string, handler http.Handler) error
	// AddHealthzHandler allows you to add Healthz handler
	AddHealthzHandler(name string, handler http.Handler) error
	// AddReadyzHandler allows you to add Readyz handler
	AddReadyzHandler(name string, handler http.Handler) error
	// Start starts all registered Controllers and blocks until the context is canceled.
	Start(ctx context.Context) error
}

type Runnable interface {
	Start(ctx context.Context) error
}

type RunnableFunc func(ctx context.Context) error

func (r RunnableFunc) Start(c context.Context) error {
	return r(c)
}

// New returns a new Manager for creating Controllers.
func New(opts ...Option) (Manager, error) {
	// Set default values for options fields
	options := defaultOptions()
	for _, o := range opts {
		o(options)
	}

	// Create the metrics listener. This will throw an error if the metrics bind
	// address is invalid or already in use.
	metricsListener, err := options.metricsListener(options.metricsBind)
	if err != nil {
		return nil, err
	}

	// By default we have no extra endpoints to expose on metrics http server.
	metricsExtraHandlers := make(map[string]http.Handler)

	// Create health probes listener. This will throw an error if the bind
	// address is invalid or already in use.
	healthProbeListener, err := options.healthProbeListener(options.healthProbeBind)
	if err != nil {
		return nil, err
	}

	return &Server{
		metricsListener:         metricsListener,
		metricsExtraHandlers:    metricsExtraHandlers,
		logger:                  log.NewHelper(log.DefaultLogger),
		elected:                 make(chan struct{}),
		leaseDuration:           defaultLeaseDuration,
		renewDeadline:           defaultRenewDeadline,
		retryPeriod:             defaultRetryPeriod,
		healthProbeListener:     healthProbeListener,
		readinessEndpointName:   options.readinessEndpoint,
		livenessEndpointName:    options.livenessEndpoint,
		gracefulShutdownTimeout: *options.gracefulShutdownTimeout,
		internalProceduresStop:  make(chan struct{}),
	}, nil
}
