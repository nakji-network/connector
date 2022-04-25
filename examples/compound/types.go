package compound

import (
	"github.com/nakji-network/connector/examples/compound/ctoken"

	"google.golang.org/protobuf/proto"
)

var TopicTypes = map[string]proto.Message{
	"nakji.compound.0_0_0.compound_mint":            &ctoken.Mint{},
	"nakji.compound.0_0_0.compound_redeem":          &ctoken.Redeem{},
	"nakji.compound.0_0_0.compound_borrow":          &ctoken.Borrow{},
	"nakji.compound.0_0_0.compound_repayborrow":     &ctoken.RepayBorrow{},
	"nakji.compound.0_0_0.compound_liquidateborrow": &ctoken.LiquidateBorrow{},
}
