# chain

Support built-in RPC clients so that connectors can enable them with a single line, eg.

```go
    ethClient := connector.ChainClients.Ethereum(context.Background())
    avaxClient := connector.ChainClients.Avalanche(context.Background())
```

RPC endpoints are sourced from the main config yaml. This package supports 
clients that offer multiple connections and https+websocket, or may return
custom structs to add this functionality on a chain by chain basis.

## Contributing

To contribute, create a new go file and follow the following example. The
method name should be the chain name when the client returned is the most 
commonly used/standard client, a modified+compatible version of the client, or 
when there is not a Golang library available.

Standard-incompatible custom clients should use use 
`<Blockchain name><Name of client implementation>` as the function name to
prevent confusion.

```go
func (c Clients) <BlockchainName>(ctx context.Context) <ClientStruct> {
  // Read config from config yaml under `rpcs.<BlockchainName>`
  rpcURLs := c.rpcMap["<BlockchainName>"].Full

	log.Info().
    Strs("RPCs", rpcURLs).
    Msg("connecting to <BlockchainName> RPC")

  // <Connect to RPCs using a library or custom client, eg. github.com/ethereum/go-ethereum/ethclient>"
  if err != nil {
    // Instead of returning an error, exit with Fatal error
    log.Fatal().Err(err).Msg("BlockchainName RPC connection error")
  }

  return client
}
```

`ethereum.go` is a reference example.

