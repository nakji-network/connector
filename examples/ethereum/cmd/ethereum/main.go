// This connector ingests real time data from any EVM compatible chain. After
// compiling this program, you can specify which chain to connect to as follows:
//
//     bin/ethereum -n avalanche
package main

import (
	"strings"

	"github.com/nakji-network/connector"
	"github.com/nakji-network/connector/examples/ethereum"
)

func main() {
	c := connector.NewConnector()

	// Get config variables using functions from Viper (https://pkg.go.dev/github.com/spf13/viper#readme-getting-values-from-viper)
	RPCURLs := c.Config.GetStringSlice("rpcs.ethereum.full")

	// For the purposes of this example, we'll just grab one of the websocket RPCs
	var RPCURL string
	for _, u := range RPCURLs {
		if strings.HasPrefix(u, "ws") {
			RPCURL = u
			break
		}
	}

	ethConnector := ethereum.EthereumConnector{
		Connector: c,
		RPCURL:   RPCURL,
	}

	ethConnector.Start()
}
