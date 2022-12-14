// This connector ingests real time data from any EVM compatible chain.
package main

import (
	"github.com/nakji-network/connector/examples/ethereum"
)

func main() {
	// These values can be hardcoded or passed via CLI or yaml file
	blockchain := "ethereum"
	var fromBlock uint64 = 0
	var numBlocks uint64 = 0

	ec := ethereum.NewConnector(blockchain, fromBlock, numBlocks)
	ec.Start()
}
