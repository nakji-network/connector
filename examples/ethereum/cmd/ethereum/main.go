// This connector ingests real time data from any EVM compatible chain.
package main

import (
	"context"

	"github.com/nakji-network/connector"
	"github.com/nakji-network/connector/examples/ethereum"
)

func main() {
	c := connector.NewConnector()

	ethConnector := ethereum.EthereumConnector{
		Connector: c,

		// connect to Ethereum RPC
		Client: c.ChainClients.Ethereum(context.Background()),
	}
	defer ethConnector.Client.Close()

	ethConnector.Start()
}
