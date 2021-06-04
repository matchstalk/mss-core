/*
 * @Author: lwnmengjing
 * @Date: 2021/6/2 6:11 下午
 * @Last Modified by: lwnmengjing
 * @Last Modified time: 2021/6/2 6:11 下午
 */

package server

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"net"
	"net/http"
	"sync"
	"time"

	log "github.com/matchstalk/mss-core/logger"
	"github.com/matchstalk/mss-core/tools/metrics"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var _ Runnable = &Server{}

type Server struct {
	// metricsListener is used to serve prometheus metrics
	metricsListener net.Listener

	// metricsHandler contains extra handlers to register on http server that serves metrics.
	metricsHandler http.Handler

	// healthProbeListener is used to serve liveness probe
	healthProbeListener net.Listener

	// metricsEndpointName metrics endpoint name
	metricsEndpointName string

	// Readiness probe endpoint name
	readinessEndpointName string

	// Liveness probe endpoint name
	livenessEndpointName string

	// Readyz probe handler
	readyzHandler http.Handler

	// Healthz probe handler
	healthzHandler http.Handler

	mu             sync.Mutex
	started        bool
	healthzStarted bool
	metricsStarted bool
	errChan        chan error

	// Logger is the logger that should be used by this manager.
	// If none is set, it defaults to log.Log global logger.
	logger *log.Helper

	// stop procedure engaged. In other words, we should not add anything else to the manager
	stopProcedureEngaged bool

	// elected is closed when this manager becomes the leader of a group of
	// managers, either because it won a leader election or because no leader
	// election was configured.
	elected chan struct{}

	startCache func(ctx context.Context) error

	// port is the port that the webhook server serves at.
	port int
	// host is the hostname that the webhook server binds to.
	host string
	// CertDir is the directory that contains the server key and certificate.
	// if not set, webhook server would look up the server key and certificate in
	// {TempDir}/k8s-webhook-server/serving-certs
	tls *tls.Config

	// leaseDuration is the duration that non-leader candidates will
	// wait to force acquire leadership.
	leaseDuration time.Duration
	// renewDeadline is the duration that the acting controlplane will retry
	// refreshing leadership before giving up.
	renewDeadline time.Duration
	// retryPeriod is the duration the LeaderElector clients should wait
	// between tries of actions.
	retryPeriod time.Duration

	// waitForRunnable is holding the number of runnables currently running so that
	// we can wait for them to exit before quitting the manager
	waitForRunnable sync.WaitGroup

	// gracefulShutdownTimeout is the duration given to runnable to stop
	// before the manager actually returns on stop.
	gracefulShutdownTimeout time.Duration

	// onStoppedLeading is callled when the leader election lease is lost.
	// It can be overridden for tests.
	onStoppedLeading func()

	// shutdownCtx is the context that can be used during shutdown. It will be cancelled
	// after the gracefulShutdownTimeout ended. It must not be accessed before internalStop
	// is closed because it will be nil.
	shutdownCtx context.Context

	internalCtx    context.Context
	internalCancel context.CancelFunc

	// internalProceduresStop channel is used internally to the manager when coordinating
	// the proper shutdown of servers. This channel is also used for dependency injection.
	internalProceduresStop chan struct{}
}

// Add sets dependencies on i, and adds it to the list of Runnables to start.
func (cm *Server) Add(r Runnable) error {
	cm.mu.Lock()
	defer cm.mu.Unlock()
	if cm.stopProcedureEngaged {
		return errors.New("can't accept new runnable as stop procedure is already engaged")
	}

	// If already started, start the controller
	cm.startRunnable(r)

	return nil
}

// AddMetricsHandler adds extra handler served on path to the http server that serves metrics.
func (cm *Server) AddMetricsHandler(handler http.Handler) error {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	if cm.stopProcedureEngaged {
		return errors.New("can't accept new metrics as stop procedure is already engaged")
	}

	if cm.metricsStarted {
		return fmt.Errorf("unable to add new checker because metrics endpoint has already been created")
	}

	cm.metricsHandler = handler
	return nil
}

// AddHealthzHandler allows you to add Healthz checker
func (cm *Server) AddHealthzHandler(check http.Handler) error {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	if cm.stopProcedureEngaged {
		return errors.New("can't accept new healthCheck as stop procedure is already engaged")
	}

	if cm.healthzStarted {
		return fmt.Errorf("unable to add new checker because healthz endpoint has already been created")
	}

	if cm.healthzHandler == nil {
		cm.healthzHandler = check
	}

	return nil
}

// AddReadyzHandler allows you to add Readyz checker
func (cm *Server) AddReadyzHandler(check http.Handler) error {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	if cm.stopProcedureEngaged {
		return errors.New("can't accept new ready check as stop procedure is already engaged")
	}

	if cm.healthzStarted {
		return fmt.Errorf("unable to add new checker because readyz endpoint has already been created")
	}

	if cm.readyzHandler == nil {
		cm.readyzHandler = check
	}
	return nil
}

func (cm *Server) GetLogger() log.Logger {
	return cm.logger
}

func (cm *Server) serveMetrics() {
	handler := promhttp.HandlerFor(metrics.Registry, promhttp.HandlerOpts{
		ErrorHandling: promhttp.HTTPErrorOnError,
	})
	if cm.metricsHandler != nil {
		handler = cm.metricsHandler
	}
	mux := http.NewServeMux()
	mux.Handle(defaultMetricsEndpoint, handler)

	cm.mu.Lock()
	defer cm.mu.Unlock()
	cm.metricsStarted = true

	server := http.Server{
		Handler: mux,
	}
	// Run the server
	cm.startRunnable(RunnableFunc(func(_ context.Context) error {
		cm.logger.Info("starting metrics server", "path", defaultMetricsEndpoint)
		if err := server.Serve(cm.metricsListener); err != nil && err != http.ErrServerClosed {
			return err
		}
		return nil
	}))

	// Shutdown the server when stop is closed
	<-cm.internalProceduresStop
	if err := server.Shutdown(cm.shutdownCtx); err != nil {
		cm.errChan <- err
	}
}

func (cm *Server) serveHealthProbes() {
	// it's done in serveMetrics.
	cm.mu.Lock()
	mux := http.NewServeMux()

	if cm.readyzHandler != nil {
		mux.Handle(cm.readinessEndpointName, http.StripPrefix(cm.readinessEndpointName, cm.readyzHandler))
		// Append '/' suffix to handle subpaths
		mux.Handle(cm.readinessEndpointName+"/", http.StripPrefix(cm.readinessEndpointName, cm.readyzHandler))
	}
	if cm.healthzHandler != nil {
		mux.Handle(cm.livenessEndpointName, http.StripPrefix(cm.livenessEndpointName, cm.healthzHandler))
		// Append '/' suffix to handle subpaths
		mux.Handle(cm.livenessEndpointName+"/", http.StripPrefix(cm.livenessEndpointName, cm.healthzHandler))
	}

	server := http.Server{
		Handler: mux,
	}
	// Run server
	cm.startRunnable(RunnableFunc(func(_ context.Context) error {
		if err := server.Serve(cm.healthProbeListener); err != nil && err != http.ErrServerClosed {
			return err
		}
		return nil
	}))
	cm.healthzStarted = true
	cm.mu.Unlock()

	// Shutdown the server when stop is closed
	<-cm.internalProceduresStop
	if err := server.Shutdown(cm.shutdownCtx); err != nil {
		cm.errChan <- err
	}
}

func (cm *Server) Start(ctx context.Context) (err error) {
	cm.internalCtx, cm.internalCancel = context.WithCancel(ctx)

	// This chan indicates that stop is complete, in other words all runnables have returned or timeout on stop request
	stopComplete := make(chan struct{})
	defer close(stopComplete)
	// This must be deferred after closing stopComplete, otherwise we deadlock.
	defer func() {
		// https://hips.hearstapps.com/hmg-prod.s3.amazonaws.com/images/gettyimages-459889618-1533579787.jpg
		stopErr := cm.engageStopProcedure(stopComplete)
		if stopErr != nil {
			if err != nil {
				// Utilerrors.Aggregate allows to use errors.Is for all contained errors
				// whereas fmt.Errorf allows wrapping at most one error which means the
				// other one can not be found anymore.
				err = fmt.Errorf("%w, stopError: %s", err, stopErr.Error())
			} else {
				err = stopErr
			}
		}
	}()

	// initialize this here so that we reset the signal channel state on every start
	// Everything that might write into this channel must be started in a new goroutine,
	// because otherwise we might block this routine trying to write into the full channel
	// and will not be able to enter the deferred cm.engageStopProcedure() which drains
	// it.
	cm.errChan = make(chan error)

	// Metrics should be served whether the controller is leader or not.
	// (If we don't serve metrics for non-leaders, prometheus will still scrape
	// the pod but will get a connection refused)
	if cm.metricsListener != nil {
		go cm.serveMetrics()
	}

	// Serve health probes
	if cm.healthProbeListener != nil {
		go cm.serveHealthProbes()
	}

	select {
	case <-ctx.Done():
		// We are done
		return nil
	case err := <-cm.errChan:
		// Error starting or running a runnable
		return err
	}
}

// engageStopProcedure signals all runnables to stop, reads potential errors
// from the errChan and waits for them to end. It must not be called more than once.
func (cm *Server) engageStopProcedure(stopComplete <-chan struct{}) error {
	// Populate the shutdown context.
	var shutdownCancel context.CancelFunc
	if cm.gracefulShutdownTimeout > 0 {
		cm.shutdownCtx, shutdownCancel = context.WithTimeout(context.Background(), cm.gracefulShutdownTimeout)
	} else {
		cm.shutdownCtx, shutdownCancel = context.WithCancel(context.Background())
	}
	defer shutdownCancel()

	// Cancel the internal stop channel and wait for the procedures to stop and complete.
	close(cm.internalProceduresStop)
	cm.internalCancel()

	// Start draining the errors before acquiring the lock to make sure we don't deadlock
	// if something that has the lock is blocked on trying to write into the unbuffered
	// channel after something else already wrote into it.
	go func() {
		for {
			select {
			case err, ok := <-cm.errChan:
				if ok {
					cm.logger.Error(err, "error received after stop sequence was engaged")
				}
			case <-stopComplete:
				return
			}
		}
	}()
	if cm.gracefulShutdownTimeout == 0 {
		return nil
	}
	cm.mu.Lock()
	defer cm.mu.Unlock()
	cm.stopProcedureEngaged = true

	// we want to close this after the other runnables stop, because we don't
	// want things like leader election to try and emit events on a closed
	// channel
	return cm.waitForRunnableToEnd(shutdownCancel)
}

// waitForRunnableToEnd blocks until all runnables ended or the
// tearDownTimeout was reached. In the latter case, an error is returned.
func (cm *Server) waitForRunnableToEnd(shutdownCancel context.CancelFunc) error {

	go func() {
		cm.waitForRunnable.Wait()
		shutdownCancel()
	}()

	<-cm.shutdownCtx.Done()
	if err := cm.shutdownCtx.Err(); err != nil && err != context.Canceled {
		return fmt.Errorf("failed waiting for all runnables to end within grace period of %s: %w", cm.gracefulShutdownTimeout, err)
	}
	return nil
}

func (cm *Server) startRunnable(r Runnable) {
	cm.waitForRunnable.Add(1)
	go func() {
		defer cm.waitForRunnable.Done()
		fmt.Print()
		if err := r.Start(cm.internalCtx); err != nil {
			cm.errChan <- err
		}
	}()
}
