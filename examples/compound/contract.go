package compound

import (
	"strings"

	"github.com/nakji-network/connector/examples/compound/cBAT"
	"github.com/nakji-network/connector/examples/compound/cCOMP"
	"github.com/nakji-network/connector/examples/compound/cDAI"
	"github.com/nakji-network/connector/examples/compound/cETH"
	"github.com/nakji-network/connector/examples/compound/cLINK"
	"github.com/nakji-network/connector/examples/compound/cREP"
	"github.com/nakji-network/connector/examples/compound/cSAI"
	"github.com/nakji-network/connector/examples/compound/cTUSD"
	"github.com/nakji-network/connector/examples/compound/cUNI"
	"github.com/nakji-network/connector/examples/compound/cUSDC"
	"github.com/nakji-network/connector/examples/compound/cUSDT"
	"github.com/nakji-network/connector/examples/compound/cWBTC"
	"github.com/nakji-network/connector/examples/compound/cWBTC2"
	"github.com/nakji-network/connector/examples/compound/cZRX"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/rs/zerolog/log"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/timestamppb"
)

var ContractAddresses = map[string]string{
	"cETH":   "0x4Ddc2D193948926D02f9B1fE9e1daa0718270ED5",
	"cBAT":   "0x6C8c6b02E7b2BE14d4fA6022Dfd6d75921D90E4E",
	"cCOMP":  "0x70e36f6BF80a52b3B46b3aF8e106CC0ed743E8e4",
	"cDAI":   "0x5d3a536E4D6DbD6114cc1Ead35777bAB948E3643",
	"cLINK":  "0xFAce851a4921ce59e912d19329929CE6da6EB0c7",
	"cREP":   "0x158079Ee67Fce2f58472A96584A73C7Ab9AC95c1",
	"cSAI":   "0xF5DCe57282A584D2746FaF1593d3121Fcac444dC",
	"cTUSD":  "0x12392f67bdf24fae0af363c24ac620a2f67dad86",
	"cUNI":   "0x35a18000230da775cac24873d00ff85bccded550",
	"cUSDC":  "0x39AA39c021dfbaE8faC545936693aC917d5E7563",
	"cUSDT":  "0xf650C3d88D12dB855b8bf7D11Be6C55A4e07dCC9",
	"cWBTC":  "0xC11b1268C1A384e55C48c2391d8d480264A3A7F4",
	"cWBTC2": "0xccF4429DB6322D5C611ee964527D42E5d685DD6a",
	"cZRX":   "0xB3319f5D18Bc0D84dD1b4825Dcde5d5f7266d407",
}

var ABIs = map[string]string{
	"cETH":   cETH.CETHMetaData.ABI,
	"cBAT":   cBAT.CBATMetaData.ABI,
	"cCOMP":  cCOMP.CCOMPMetaData.ABI,
	"cDAI":   cDAI.CDAIMetaData.ABI,
	"cLINK":  cLINK.CLINKMetaData.ABI,
	"cREP":   cREP.CREPMetaData.ABI,
	"cSAI":   cSAI.CSAIMetaData.ABI,
	"cTUSD":  cTUSD.CTUSDMetaData.ABI,
	"cUNI":   cUNI.CUNIMetaData.ABI,
	"cUSDC":  cUSDC.CUSDCMetaData.ABI,
	"cUSDT":  cUSDT.CUSDTMetaData.ABI,
	"cWBTC":  cWBTC.CWBTCMetaData.ABI,
	"cWBTC2": cWBTC2.CWBTC2MetaData.ABI,
	"cZRX":   cZRX.CZRXMetaData.ABI,
}

var EventParsers = map[string]ProtoMessageGetter{
	"cETH":   &cETH.EventParser{},
	"cBAT":   &cBAT.EventParser{},
	"cCOMP":  &cCOMP.EventParser{},
	"cDAI":   &cDAI.EventParser{},
	"cLINK":  &cLINK.EventParser{},
	"cREP":   &cREP.EventParser{},
	"cSAI":   &cSAI.EventParser{},
	"cTUSD":  &cTUSD.EventParser{},
	"cUNI":   &cUNI.EventParser{},
	"cUSDC":  &cUSDC.EventParser{},
	"cUSDT":  &cUSDT.EventParser{},
	"cWBTC":  &cWBTC.EventParser{},
	"cWBTC2": &cWBTC2.EventParser{},
	"cZRX":   &cZRX.EventParser{},
}

type ProtoMessageGetter interface {
	Get(eventName string, contractAbi *abi.ABI, vLog types.Log, timestamp *timestamppb.Timestamp) proto.Message
}

type Contract struct {
	Name string
	ABI  *abi.ABI
	Pmg  ProtoMessageGetter
}

func BuildContracts(addresses map[string]string) map[string]*Contract {
	contracts := make(map[string]*Contract)

	for k, v := range addresses {
		if abiStr, ok := ABIs[k]; ok {
			abiObj, err := abi.JSON(strings.NewReader(abiStr))
			if err != nil {
				log.Fatal().Err(err).Msg("failed to read contract ABI")
			}

			contracts[v] = &Contract{
				ABI:  &abiObj,
				Name: k,
				Pmg:  EventParsers[k],
			}
		}
	}

	return contracts
}

func GetAddresses(addresses map[string]string) []common.Address {
	addressSlice := make([]common.Address, len(addresses))
	i := 0
	for _, v := range addresses {
		addressSlice[i] = common.HexToAddress(v)
		i++
	}

	return addressSlice
}
