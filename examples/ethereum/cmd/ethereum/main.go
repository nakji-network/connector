// This connector ingests real time data from any EVM compatible chain.
package main

import (
	"github.com/nakji-network/connector"
	"github.com/nakji-network/connector/examples/ethereum"

	"github.com/rs/zerolog/log"
)

func main() {
	c, err := connector.NewConnector()
	if err != nil {
		log.Fatal().Err(err).Msg("failed to instantiate connector")
	}

	ethConnector := ethereum.EthereumConnector{
		Connector: c,

		// Any additional custom connections not supported natively by Nakji
		// Client: DogecoinClient(context.Background()),

		// Any additional command line arguments, such as chain selection override. Set up via https://pkg.go.dev/github.com/spf13/viper#readme-working-with-flags
		// Chain: "bsc",

		// Any additional config vars from the config yaml, using functions from Viper (https://pkg.go.dev/github.com/spf13/viper#readme-getting-values-from-viper)
		// This is namespaced via connector id (author-name-version)
		// CustomOption: c.Config.GetString("custom_option"),
	}

	ethConnector.Start()
}
