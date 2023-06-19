package generic

import (
	"context"
	"sync/atomic"

	"github.com/cloudwego/kitex/pkg/generic/descriptor"
	"github.com/cloudwego/kitex/pkg/remote"
	"github.com/cloudwego/kitex/pkg/remote/codec"
	"github.com/cloudwego/kitex/pkg/remote/codec/perrors"
	"github.com/cloudwego/kitex/pkg/serviceinfo"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

var (
	_ remote.PayloadCodec = &jsonProtobufCodec{}
	_ Closer              = &jsonProtobufCodec{}
)

type jsonProtobufCodec struct {
	svcDsc   atomic.Value // *idl
	provider PbDescriptorProvider
	codec    remote.PayloadCodec
}

func newJsonProtobufCodec(p PbDescriptorProvider, codec remote.PayloadCodec) (*jsonProtobufCodec, error) {
	svc := <-p.Provide()
	c := &jsonProtobufCodec{codec: codec, provider: p}
	c.svcDsc.Store(svc)
	go c.update()
	return c, nil
}

func (c *jsonProtobufCodec) update() {
	for {
		svc, ok := <-c.provider.Provide()
		if !ok {
			return
		}
		c.svcDsc.Store(svc)
	}
}

func (c *jsonProtobufCodec) Marshal(ctx context.Context, msg remote.Message, out remote.ByteBuffer) error {
	method := msg.RPCInfo().Invocation().MethodName()
	if method == "" {
		return perrors.NewProtocolErrorWithMsg("empty methodName in protobuf Marshal")
	}

	if msg.MessageType() == remote.Exception {
		return c.codec.Marshal(ctx, msg, out)
	}

	message, ok := msg.Data().(proto.Message)
	if !ok {
		return perrors.NewProtocolErrorWithMsg("could not convert msg.Data() to proto.Message")
	}

	jsonBytes, err := protojson.MarshalOptions{}.Marshal(message)
	if err != nil {
		return err
	}
	out.Write(jsonBytes)
	return nil
}

func (c *jsonProtobufCodec) Unmarshal(ctx context.Context, msg remote.Message, in remote.ByteBuffer) error {
	if err := codec.NewDataIfNeeded(serviceinfo.GenericMethod, msg); err != nil {
		return err
	}

	unm := &protojson.UnmarshalOptions{}
	message, ok := msg.Data().(proto.Message)
	if !ok {
		return perrors.NewProtocolErrorWithMsg("could not convert msg.Data() to proto.Message")
	}

	data, err := in.Bytes()
	if err != nil {
		return err
	}

	err = unm.Unmarshal(data, message)
	if err != nil {
		return err
	}
	return nil
}

func (c *jsonProtobufCodec) getMethod(req interface{}, method string) (*Method, error) {
	fnSvc, err := c.svcDsc.Load().(*descriptor.ServiceDescriptor).LookupFunctionByMethod(method)
	if err != nil {
		return nil, err
	}
	return &Method{method, fnSvc.Oneway}, nil
}

func (c *jsonProtobufCodec) Name() string {
	return "JSONProtobuf"
}

func (c *jsonProtobufCodec) Close() error {
	return c.provider.Close()
}

// package generic

// import (
// 	"context"
// 	"errors"

// 	"github.com/cloudwego/kitex/pkg/generic"
// 	"github.com/cloudwego/kitex/pkg/remote"
// 	"github.com/golang/protobuf/jsonpb"
// 	"github.com/golang/protobuf/proto"
// )

// type jsonProtobufCodec struct {
// 	marshaler   jsonpb.Marshaler
// 	unmarshaler jsonpb.Unmarshaler
// }

// // NewJSONProtobufCodec creates a new codec for JSON-Protobuf conversion.
// func NewJSONProtobufCodec() generic.Codec {
// 	return &jsonProtobufCodec{
// 		marshaler:   jsonpb.Marshaler{},
// 		unmarshaler: jsonpb.Unmarshaler{},
// 	}
// }

// // Encode encodes a message into JSON format.
// func (c *jsonProtobufCodec) Marshal(ctx context.Context, message remote.Message, request interface{}) (data []byte, err error) {
// 	if pb, ok := request.(proto.Message); ok {
// 		return []byte(c.marshaler.MarshalToString(pb)), nil
// 	}
// 	return nil, errors.New("Invalid request for protobuf encoding")
// }

// // Decode decodes a message from JSON format.
// func (c *jsonProtobufCodec) Unmarshal(ctx context.Context, message remote.Message, data []byte, response interface{}) (err error) {
// 	if pb, ok := response.(proto.Message); ok {
// 		return c.unmarshaler.UnmarshalString(string(data), pb)
// 	}
// 	return errors.New("Invalid response for protobuf decoding")
// }

// // Name provides the name of the codec.
// func (c *jsonProtobufCodec) Name(ctx context.Context, message remote.Message) (name string, err error) {
// 	return "jsonprotobuf", nil
// }
