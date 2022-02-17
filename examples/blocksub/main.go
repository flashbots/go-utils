package main

import (
	"context"
	"os"

	"github.com/ethereum/go-ethereum/log"
	"github.com/flashbots/go-utils/blocksub"
)

var (
	httpURI  = os.Getenv("ETH_HTTP") // usually port 8545
	wsURI    = os.Getenv("ETH_WS")   // usually port 8546
	logJSON  = os.Getenv("LOG_JSON") == "1"
	logDebug = os.Getenv("DEBUG") == "1"
)

func logSetup() {
	logLevel := log.LvlInfo
	if logDebug {
		logLevel = log.LvlDebug
	}

	logFormat := log.TerminalFormat(true)
	if logJSON {
		logFormat = log.JSONFormat()
	}

	log.Root().SetHandler(log.LvlFilterHandler(logLevel, log.StreamHandler(os.Stderr, logFormat)))
}

func main() {
	logSetup()

	DemoSimpleSub(httpURI, wsURI)
	// DemoMultiSub(httpURI, wsURI)
}

func DemoSimpleSub(httpURI, wsURI string) {
	// Create and start a BlockSub instance
	blocksub := blocksub.NewBlockSub(context.Background(), httpURI, wsURI)
	blocksub.DebugOutput = true
	if err := blocksub.Start(); err != nil {
		log.Crit(err.Error())
	}

	// Create a subscription to new headers
	sub := blocksub.Subscribe(context.Background())
	for header := range sub.C {
		log.Info("new header", "number", header.Number.Uint64(), "hash", header.Hash().Hex())
	}
}
