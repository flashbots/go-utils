// Package blocksub implements an Ethereum block subscriber that works with either a websocket or a polling or both.
package blocksub

import (
	"context"
	"time"

	"github.com/ethereum/go-ethereum"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/log"
	"go.uber.org/atomic"
)

type BlockSub struct {
	PollTimeout time.Duration // 10 seconds by default (8,640 requests per day)
	SubTimeout  time.Duration // 60 seconds by default, after this timeout the subscriber will reconnect
	DebugOutput bool

	ethNodeHTTPURI      string // usually port 8545
	ethNodeWebsocketURI string // usually port 9546

	headerC chan<- *ethtypes.Header
	ctx     context.Context

	httpClient      *ethclient.Client
	wsClient        *ethclient.Client
	wsClientSub     ethereum.Subscription
	internalHeaderC chan *ethtypes.Header // internal subscription channel

	CurrentHeader      *ethtypes.Header
	CurrentBlockNumber uint64
	CurrentBlockHash   string

	latestWsHeader *ethtypes.Header
	wsIsConnecting atomic.Bool
}

func NewBlockSub(ctx context.Context, ethNodeHTTPURI string, ethNodeWebsocketURI string, ch chan<- *ethtypes.Header) *BlockSub {
	sub := &BlockSub{
		PollTimeout:         10 * time.Second,
		SubTimeout:          60 * time.Second,
		ethNodeHTTPURI:      ethNodeHTTPURI,
		ethNodeWebsocketURI: ethNodeWebsocketURI,
		headerC:             ch,
		ctx:                 ctx,
		internalHeaderC:     make(chan *ethtypes.Header),
	}
	return sub
}

func (s *BlockSub) Start() (err error) {
	go s.runListener()

	if s.ethNodeHTTPURI != "" {
		log.Info("BlockSub:Start - HTTP connecting...", "uri", s.ethNodeHTTPURI)
		s.httpClient, err = ethclient.Dial(s.ethNodeHTTPURI)
		if err != nil {
			return err
		}
		log.Info("BlockSub:Start - HTTP connected", "uri", s.ethNodeHTTPURI)
		go s.runPollThread()
	}

	if s.ethNodeWebsocketURI != "" {
		go s.startWebsocket()
	}

	return nil
}

// Listens to internal headers and forwards them to the subscriber if it's the header has a greater blockNumber or different hash than the previous one.
// Quits if the context is done.
func (s *BlockSub) runListener() {
	for {
		select {
		case <-s.ctx.Done():
			return
		case header := <-s.internalHeaderC:
			// use the new header if it's later or has a different hash than the previous known one
			if header.Number.Uint64() >= s.CurrentBlockNumber && header.Hash().Hex() != s.CurrentBlockHash {
				s.CurrentHeader = header
				s.CurrentBlockNumber = header.Number.Uint64()
				s.CurrentBlockHash = header.Hash().Hex()

				// Send to subscriber
				s.headerC <- header
			}
		}
	}
}

func (s *BlockSub) runPollThread() {
	for {
		if s.ctx.Err() != nil {
			return
		}

		header, err := s.httpClient.HeaderByNumber(s.ctx, nil)
		if err != nil {
			log.Error("BlockSub: polling latest block failed", "err", err)
			time.Sleep(s.PollTimeout)
			continue
		}

		if s.DebugOutput {
			log.Debug("BlockSub: polled block", "number", header.Number.Uint64(), "hash", header.Hash().Hex())
		}
		s.internalHeaderC <- header

		// Ensure websocket is still working (force a reconnect if it lags behind)
		if s.latestWsHeader != nil && s.latestWsHeader.Number.Uint64() < header.Number.Uint64()-2 {
			log.Warn("BlockSub: forcing websocket reconnect from polling", "wsBlockNum", s.latestWsHeader.Number.Uint64(), "pollBlockNum", header.Number.Uint64())
			go s.startWebsocket()
		}

		time.Sleep(s.PollTimeout)
	}
}

// startWebsocket repeatedly tries to establish a websocket connection to the node until it is connected.
// If another instance is connecting it will return immediately.
func (s *BlockSub) startWebsocket() {
	if isAlreadyConnecting := s.wsIsConnecting.Swap(true); isAlreadyConnecting {
		return
	}

	defer s.wsIsConnecting.Store(false)

	for {
		if s.wsClient != nil {
			s.wsClient.Close()
		}

		err := s._startWebsocket()
		if err != nil {
			log.Error("BlockSub: Websocket connection failed", "err", err)
		} else {
			return
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
		var t = time.NewTimer(s.SubTimeout)

		for {
			select {
			case err := <-s.wsClientSub.Err():
				if err == nil { // shutdown
					return
				}

				// reconnect
				log.Error("BlockSub: headerSub failed, reconnect now", "err", err)
				s.startWebsocket()
				return

			case <-t.C:
				log.Error("BlockSub: timeout, reconnect now", "timeout", s.SubTimeout)
				s.startWebsocket()
				return

			case header := <-wsHeaderC:
				t.Reset(s.SubTimeout)
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
