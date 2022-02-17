package blocksub

import (
	"context"

	ethtypes "github.com/ethereum/go-ethereum/core/types"
	"go.uber.org/atomic"
)

// Subscription will push new headers to a subscriber until the context is done or Unsubscribe() is called,
// at which point the subscription is stopped and the header channel closed.
type Subscription struct {
	C chan *ethtypes.Header // Channel to receive the headers on.

	ctx    context.Context
	cancel context.CancelFunc

	stopped atomic.Bool
}

func NewSubscription(ctx context.Context) Subscription {
	ctxWithCancel, cancel := context.WithCancel(ctx)
	return Subscription{
		C:      make(chan *ethtypes.Header),
		ctx:    ctxWithCancel,
		cancel: cancel,
	}
}

func (sub *Subscription) run() {
	<-sub.ctx.Done()
	sub.Unsubscribe()
}

// Unsubscribe unsubscribes the notification and closes the header channel.
// It can safely be called more than once.
func (sub *Subscription) Unsubscribe() {
	if sub.stopped.Swap(true) {
		return
	}
	sub.cancel()
	close(sub.C)
}

func (sub *Subscription) Done() <-chan struct{} {
	return sub.ctx.Done()
}
