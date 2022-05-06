package protoregistry

import (
	"testing"

	"google.golang.org/protobuf/proto"

	"github.com/nakji-network/connector/protoregistry/prtest"
)

var topicTypes = map[string]proto.Message{
	"nakji.protoregistry.0_0_0.prtest_mint":            &prtest.Mint{},
	"nakji.protoregistry.0_0_0.prtest_redeem":          &prtest.Redeem{},
	"nakji.protoregistry.0_0_0.prtest_borrow":          &prtest.Borrow{},
	"nakji.protoregistry.0_0_0.prtest_repayborrow":     &prtest.RepayBorrow{},
	"nakji.protoregistry.0_0_0.prtest_liquidateborrow": &prtest.LiquidateBorrow{},
}

func Test_buildTopicProtoMsgs(t *testing.T) {
	want := map[string]TopicProtoMsg{
		"prtest.Mint":            {"sys", "nakji.protoregistry.0_0_0.prtest_mint", "prtest.Mint"},
		"prtest.Redeem":          {"sys", "nakji.protoregistry.0_0_0.prtest_redeem", "prtest.Redeem"},
		"prtest.Borrow":          {"sys", "nakji.protoregistry.0_0_0.prtest_borrow", "prtest.Borrow"},
		"prtest.RepayBorrow":     {"sys", "nakji.protoregistry.0_0_0.prtest_repayborrow", "prtest.RepayBorrow"},
		"prtest.LiquidateBorrow": {"sys", "nakji.protoregistry.0_0_0.prtest_liquidateborrow", "prtest.LiquidateBorrow"},
	}

	got := buildTopicProtoMsgs(topicTypes, "sys")

	for _, tpm := range got {
		if tpm.MsgType != "sys" {
			t.Errorf("msg type got = %v, want = %v", tpm.MsgType, "sys")
		}
		if _, ok := want[tpm.ProtoMsgName]; !ok {
			t.Errorf("key missing got = %v", tpm.ProtoMsgName)
			continue
		}
		if tpm.ProtoMsgName != want[tpm.ProtoMsgName].ProtoMsgName {
			t.Errorf("ProtoMsgName got = %v, want = %v", tpm.ProtoMsgName, want[tpm.ProtoMsgName].ProtoMsgName)
		}
		if tpm.TopicName != want[tpm.ProtoMsgName].TopicName {
			t.Errorf("TopicName got = %v, want = %v", tpm.TopicName, want[tpm.ProtoMsgName].TopicName)
		}
	}
}
