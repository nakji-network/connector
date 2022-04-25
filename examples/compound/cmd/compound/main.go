package main

import (
	"path/filepath"
	"runtime"

	"github.com/rs/zerolog/log"

	"github.com/nakji-network/connector"
	"github.com/nakji-network/connector/config"
	"github.com/nakji-network/connector/examples/compound"
	"github.com/nakji-network/connector/kafkautils"
	"github.com/nakji-network/connector/protoregistry"
)

func main() {
	conf := config.GetConfig()
	conf.SetDefault("kafka.url", "localhost:9092")
	conf.SetDefault("kafka.env", "dev")
	conf.SetDefault("compound.kafka.topic.mint", ".fct.nakji.compound.0_0_0.compound_mint")
	conf.SetDefault("compound.kafka.topic.redeem", ".fct.nakji.compound.0_0_0.compound_redeem")
	conf.SetDefault("compound.kafka.topic.borrow", ".fct.nakji.compound.0_0_0.compound_borrow")
	conf.SetDefault("compound.kafka.topic.repayborrow", ".fct.nakji.compound.0_0_0.compound_repayborrow")
	conf.SetDefault("compound.kafka.topic.liquidateborrow", ".fct.nakji.compound.0_0_0.compound_liquidateborrow")
	conf.SetDefault("compound.kafka.txID", "compound-source")
	conf.SetDefault("compound.contracts.address", []string{
		"0x4Ddc2D193948926D02f9B1fE9e1daa0718270ED5",
		"0x6C8c6b02E7b2BE14d4fA6022Dfd6d75921D90E4E",
		"0x70e36f6BF80a52b3B46b3aF8e106CC0ed743E8e4",
		"0x5d3a536E4D6DbD6114cc1Ead35777bAB948E3643",
		"0x4Ddc2D193948926D02f9B1fE9e1daa0718270ED5",
		"0xFAce851a4921ce59e912d19329929CE6da6EB0c7",
		"0x158079Ee67Fce2f58472A96584A73C7Ab9AC95c1",
		"0xF5DCe57282A584D2746FaF1593d3121Fcac444dC",
		"0x12392F67bdf24faE0AF363c24aC620a2f67DAd86",
		"0x35A18000230DA775CAc24873d00Ff85BccdeD550",
		"0x39AA39c021dfbaE8faC545936693aC917d5E7563",
		"0xf650C3d88D12dB855b8bf7D11Be6C55A4e07dCC9",
		"0xC11b1268C1A384e55C48c2391d8d480264A3A7F4",
		"0xccF4429DB6322D5C611ee964527D42E5d685DD6a",
		"0xB3319f5D18Bc0D84dD1b4825Dcde5d5f7266d407",
	})
	conf.SetDefault("protoregistry.host", "localhost:9191")

	kafkautils.TopicTypeRegistry.Load(compound.TopicTypes)

	topics := map[string]kafkautils.Topic{
		"mint":            kafkautils.MustParseTopic(conf.GetString("compound.kafka.topic.mint"), conf.GetString("kafka.env")),
		"redeem":          kafkautils.MustParseTopic(conf.GetString("compound.kafka.topic.redeem"), conf.GetString("kafka.env")),
		"borrow":          kafkautils.MustParseTopic(conf.GetString("compound.kafka.topic.borrow"), conf.GetString("kafka.env")),
		"repayborrow":     kafkautils.MustParseTopic(conf.GetString("compound.kafka.topic.repayborrow"), conf.GetString("kafka.env")),
		"liquidateborrow": kafkautils.MustParseTopic(conf.GetString("compound.kafka.topic.liquidateborrow"), conf.GetString("kafka.env")),
	}

	rawAddrs := conf.GetStringSlice("compound.contracts.address")
	addresses := compound.ConvertRawAddress(rawAddrs...)

	_, filename, _, _ := runtime.Caller(0)
	path := filepath.Join(filepath.Dir(filename), "../..", "manifest.yaml")
	c := connector.NewConnector(path)

	// Create a proto registry client
	prc := protoregistry.NewClient(conf.GetString("protoregistry.host"))

	// Register topic and protobuf type mappings
	err := prc.RegisterDynamicTopics(compound.TopicTypes, kafkautils.Fct)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to register dynamic topics")
	}

	compoundConnector := compound.Connector{
		Connector: c,

		// Any additional custom connections not supported natively by Nakji
		// Client: DogecoinClient(context.Background()),

		// Any additional command line arguments, such as chain selection override. Set up via https://pkg.go.dev/github.com/spf13/viper#readme-working-with-flags
		// Chain: "bsc",

		// Any additional config vars from the config yaml, using functions from Viper (https://pkg.go.dev/github.com/spf13/viper#readme-getting-values-from-viper)
		// This is namespaced via connector id (author-name-version)
		// CustomOption: c.Config.GetString("custom_option"),
		Topics:            topics,
		ContractAddresses: addresses,
	}

	compoundConnector.Start()
}
