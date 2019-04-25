package registration

import (
	"context"
	"sync"
	"time"

	"github.com/solo-io/go-utils/contextutils"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients"
	"github.com/solo-io/solo-kit/pkg/api/v1/eventloop"
	"go.uber.org/zap"
)

type Receiver <-chan EnabledConfigLoops

type PubSub struct {
	subscriberCache []chan EnabledConfigLoops
	subscriberLock  sync.RWMutex
}

func NewPubsub() *PubSub {
	return &PubSub{}
}

func (r *PubSub) subscribe() Receiver {
	r.subscriberLock.Lock()
	defer r.subscriberLock.Unlock()
	c := make(chan EnabledConfigLoops, 10)
	r.subscriberCache = append(r.subscriberCache, c)
	return c
}

func (r *PubSub) unsubscribe(c Receiver) {
	r.subscriberLock.Lock()
	defer r.subscriberLock.Unlock()
	for i, subscriber := range r.subscriberCache {
		if subscriber == c {
			r.subscriberCache = append(r.subscriberCache[:i], r.subscriberCache[i+1:]...)
			return
		}
	}
}

func (r *PubSub) publish(ctx context.Context, config EnabledConfigLoops) {
	r.subscriberLock.RLock()
	defer r.subscriberLock.RUnlock()
	for _, subscriber := range r.subscriberCache {
		select {
		case <-ctx.Done():
			return
		case subscriber <- config:
		default:

		}
	}
}

type Subscriber struct {
	enabledChannel Receiver
	configLoop     ConfigLoop
}

func NewSubscriber(ctx context.Context, pubsub *PubSub, cl ConfigLoop) *Subscriber {
	ch := pubsub.subscribe()
	go func() {
		<-ctx.Done()
		pubsub.unsubscribe(ch)
	}()
	return &Subscriber{enabledChannel: ch, configLoop: cl}
}

func (l *Subscriber) Listen(parentCtx context.Context) {
	go func() {
		previousState := EnabledConfigLoops{}
		childCtx, cancel := context.WithCancel(parentCtx)
		defer cancel()
		logger := contextutils.LoggerFrom(parentCtx)

		for {
			select {
			case nextState := <-l.enabledChannel:

				previouslyEnabled := l.configLoop.Enabled(previousState)
				currentlyEnabled := l.configLoop.Enabled(nextState)

				switch {
				case previouslyEnabled && !currentlyEnabled:
					// disabled
					cancel()
				case !previouslyEnabled && currentlyEnabled:
					// enabled
					childCtx, cancel = context.WithCancel(parentCtx)
					err := RunConfigLoop(childCtx, nextState, l.configLoop.Start)
					if err != nil {
						logger.Errorw("could not start config loop", zap.Error(err))
						return
					}
				}
				previousState = nextState

			case <-parentCtx.Done():
				return
			}
		}
	}()
}

func RunConfigLoop(ctx context.Context, enabledFeatures EnabledConfigLoops, starter ConfigLoopStarter) error {
	watchOpts := clients.WatchOpts{
		Ctx:         ctx,
		RefreshRate: time.Minute * 1,
	}

	loop, err := starter(ctx, enabledFeatures)
	if err != nil {
		return err
	}

	if loop == nil {
		return nil
	}

	return RunEventLoop(ctx, loop, watchOpts)

}

func RunEventLoop(ctx context.Context, loop eventloop.EventLoop, opts clients.WatchOpts) error {
	logger := contextutils.LoggerFrom(ctx)
	combinedErrs, err := loop.Run(nil, opts)
	if err != nil {
		return err
	}
	go func() {
		for {
			select {
			case err := <-combinedErrs:
				if err != nil {
					logger.With(zap.Error(err)).Info("config event loop failure")
				}
			case <-ctx.Done():
			}
		}
	}()
	return nil
}
