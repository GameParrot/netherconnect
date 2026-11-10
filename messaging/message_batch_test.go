package messaging

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"testing"

	"github.com/creachadair/jrpc2"
	"github.com/creachadair/jrpc2/handler"
	"github.com/creachadair/jrpc2/server"
	"github.com/df-mc/go-nethernet"
	"github.com/google/uuid"
)

func TestMessageBatchForms(t *testing.T) {
	for _, input := range []string{`{"From":"00000000-0000-0000-0000-000000000001","Message":"x","Id":"y"}`, `[{"From":"00000000-0000-0000-0000-000000000001","Message":"x","Id":"y"}]`} {
		var batch messageBatch
		if err := json.Unmarshal([]byte(input), &batch); err != nil || len(batch) != 1 || batch[0].Inner != "x" {
			t.Fatalf("decode: %v %#v", err, batch)
		}
	}
}

func TestCallbackProcessesEverySignalInBatch(t *testing.T) {
	var acknowledged []string
	rpc := server.NewLocal(handler.Map{
		"Signaling_SendClientMessage_v1_0": func(_ context.Context, req *jrpc2.Request) (any, error) {
			var params struct {
				MessageID string `json:"messageId"`
			}
			if err := req.UnmarshalParams(&params); err != nil {
				return nil, err
			}
			acknowledged = append(acknowledged, params.MessageID)
			return struct{}{}, nil
		},
	}, nil)
	defer rpc.Close()
	conn := &Conn{client: rpc.Client, log: slog.New(slog.NewTextHandler(io.Discard, nil)), signals: make(chan *nethernet.Signal, 2)}
	from := uuid.MustParse("00000000-0000-0000-0000-000000000001")
	inner := `{"jsonrpc":"2.0","method":"Signaling_WebRtc_v1_0","params":{"netherNetId":"1","message":"CONNECTRESPONSE 1 answer"}}`
	params, err := json.Marshal(messageBatch{{From: from, Inner: inner, ID: "first"}, {From: from, Inner: inner, ID: "second"}})
	if err != nil {
		t.Fatal(err)
	}
	req := (&jrpc2.ParsedRequest{Method: "Signaling_ReceiveMessage_v1_0", Params: params}).ToRequest()
	if _, err := conn.handleCallback(context.Background(), req); err != nil {
		t.Fatal(err)
	}
	if len(conn.signals) != 2 || len(acknowledged) != 2 || acknowledged[0] != "first" || acknowledged[1] != "second" {
		t.Fatalf("signals = %d, acknowledgements = %v", len(conn.signals), acknowledged)
	}
}

func TestCallbackSkipsMalformedMessageAndProcessesFollowingSignal(t *testing.T) {
	var acknowledged []string
	rpc := server.NewLocal(handler.Map{
		"Signaling_SendClientMessage_v1_0": func(_ context.Context, req *jrpc2.Request) (any, error) {
			var params struct {
				MessageID string `json:"messageId"`
			}
			if err := req.UnmarshalParams(&params); err != nil {
				return nil, err
			}
			acknowledged = append(acknowledged, params.MessageID)
			return struct{}{}, nil
		},
	}, nil)
	defer rpc.Close()
	conn := &Conn{client: rpc.Client, log: slog.New(slog.NewTextHandler(io.Discard, nil)), signals: make(chan *nethernet.Signal, 1)}
	from := uuid.MustParse("00000000-0000-0000-0000-000000000001")
	inner := `{"jsonrpc":"2.0","method":"Signaling_WebRtc_v1_0","params":{"netherNetId":"1","message":"CONNECTRESPONSE 1 answer"}}`
	params, err := json.Marshal(messageBatch{{From: from, Inner: "{", ID: "bad"}, {From: from, Inner: inner, ID: "good"}})
	if err != nil {
		t.Fatal(err)
	}
	req := (&jrpc2.ParsedRequest{Method: "Signaling_ReceiveMessage_v1_0", Params: params}).ToRequest()
	if _, err := conn.handleCallback(context.Background(), req); err != nil {
		t.Fatal(err)
	}
	if len(conn.signals) != 1 || len(acknowledged) != 1 || acknowledged[0] != "good" {
		t.Fatalf("signals = %d, acknowledgements = %v", len(conn.signals), acknowledged)
	}
}

func TestCallbackProcessesFollowingSignalAfterAcknowledgementFailure(t *testing.T) {
	var acknowledged []string
	rpc := server.NewLocal(handler.Map{
		"Signaling_SendClientMessage_v1_0": func(_ context.Context, req *jrpc2.Request) (any, error) {
			var params struct {
				MessageID string `json:"messageId"`
			}
			if err := req.UnmarshalParams(&params); err != nil {
				return nil, err
			}
			acknowledged = append(acknowledged, params.MessageID)
			if params.MessageID == "first" {
				return nil, errors.New("delivery failed")
			}
			return struct{}{}, nil
		},
	}, nil)
	defer rpc.Close()
	conn := &Conn{client: rpc.Client, log: slog.New(slog.NewTextHandler(io.Discard, nil)), signals: make(chan *nethernet.Signal, 2)}
	from := uuid.MustParse("00000000-0000-0000-0000-000000000001")
	inner := `{"jsonrpc":"2.0","method":"Signaling_WebRtc_v1_0","params":{"netherNetId":"1","message":"CONNECTRESPONSE 1 answer"}}`
	params, err := json.Marshal(messageBatch{{From: from, Inner: inner, ID: "first"}, {From: from, Inner: inner, ID: "second"}})
	if err != nil {
		t.Fatal(err)
	}
	req := (&jrpc2.ParsedRequest{Method: "Signaling_ReceiveMessage_v1_0", Params: params}).ToRequest()
	if _, err := conn.handleCallback(context.Background(), req); err != nil {
		t.Fatal(err)
	}
	if len(conn.signals) != 2 || len(acknowledged) != 2 {
		t.Fatalf("signals = %d, acknowledgements = %v", len(conn.signals), acknowledged)
	}
}

func TestDeliveryNotificationIdentifiesReceivedMessage(t *testing.T) {
	const messageID = "00000000-0000-0000-0000-000000000001"
	data, err := deliveryNotification(messageID)
	if err != nil {
		t.Fatal(err)
	}
	var message struct {
		Method string `json:"method"`
		Params struct {
			MessageID string `json:"messageId"`
		} `json:"params"`
	}
	if err := json.Unmarshal(data, &message); err != nil {
		t.Fatal(err)
	}
	if message.Method != "Signaling_DeliveryNotification_V1_0" || message.Params.MessageID != messageID {
		t.Fatalf("unexpected delivery notification: %+v", message)
	}
}
