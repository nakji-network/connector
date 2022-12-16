package main

import (
	"github.com/nakji-network/connector/examples/compound"
)

func main() {

	// These values can be hardcoded or passed via CLI or yaml file
	blockchain := "ethereum"
	var fromBlock uint64 = 0
	var numBlocks uint64 = 0

	cc := compound.NewConnector(blockchain, fromBlock, numBlocks)
	cc.Start()
}
