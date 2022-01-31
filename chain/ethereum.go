package chain

import (
	"context"
	"strings"

	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/rs/zerolog/log"
)

// Ethereum returns an ethereum client connected to websockets RPC
func (c Clients) Ethereum(ctx context.Context, chainOverride ...string) *ethclient.Client {
	chain := "ethereum"
	if len(chainOverride) > 0 && chainOverride[0] != "" {
		chain = chainOverride[0]
	}

	// Read config from config yaml under `rpcs.ethereum`
	rpcs := c.rpcMap[chain].Full

	// go-ethereum client only supports 1 rpc connection currently so we do this hack
	var RPCURL string
	for _, u := range rpcs {
		if strings.HasPrefix(u, "ws") {
			RPCURL = u
			break
		}
	}
	log.Info().Str("url", RPCURL).Msg("connecting to Ethereum RPC")
	client, err := ethclient.DialContext(ctx, RPCURL)
	if err != nil {
		log.Fatal().Err(err).Msg("Ethereum RPC connection error")
	}
	return client
}
