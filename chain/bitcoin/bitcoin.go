package bitcoin

import (
	"net/url"

	"github.com/nakji-network/connector"

	"github.com/btcsuite/btcd/rpcclient"
	"github.com/rs/zerolog/log"
)

const chain = "bitcoin"

// Config yaml(local.yaml) file must include the following
//	rpcs:
//		bitcoin:
//			url: wss://blep:password@localhost:8335
//			cert: |
//				-----BEGIN CERTIFICATE-----
//				<your bitcoin rpc certificate for ssl>
//				-----END CERTIFICATE-----

type Connector struct {
	*connector.Connector
	*rpcclient.Client
}

func NewConnector(ntfnHandlers *rpcclient.NotificationHandlers) *Connector {
	c, err := connector.NewConnector()
	if err != nil {
		log.Fatal().Err(err).Msg("failed to instantiate nakji connector")
	}

	// Read config from config yaml under `rpcs.bitcoin.url`
	rpc := c.RPCMap[chain].Url

	// bitcoin client only supports 1 rpc connection currently, so we do this hack
	RPCURL, err := url.Parse(rpc)
	if err != nil {
		log.Fatal().Err(err).Msg("invalid rpc url")
	}

	// Connect to local btcd RPC server using websockets.
	pw, _ := RPCURL.User.Password()
	connCfg := &rpcclient.ConnConfig{
		Host:         RPCURL.Host,
		Endpoint:     "ws",
		User:         RPCURL.User.Username(),
		Pass:         pw,
		Certificates: []byte(c.RPCMap[chain].Cert),
	}
	client, err := rpcclient.New(connCfg, ntfnHandlers)
	if err != nil {
		log.Fatal().Err(err).Msg("BTC: RPC Client connection error")
	}
	return &Connector{
		c,
		client,
	}
}
