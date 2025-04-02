// Package blocksub implements an Ethereum block subscriber that works with polling and/or websockets.
package blocksub

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/log"
	"go.uber.org/atomic"
)

var ErrStopped = errors.New("already stopped")
var (
	defaultPollTimeout = 10 * time.Second
	defaultSubTimeout  = 60 * time.Second
)

type BlockSubscriber interface {
	IsRunning() bool
	Subscribe(ctx context.Context) Subscription
	Start() (err error)
	Stop()
}

type BlockSub struct {
	PollTimeout time.Duration // 10 seconds by default (8,640 requests per day)
	SubTimeout  time.Duration // 60 seconds by default, after this timeout the subscriber will reconnect
	DebugOutput bool

	ethNodeHTTPURI      string // usually port 8545
	ethNodeWebsocketURI string // usually port 8546

	subscriptions []*Subscription

	ctx     context.Context
	cancel  context.CancelFunc
	stopped atomic.Bool

	httpClient      *ethclient.Client
	wsClient        *ethclient.Client
	wsClientSub     ethereum.Subscription
	internalHeaderC chan *ethtypes.Header // internal subscription channel

	CurrentHeader      *ethtypes.Header
	CurrentBlockNumber uint64
	CurrentBlockHash   string

	latestWsHeader   *ethtypes.Header
	wsIsConnecting   atomic.Bool
	wsConnectingCond *sync.Cond
}

func NewBlockSub(ctx context.Context, ethNodeHTTPURI, ethNodeWebsocketURI string) *BlockSub {
	return NewBlockSubWithTimeout(ctx, ethNodeHTTPURI, ethNodeWebsocketURI, defaultPollTimeout, defaultSubTimeout)
}

func NewBlockSubWithTimeout(ctx context.Context, ethNodeHTTPURI, ethNodeWebsocketURI string, pollTimeout, subTimeout time.Duration) *BlockSub {
	ctx, cancel := context.WithCancel(ctx)
	sub := &BlockSub{
		PollTimeout:         pollTimeout,
		SubTimeout:          subTimeout,
		ethNodeHTTPURI:      ethNodeHTTPURI,
		ethNodeWebsocketURI: ethNodeWebsocketURI,
		ctx:                 ctx,
		cancel:              cancel,
		internalHeaderC:     make(chan *ethtypes.Header),
		wsConnectingCond:    sync.NewCond(new(sync.Mutex)),
	}
	return sub
}

func (s *BlockSub) IsRunning() bool {
	return !s.stopped.Load()
}

// Subscribe is used to create a new subscription.
func (s *BlockSub) Subscribe(ctx context.Context) Subscription {
	sub := NewSubscription(ctx)
	if s.stopped.Load() {
		sub.Unsubscribe()
	} else {
		go sub.run()
		s.subscriptions = append(s.subscriptions, &sub)
	}
	return sub
}

// Start starts polling and websocket threads.
func (s *BlockSub) Start() (err error) {
	if s.stopped.Load() {
		return ErrStopped
	}

	go s.runListener()

	if s.ethNodeWebsocketURI != "" {
		err = s.startWebsocket(false)
		if err != nil {
			return err
		}
	}

	if s.ethNodeHTTPURI != "" {
		log.Info("BlockSub:Start - HTTP connecting...", "uri", s.ethNodeHTTPURI)
		s.httpClient, err = ethclient.Dial(s.ethNodeHTTPURI)
		if err != nil { // using an invalid port will NOT return an error here, only at polling
			return err
		}

		// Ensure that polling works
		err = s._pollNow()
		if err != nil {
			return err
		}

		log.Info("BlockSub:Start - HTTP connected", "uri", s.ethNodeHTTPURI)
		go s.runPoller()
	}

	return nil
}

// Stop closes all subscriptions and stops the polling and websocket threads.
func (s *BlockSub) Stop() {
	if s.stopped.Swap(true) {
		return
	}

	for _, sub := range s.subscriptions {
		sub.Unsubscribe()
	}

	s.cancel()
}

// Listens to internal headers and forwards them to the subscriber if the header has a greater blockNumber or different hash than the previous one.
func (s *BlockSub) runListener() {
	for {
		select {
		case <-s.ctx.Done():
			s.Stop() // ensures all subscribers are properly closed
			return

		case header := <-s.internalHeaderC:
			// use the new header if it's later or has a different hash than the previous known one
			if header.Number.Uint64() >= s.CurrentBlockNumber && header.Hash().Hex() != s.CurrentBlockHash {
				s.CurrentHeader = header
				s.CurrentBlockNumber = header.Number.Uint64()
				s.CurrentBlockHash = header.Hash().Hex()

				// Send to each subscriber
				for _, sub := range s.subscriptions {
					if sub.stopped.Load() {
						continue
					}

					select {
					case sub.C <- header:
					default:
					}
				}
			}
		}
	}
}

func (s *BlockSub) runPoller() {
	ch := time.After(s.PollTimeout)
	for {
		select {
		case <-s.ctx.Done():
			return
		case <-ch:
			err := s._pollNow()
			if err != nil {
				log.Error("BlockSub: polling latest block failed", "err", err)
			}
			ch = time.After(s.PollTimeout)
		}
	}
}

func (s *BlockSub) _pollNow() error {
	header, err := s.httpClient.HeaderByNumber(s.ctx, nil)
	if err != nil {
		return err
	}

	if s.DebugOutput {
		log.Debug("BlockSub: polled block", "number", header.Number.Uint64(), "hash", header.Hash().Hex())
	}
	s.internalHeaderC <- header

	// Ensure websocket is still working (force a reconnect if it lags behind)
	if s.latestWsHeader != nil && s.latestWsHeader.Number.Uint64() < header.Number.Uint64()-2 {
		log.Warn("BlockSub: forcing websocket reconnect from polling", "wsBlockNum", s.latestWsHeader.Number.Uint64(), "pollBlockNum", header.Number.Uint64())
		go s.startWebsocket(true)
	}

	return nil
}

// startWebsocket tries to establish a websocket connection to the node. If retryForever is true it will retry forever, until it is connected.
// Also blocks if another instance is currently connecting.
func (s *BlockSub) startWebsocket(retryForever bool) error {
	if isAlreadyConnecting := s.wsIsConnecting.Swap(true); isAlreadyConnecting {
		s.wsConnectingCond.L.Lock()
		s.wsConnectingCond.Wait()
		s.wsConnectingCond.L.Unlock()
		return nil
	}

	defer func() {
		s.wsIsConnecting.Store(false)
		s.wsConnectingCond.Broadcast()
	}()

	for {
		if s.wsClient != nil {
			s.wsClient.Close()
		}

		err := s._startWebsocket()
		if err != nil && retryForever {
			log.Error("BlockSub:startWebsocket failed, retrying...", "err", err)
		} else {
			return err
		}
	}
}

func (s *BlockSub) _startWebsocket() (err error) {
	log.Info("BlockSub:_startWebsocket - connecting...", "uri", s.ethNodeWebsocketURI)

	s.wsClient, err = ethclient.Dial(s.ethNodeWebsocketURI)
	if err != nil {
		return err
	}

	wsHeaderC := make(chan *ethtypes.Header)
	s.wsClientSub, err = s.wsClient.SubscribeNewHead(s.ctx, wsHeaderC)
	if err != nil {
		return err
	}

	// Listen for headers and errors, and reconnect if needed
	go func() {
		timer := time.NewTimer(s.SubTimeout)

		for {
			select {
			case <-s.ctx.Done():
				return

			case err := <-s.wsClientSub.Err():
				if err == nil { // shutdown
					return
				}

				// reconnect
				log.Warn("BlockSub: headerSub failed, reconnect now", "err", err)
				go s.startWebsocket(true)
				return

			case <-timer.C:
				log.Warn("BlockSub: timeout, reconnect now", "timeout", s.SubTimeout)
				go s.startWebsocket(true)
				return

			case header := <-wsHeaderC:
				timer.Reset(s.SubTimeout)
				if s.DebugOutput {
					log.Debug("BlockSub: sub block", "number", header.Number.Uint64(), "hash", header.Hash().Hex())
				}
				s.latestWsHeader = header
				s.internalHeaderC <- header
			}
		}
	}()

	log.Info("BlockSub:_startWebsocket - connected", "uri", s.ethNodeWebsocketURI)
	return nil
}
