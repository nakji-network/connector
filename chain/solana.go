package chain

import (
	"context"

	"github.com/gagliardetto/solana-go/rpc"
	"github.com/gagliardetto/solana-go/rpc/ws"
	"github.com/rs/zerolog/log"
)

const chain = "solana"

// Config yaml(local.yaml) file must include the following
//	rpcs:
//		solana:
//			url: https://rpcprovider.com/

func (c Clients) Solana() (*rpc.Client, *ws.Client) {
	client := rpc.New(c.rpcMap[chain].Url)

	wsClient, err := ws.Connect(context.Background(), c.rpcMap[chain].Url)
	if err != nil {
		log.Fatal().Err(err).Msg("Solana: WS Client connection error")
	}

	return client, wsClient
}
