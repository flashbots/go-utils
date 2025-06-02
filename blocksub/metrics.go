package blocksub

import (
	"github.com/VictoriaMetrics/metrics"
)

var blockNumberGauge = metrics.NewGauge(`goutils_blocksub_latest_block_number`, nil)

func setBlockNumber(blockNumber uint64) {
	blockNumberGauge.Set(float64(blockNumber))
}
