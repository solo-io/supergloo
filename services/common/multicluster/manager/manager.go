package mc_manager

import (
	"context"
	"time"

	"github.com/rotisserie/eris"
	"github.com/solo-io/go-utils/contextutils"
	"github.com/solo-io/mesh-projects/services/internal/errutils"
	"go.uber.org/zap"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/manager"
)

var (
	ManagerCacheSyncError = eris.New("unable to sync async manager cache")
	ManagerStartError     = func(err error) error {
		return eris.Wrapf(err, "unable to start async manager")
	}
	ManagerStartOptionsFuncError = func(err error) error {
		return eris.Wrapf(err, "error evaluating start options function for async manager")
	}
)

//go:generate mockgen -source manager.go -destination ./mocks/async_manager.go

/*
	This interface is meant to represent an asynchronous wrapper on top of a controller-runtime manager.
	It comes with opinionated ways to start and stop managers in go routines, as well as check their error
	status
*/
type AsyncManager interface {
	// returns the manager associated with the async manager
	Manager() manager.Manager
	// returns the context of the async manager
	Context() context.Context
	// returns the err which has occured
	Error() error
	// returns the channel which is closed when an err occurs
	GotError() <-chan struct{}
	// start the async manager, does not block, will signal the `GotError()` channel if an error occurs
	Start(opts ...AsyncManagerStartOptionsFunc) error
	// stops the async manager, does not block
	Stop()
}

type AsyncManagerFactory interface {
	New(parentCtx context.Context, cfg *rest.Config,
		opts AsyncManagerOptions) (AsyncManager, error)
}

func NewAsyncManagerFactory() AsyncManagerFactory {
	return &asyncManagerFactory{}
}

type asyncManagerFactory struct{}

func (a *asyncManagerFactory) New(parentCtx context.Context, cfg *rest.Config,
	opts AsyncManagerOptions) (AsyncManager, error) {

	mgr, err := manager.New(cfg, opts.Convert())
	if err != nil {
		return nil, err
	}
	ctx, cancel := context.WithCancel(parentCtx)
	return &asyncManager{
		mgr:      mgr,
		ctx:      ctx,
		cancel:   cancel,
		signaler: errutils.NewErrorSignaler(),
	}, nil
}

type asyncManager struct {
	mgr      manager.Manager
	ctx      context.Context
	cancel   context.CancelFunc
	signaler errutils.ErrorSignaler
}

type AsyncManagerOptions struct {
	Namespace              string
	MetricsBindAddress     string
	HealthProbeBindAddress string
}

/*
	This function defines some sane defaults for the manager.
	A port set to 0 in the controller-runtime disables the service. So by default this will
	disable the health check server, as well as the metrics server.

	This is useful when starting managers asynchronously as they will by default all try to bind to port :8081
	for metrics which will fail when multiple are started.
*/
func (c AsyncManagerOptions) Convert() manager.Options {
	managerOpts := manager.Options{}
	if c.HealthProbeBindAddress == "" {
		managerOpts.HealthProbeBindAddress = "0"
	}
	if c.MetricsBindAddress == "" {
		managerOpts.MetricsBindAddress = "0"
	}
	return managerOpts
}

func NewAsyncManager(parentCtx context.Context, mgr manager.Manager) AsyncManager {
	ctx, cancel := context.WithCancel(parentCtx)
	return &asyncManager{
		mgr:      mgr,
		ctx:      ctx,
		cancel:   cancel,
		signaler: errutils.NewErrorSignaler(),
	}
}

func (c *asyncManager) Manager() manager.Manager {
	return c.mgr
}

func (c *asyncManager) Context() context.Context {
	return c.ctx
}

func (c *asyncManager) Error() error {
	return c.signaler.Error()
}

func (c *asyncManager) GotError() <-chan struct{} {
	return c.signaler.GotError()
}

/*
	This fn type is meant to be used as a way of recieving the context and manager of an async manager upon start up.
	This is meant to consolidate the surface area where controllers and other context/manager dependent functions are
	run.
*/
type AsyncManagerStartOptionsFunc func(ctx context.Context, mgr manager.Manager) error

func (c *asyncManager) Start(opts ...AsyncManagerStartOptionsFunc) error {
	go func() {
		if err := c.Manager().Start(c.Context().Done()); err != nil {
			contextutils.LoggerFrom(c.Context()).Errorw("error starting async manager", zap.Error(err))
			c.signaler.SignalError(ManagerStartError(err))
		}
	}()
	// timeout cache sync after 2 seconds if it fails
	withDeadline, _ := context.WithTimeout(c.Context(), 2*time.Second)
	if synced := c.Manager().GetCache().WaitForCacheSync(withDeadline.Done()); !synced {
		c.Stop()
		return ManagerCacheSyncError
	}
	for _, v := range opts {
		if err := v(c.Context(), c.Manager()); err != nil {
			c.Stop()
			return ManagerStartOptionsFuncError(err)
		}
	}
	return nil
}

func (c *asyncManager) Stop() {
	c.cancel()
}
