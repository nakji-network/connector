// ethereum package follows https://goethereumbook.org/block-subscribe/ to
// subscribe to new Blocks and Transactions and writes the results to Nakji.
package ethereum

import (
	"context"
	"fmt"
	"os"
	"os/signal"

	"github.com/ethereum/go-ethereum/core/types"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/rs/zerolog/log"

	"github.com/nakji-network/connector"
)

type EthereumConnector struct {
	*connector.Connector

	RPCURL string
}

func (c *EthereumConnector) Start() {
	// Listen for interrupt in order to cleanly close connections later
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	// connect to Ethereum RPC websockets
	log.Info().Str("url", c.RPCURL).Msg("connecting to Ethereum RPC")
	client, err := ethclient.DialContext(context.Background(), c.RPCURL)
	if err != nil {
		log.Fatal().Err(err).Msg("Ethereum RPC connection error")
	}
	defer client.Close()

	// Subscribe to headers
  headers := make(chan *types.Header)
	sub, err := client.SubscribeNewHead(context.Background(), headers)
	if err != nil {
		log.Fatal().Err(err)
	}

	// Main loop to process errors and headers
	go func() {
		for {
			select {
			case err := <-sub.Err():
				log.Fatal().Err(err)
			case header := <-headers:
				block, err := client.BlockByHash(context.Background(), header.Hash())
				if err != nil {
					log.Fatal().Err(err).Msg("BlockByHash error")
				}

				PrintBlock(block)

				// EthBlock -> Block -> Protobuf -> kafka
				var blockData Block
				blockData.UnmarshalEthBlock(block)
				err = c.ProduceMessage("ethereum", "ethereum", &blockData)
				if err != nil {
					log.Error().Err(err).Msg("Kafka write proto")
				}

				// EthTransaction -> Transaction -> Protobuf -> Kafka
				for _, tx := range block.Transactions() {
					txData := Transaction{}
					txData.UnmarshalEthTransaction(tx)
					txData.Timestamp = blockData.Timestamp // Timestamp isn't in the raw transaction from geth

					err = c.ProduceMessage("ethereum", "ethereum", &txData)
					if err != nil {
						log.Error().Err(err).Msg("Kafka write proto")
					}
				}

				// Commit Kafka Transaction
				err = c.Producer.CommitTransaction(nil)
				if err != nil {
					log.Error().Err(err).Msg("Processor: Failed to commit transaction")

					err = c.Producer.AbortTransaction(nil)
					if err != nil {
						log.Fatal().Err(err).Msg("")
					}
				}
				// Start a new transaction
				err = c.BeginTransaction()
				if err != nil {
					log.Fatal().Err(err).Msg("")
				}

			}
		}
	}()

	for {
		select {
		case <-interrupt:
			log.Debug().Msg("interrupt")

			// Cleanly close the connection by sending a close message and then
			// waiting (with timeout) for the server to close the connection.
			client.Close()
			c.Producer.Close()
			return
		}
	}
}

func PrintBlock(block *ethtypes.Block) {
	fmt.Printf("hash: %s\n", block.Hash().Hex())       // 0xbc10defa8dda384c96a17640d84de5578804945d347072e091b4e5f390ddea7f
	fmt.Printf("num: %v\n", block.Number().Uint64())   // 3477413
	fmt.Printf("time: %v\n", block.Time())             // 1529525947
	fmt.Printf("nonce: %v\n", block.Nonce())           // 130524141876765836
	fmt.Printf("#tx: %v\n", len(block.Transactions())) // 7
	fmt.Printf("gaslim: %v\n", block.GasLimit())       // 1529525947
	fmt.Printf("gasuse: %v\n", block.GasUsed())        // 1529525947
	fmt.Printf("diff: %v\n", block.Difficulty())       // 1529525947
}
