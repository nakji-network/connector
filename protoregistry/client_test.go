package protoregistry

import (
	"bytes"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/rs/zerolog/log"
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

var desc []byte

func readDescriptor() {
	wd, err := os.Getwd()
	if err != nil {
		log.Fatal().Err(err).Msg("error getting pwd (required for relative imports)")
	}

	var file string

	err = filepath.WalkDir(wd, func(path string, info os.DirEntry, err error) error {
		if err != nil {
			log.Error().Err(err).Msg("WalkDirFunc error")
			return err
		}

		if !info.IsDir() && filepath.Base(path) == "prtest.proto.desc" {
			file = path
			log.Debug().Str("path", path).Msg("found the proto file")
			return errProtoFound
		}

		return nil
	})

	if err != nil && err != errProtoFound {
		log.Fatal().Err(err).Msg("error finding prtest.proto.desc")
	}

	desc, err = ioutil.ReadFile(file)
	if err != nil {
		log.Fatal().Err(err).Msg("error reading prtest.proto.desc")
	}
}

func setup() {
	readDescriptor()
}

func TestMain(m *testing.M) {
	setup()
	exitCode := m.Run()
	os.Exit(exitCode)
}

func Test_buildTopicProtoMsgs(t *testing.T) {
	want := map[string]TopicProtoMsg{
		"prtest.Mint":            {"sys", "nakji.protoregistry.0_0_0.prtest_mint", "prtest.Mint", desc},
		"prtest.Redeem":          {"sys", "nakji.protoregistry.0_0_0.prtest_redeem", "prtest.Redeem", desc},
		"prtest.Borrow":          {"sys", "nakji.protoregistry.0_0_0.prtest_borrow", "prtest.Borrow", desc},
		"prtest.RepayBorrow":     {"sys", "nakji.protoregistry.0_0_0.prtest_repayborrow", "prtest.RepayBorrow", desc},
		"prtest.LiquidateBorrow": {"sys", "nakji.protoregistry.0_0_0.prtest_liquidateborrow", "prtest.LiquidateBorrow", desc},
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

func Test_generateDescriptorFiles(t *testing.T) {
	tpmList := []*TopicProtoMsg{
		{"sys", "nakji.protoregistry.0_0_0.prtest_mint", "prtest.Mint", nil},
		{"sys", "nakji.protoregistry.0_0_0.prtest_redeem", "prtest.Redeem", nil},
		{"sys", "nakji.protoregistry.0_0_0.prtest_borrow", "prtest.Borrow", nil},
		{"sys", "nakji.protoregistry.0_0_0.prtest_repayborrow", "prtest.RepayBorrow", nil},
		{"sys", "nakji.protoregistry.0_0_0.prtest_liquidateborrow", "prtest.LiquidateBorrow", nil},
	}

	err := generateDescriptorFiles(tpmList)
	if err != nil {
		t.Fatalf("failed to generate descriptor files %v", err)
	}

	for _, tpm := range tpmList {
		if bytes.Compare(tpm.Descriptor, desc) != 0 {
			t.Errorf("expected Descriptor got = %v, want = %v", tpm.Descriptor, desc)
		}
	}
}
