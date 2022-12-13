// This connector ingests real time data from any EVM compatible chain.
package main

import (
	"github.com/nakji-network/connector/examples/ethereum"
)

func main() {
	ec := ethereum.NewConnector()
	ec.Start()
}
