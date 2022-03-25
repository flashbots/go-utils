// Example for multiple subscribers
package main

import (
	"context"
	"time"

	"github.com/ethereum/go-ethereum/log"
	"github.com/flashbots/go-utils/blocksub"
)

func DemoMultiSub(httpURI, wsURI string) {
	// Create and start a BlockSub instance
	blocksub := blocksub.NewBlockSub(context.Background(), httpURI, wsURI)
	blocksub.DebugOutput = true
	if err := blocksub.Start(); err != nil {
		log.Crit(err.Error())
	}

	// Create two regular subscriptions
	go listen(1, blocksub.Subscribe(context.Background()))
	go listen(2, blocksub.Subscribe(context.Background()))

	// Create a third subscription, which will be cancelled after 10 seconds
	ctx, cancel := context.WithCancel(context.Background())
	go listen(3, blocksub.Subscribe(ctx))
	time.Sleep(10 * time.Second)
	cancel()

	// // Wait 10 seconds and then stop the blocksub
	// time.Sleep(10 * time.Second)
	// blocksub.Stop()

	// Sleep forever
	select {}
}

func listen(id int, subscription blocksub.Subscription) {
	for {
		select {
		case <-subscription.Done():
			log.Info("sub finished", "id", id)
			return
		case header := <-subscription.C:
			log.Info("new header", "id", id, "number", header.Number.Uint64(), "hash", header.Hash().Hex())
		}
	}
}
