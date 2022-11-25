package protoregistry

import (
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"reflect"
	"testing"

	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protodesc"
	"google.golang.org/protobuf/types/descriptorpb"

	"github.com/nakji-network/connector/protoregistry/prtest"
)

const host = "127.0.0.1:12345"

var topicTypes = map[string]proto.Message{
	"nakji.protoregistry.0_0_0.prtest_mint":            &prtest.Mint{},
	"nakji.protoregistry.0_0_0.prtest_redeem":          &prtest.Redeem{},
	"nakji.protoregistry.0_0_0.prtest_borrow":          &prtest.Borrow{},
	"nakji.protoregistry.0_0_0.prtest_repayborrow":     &prtest.RepayBorrow{},
	"nakji.protoregistry.0_0_0.prtest_liquidateborrow": &prtest.LiquidateBorrow{},
}

func TestMain(m *testing.M) {
	exitCode := m.Run()
	os.Exit(exitCode)
}

func TestRegisterDynamicTopics(t *testing.T) {
	handler := http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {})

	s, err := NewTestServerWithURL(host, handler)
	defer s.Close()
	if err != nil {
		t.Fatalf("failed to create new test server: %v", err)
	}

	err = RegisterDynamicTopics(host, topicTypes, "cmd")
	if err != nil {
		t.Fatalf("failed to register dynamic topics: %v", err)
	}

}

func Test_buildTopicProtoMsgs(t *testing.T) {
	toFileDescriptorProtoFunc := func(message proto.Message) *descriptorpb.FileDescriptorProto {
		return protodesc.ToFileDescriptorProto(message.ProtoReflect().Descriptor().ParentFile())
	}

	want := map[string]TopicProtoMsg{
		"prtest.Mint":            {"sys", "nakji.protoregistry.0_0_0.prtest_mint", "prtest.Mint", toFileDescriptorProtoFunc(&prtest.Mint{}), nil},
		"prtest.Redeem":          {"sys", "nakji.protoregistry.0_0_0.prtest_redeem", "prtest.Redeem", toFileDescriptorProtoFunc(&prtest.Redeem{}), nil},
		"prtest.Borrow":          {"sys", "nakji.protoregistry.0_0_0.prtest_borrow", "prtest.Borrow", toFileDescriptorProtoFunc(&prtest.Borrow{}), nil},
		"prtest.RepayBorrow":     {"sys", "nakji.protoregistry.0_0_0.prtest_repayborrow", "prtest.RepayBorrow", toFileDescriptorProtoFunc(&prtest.RepayBorrow{}), nil},
		"prtest.LiquidateBorrow": {"sys", "nakji.protoregistry.0_0_0.prtest_liquidateborrow", "prtest.LiquidateBorrow", toFileDescriptorProtoFunc(&prtest.LiquidateBorrow{}), nil},
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
		if tpm.Descriptor != nil {
			t.Errorf("Descriptor got = %v, want = %v", tpm.Descriptor, nil)
		}
		if !reflect.DeepEqual(tpm.FileDescriptorProto, want[tpm.ProtoMsgName].FileDescriptorProto) {
			t.Errorf("FileDescriptorProto got = %v, want = %v", tpm.FileDescriptorProto, want[tpm.ProtoMsgName].FileDescriptorProto)
		}
	}
}

func NewTestServerWithURL(URL string, handler http.Handler) (*httptest.Server, error) {
	ts := httptest.NewUnstartedServer(handler)
	if URL != "" {
		l, err := net.Listen("tcp", URL)
		if err != nil {
			return nil, err
		}
		ts.Listener.Close()
		ts.Listener = l
	}
	ts.Start()
	return ts, nil
}
