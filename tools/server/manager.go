/*
 * @Author: lwnmengjing
 * @Date: 2021/6/2 3:01 下午
 * @Last Modified by: lwnmengjing
 * @Last Modified time: 2021/6/2 3:01 下午
 */

package server

import (
	"context"
	"net/http"

	log "github.com/matchstalk/mss-core/logger"
	"github.com/matchstalk/mss-core/tools/metrics"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// Manager initializes shared dependencies such as Caches and Clients, and provides them to Runnable.
type Manager interface {
	// Add will set requested dependencies on the component, and cause the component to be
	// started when Start is called.
	Add(Runnable) error
	// AddMetricsHandler adds an handler served on path to the http server that serves metrics.
	AddMetricsHandler(handler http.Handler) error
	// AddHealthzHandler allows you to add Healthz handler
	AddHealthzHandler(handler http.Handler) error
	// AddReadyzHandler allows you to add Readyz handler
	AddReadyzHandler(handler http.Handler) error
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

	// Create health probes listener. This will throw an error if the bind
	// address is invalid or already in use.
	healthProbeListener, err := options.healthProbeListener(options.healthProbeBind)
	if err != nil {
		return nil, err
	}

	return &Server{
		metricsListener: metricsListener,
		metricsHandler: promhttp.HandlerFor(metrics.Registry, promhttp.HandlerOpts{
			EnableOpenMetrics: true,
			ErrorHandling:     promhttp.HTTPErrorOnError,
		}),
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
