package generic

import (
	"context"
	// "fmt"
	"testing"

	"github.com/cloudwego/kitex/internal/mocks"
	"github.com/cloudwego/kitex/internal/test"
	"github.com/cloudwego/kitex/pkg/remote"
	"github.com/cloudwego/kitex/pkg/rpcinfo"
	"github.com/cloudwego/kitex/transport"
	// "google.golang.org/protobuf/proto"
)

type MockRequest struct {
	Msg     string            `protobuf:"bytes,1,opt,name=Msg,proto3" json:"Msg,omitempty"`
	StrMap  map[string]string `protobuf:"bytes,2,rep,name=StrMap,proto3" json:"StrMap,omitempty"`
	StrList []string          `protobuf:"bytes,3,rep,name=StrList,proto3" json:"StrList,omitempty"`
}

func (m *MockRequest) Reset() { *m = MockRequest{} }

// func (m *MockRequest) String() string { return proto.CompactTextString(m) }
func (*MockRequest) ProtoMessage() {}

func TestJsonProtobufCodec(t *testing.T) {
	mainProtoContent := `
		syntax = "proto3";
		package test;

		message MockRequest {
			string Msg = 1;
			map<string, string> StrMap = 2;
			repeated string StrList = 3;
		}
		
		service Mock {
			rpc Test(MockRequest) returns (MockRequest);
		}
	`
	p, err := NewPbContentProvider("mock.proto", map[string]string{"mock.proto": mainProtoContent})
	test.Assert(t, err == nil)
	jpc, err := newJsonProtobufCodec(p, protoCodec)
	test.Assert(t, err == nil)
	defer jpc.Close()
	test.Assert(t, jpc.Name() == "JSONProtobuf")

	ctx := context.Background()
	// sendMsg := initProtoSendMsg(transport.TTHeader)

	// Marshal side
	out := remote.NewWriterBuffer(256)
	emptyMethodInk := rpcinfo.NewInvocation("", "")
	emptyMethodRi := rpcinfo.NewRPCInfo(nil, nil, emptyMethodInk, nil, nil)
	emptyMethodMsg := remote.NewMessage(nil, nil, emptyMethodRi, remote.Exception, remote.Client)
	err = jpc.Marshal(ctx, emptyMethodMsg, out)
	test.Assert(t, err.Error() == "empty methodName in protobuf Marshal")

	// Exception MsgType test
	exceptionMsgTypeInk := rpcinfo.NewInvocation("", "Test")
	exceptionMsgTypeRi := rpcinfo.NewRPCInfo(nil, nil, exceptionMsgTypeInk, nil, nil)
	exceptionMsgTypeMsg := remote.NewMessage(&remote.TransError{}, nil, exceptionMsgTypeRi, remote.Exception, remote.Client)
	// Marshal side
	err = jpc.Marshal(ctx, exceptionMsgTypeMsg, out)
	test.Assert(t, err == nil)

	// err = jpc.Marshal(ctx, emptyMethodMsg, out)
	// fmt.Print(err)
	// test.Assert(t, err == nil)

	// // UnMarshal side
	// recvMsg := initProtoRecvMsg()
	// buf, err := out.Bytes()
	// test.Assert(t, err == nil)
	// recvMsg.SetPayloadLen(len(buf))
	// in := remote.NewReaderBuffer(buf)
	// err = jpc.Unmarshal(ctx, recvMsg, in)
	// test.Assert(t, err == nil)
}

func initProtoSendMsg(tp transport.Protocol) remote.Message {
	req := &MockRequest{
		Msg: "Test",
		StrMap: map[string]string{
			"key": "value",
		},
		StrList: []string{"Test"},
	}
	svcInfo := mocks.ServiceInfo()
	ink := rpcinfo.NewInvocation("", "Test")
	ri := rpcinfo.NewRPCInfo(nil, nil, ink, nil, rpcinfo.NewRPCStats())
	msg := remote.NewMessage(req, svcInfo, ri, remote.Call, remote.Client)
	msg.SetProtocolInfo(remote.NewProtocolInfo(tp, svcInfo.PayloadCodec))
	return msg
}

func initProtoRecvMsg() remote.Message {
	req := &MockRequest{
		Msg: "Test",
		StrMap: map[string]string{
			"key": "value",
		},
		StrList: []string{"Test"},
	}
	ink := rpcinfo.NewInvocation("", "Test")
	ri := rpcinfo.NewRPCInfo(nil, nil, ink, nil, rpcinfo.NewRPCStats())
	msg := remote.NewMessage(req, mocks.ServiceInfo(), ri, remote.Call, remote.Server)
	return msg
}
