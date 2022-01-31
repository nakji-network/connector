// This connector ingests real time data from any EVM compatible chain.
package main

import (
	"github.com/nakji-network/connector"
	"github.com/nakji-network/connector/examples/ethereum"
)

func main() {
	c := connector.NewConnector()

	ethConnector := ethereum.EthereumConnector{
		Connector: c,

		// Any additional custom connections not supported natively by Nakji
		// Client: c.ChainClients.Ethereum(context.Background()),
	}

	ethConnector.Start()
}
