package chain

import (
	"net/url"

	"github.com/btcsuite/btcd/rpcclient"
	"github.com/rs/zerolog/log"
)

// Config yaml(local.yaml) file must include the following
//	rpcs:
//		bitcoin:
//			url: wss://blep:password@localhost:8335
//			cert: |
//				-----BEGIN CERTIFICATE-----
//				<your bitcoin rpc certificate for ssl>
//				-----END CERTIFICATE-----

func (c Clients) Bitcoin(ntfnHandlers *rpcclient.NotificationHandlers, chainOverride ...string) *rpcclient.Client {
	chain := "bitcoin"
	if len(chainOverride) > 0 && chainOverride[0] != "" {
		chain = chainOverride[0]
	}

	// Read config from config yaml under `rpcs.bitcoin`
	rpc := c.rpcMap[chain].Url

	// bitcoin client only supports 1 rpc connection currently so we do this hack
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
		Certificates: []byte(c.rpcMap[chain].Cert),
	}
	client, err := rpcclient.New(connCfg, ntfnHandlers)
	if err != nil {
		log.Fatal().Err(err).Msg("BTC: RPC Client connection error")
	}
	return client
}
