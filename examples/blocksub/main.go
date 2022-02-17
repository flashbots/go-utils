package main

import (
	"context"
	"os"

	ethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/log"
	"github.com/flashbots/go-utils/blocksub"
	"github.com/flashbots/go-utils/cli"
)

var (
	httpURI  = os.Getenv("ETH_HTTP") // usually port 8545
	wsURI    = os.Getenv("ETH_WS")   // usually port 8546
	logJSON  = os.Getenv("LOG_JSON") == "1"
	logDebug = os.Getenv("DEBUG") == "1"
)

func main() {
	// Setup logging
	logLevel := log.LvlInfo
	if logDebug {
		logLevel = log.LvlDebug
	}

	logFormat := log.TerminalFormat(true)
	if logJSON {
		logFormat = log.JSONFormat()
	}

	log.Root().SetHandler(log.LvlFilterHandler(logLevel, log.StreamHandler(os.Stderr, logFormat)))

	// Setup BlockSub
	ch := make(chan *ethtypes.Header)
	blocksub := blocksub.NewBlockSub(context.Background(), httpURI, wsURI, ch)
	blocksub.DebugOutput = true
	err := blocksub.Start()
	cli.CheckErr(err)

	// Wait for new headers to arrive
	for h := range ch {
		log.Info("got header", "number", h.Number.Uint64(), "hash", h.Hash().Hex())
	}
}
