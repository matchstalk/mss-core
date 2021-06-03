/*
 * @Author: lwnmengjing
 * @Date: 2021/6/2 3:05 下午
 * @Last Modified by: lwnmengjing
 * @Last Modified time: 2021/6/2 3:05 下午
 */

package server

import (
	"net"
	"time"
)

const (
	defaultLeaseDuration          = 15 * time.Second
	defaultRenewDeadline          = 15 * time.Second
	defaultRetryPeriod            = 2 * time.Second
	defaultGracefulShutdownPeriod = 30 * time.Second

	defaultReadinessEndpoint = "/readyz"
	defaultLivenessEndpoint  = "/healthz"
	defaultMetricsEndpoint   = "/metrics"
)

type Options struct {
	// metricsBind is the TCP address that the controller should bind to
	// for serving prometheus metrics.
	// It can be set to "0" to disable the metrics serving.
	metricsBind string

	// healthProbeBind is the TCP address that the controller should bind to
	// for serving health probes
	healthProbeBind string

	// readinessEndpoint probe endpoint name, defaults to "readyz"
	readinessEndpoint string

	// livenessEndpoint Liveness probe endpoint name, defaults to "healthz"
	livenessEndpoint string

	// gracefulShutdownTimeout is the duration given to runnable to stop before the manager actually returns on stop.
	// To disable graceful shutdown, set to time.Duration(0)
	// To use graceful shutdown without timeout, set to a negative duration, e.G. time.Duration(-1)
	// The graceful shutdown is skipped for safety reasons in case the leader election lease is lost.
	gracefulShutdownTimeout *time.Duration

	metricsListener     func(string) (net.Listener, error)
	healthProbeListener func(string) (net.Listener, error)
}

func defaultOptions() *Options {
	gracefulShutdownTimeout := defaultGracefulShutdownPeriod
	return &Options{
		metricsListener:         NewListener,
		metricsBind:             "0",
		readinessEndpoint:       defaultReadinessEndpoint,
		livenessEndpoint:        defaultLivenessEndpoint,
		healthProbeListener:     NewListener,
		healthProbeBind:         "0",
		gracefulShutdownTimeout: &gracefulShutdownTimeout,
	}

}

type Option func(*Options)

func WithMetricsBindOption(s string) Option {
	return func(o *Options) {
		o.metricsBind = s
	}
}

func WithHealthProbeBindOption(s string) Option {
	return func(o *Options) {
		o.healthProbeBind = s
	}
}

func WithReadinessEndpointOption(s string) Option {
	return func(o *Options) {
		o.readinessEndpoint = s
	}
}

func WithLivenessEndpointOption(s string) Option {
	return func(o *Options) {
		o.livenessEndpoint = s
	}
}

func WithGracefulShutdownTimeoutOption(t time.Duration) Option {
	return func(o *Options) {
		o.gracefulShutdownTimeout = &t
	}
}

func WithMetricsListenerOption(f func(string) (net.Listener, error)) Option {
	return func(o *Options) {
		o.metricsListener = f
	}
}

func WithHealthProbeListenerOption(f func(string) (net.Listener, error)) Option {
	return func(o *Options) {
		o.healthProbeListener = f
	}
}
