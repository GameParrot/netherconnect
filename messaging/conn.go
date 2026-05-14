package messaging

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/coder/websocket"
	"github.com/creachadair/jrpc2"
	"github.com/df-mc/go-nethernet"
	"github.com/google/uuid"
)

type Conn struct {
	log *slog.Logger

	ctx    context.Context
	cancel context.CancelCauseFunc

	conn           *websocket.Conn
	client         *jrpc2.Client
	messagingID    uuid.UUID
	localNetworkID string

	signals chan *nethernet.Signal

	credentials     *nethernet.Credentials
	credentialsTime time.Time
	credentialsMu   sync.Mutex

	once sync.Once

	pendingStore   map[uint64]*pendingStore
	pendingStoreMu sync.Mutex
}

type pendingStore struct {
	c                         *Conn
	pendingSignals            []*nethernet.Signal
	hasReceivedConnectRequest bool
	lastActivity              time.Time
}

func (p *pendingStore) trySend(s *nethernet.Signal) {
	p.lastActivity = time.Now()
	if s.Type == nethernet.SignalTypeOffer {
		p.c.signals <- s
		p.hasReceivedConnectRequest = true
		for _, s := range p.pendingSignals {
			p.c.signals <- s
		}
		p.pendingSignals = []*nethernet.Signal{}
		return
	}

	if p.hasReceivedConnectRequest || s.Type != nethernet.SignalTypeCandidate {
		p.c.signals <- s
		return
	}

	p.pendingSignals = append(p.pendingSignals, s)
}

func newConn(conn *websocket.Conn, messagingID uuid.UUID, networkID string, ctx context.Context, log *slog.Logger) *Conn {

	connCtx, cancel := context.WithCancelCause(ctx)
	j := &Conn{
		log:            log,
		signals:        make(chan *nethernet.Signal),
		ctx:            connCtx,
		cancel:         cancel,
		conn:           conn,
		messagingID:    messagingID,
		localNetworkID: networkID,
		pendingStore:   make(map[uint64]*pendingStore),
	}
	j.client = jrpc2.NewClient(&websocketChannel{conn}, &jrpc2.ClientOptions{
		OnCallback: j.handleCallback,
		OnStop: func(cli *jrpc2.Client, err error) {
			fmt.Println("Stopped: " + err.Error())
			j.cancel(err)
		},
	})
	go j.pendingStoreTicker()
	return j
}

func (c *Conn) pendingStoreTicker() {
	ticker := time.NewTicker(time.Second * 15)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			c.client.Call(context.Background(), "System_Ping_v1_0", map[string]any{})

			c.pendingStoreMu.Lock()
			for id, s := range c.pendingStore {
				if time.Since(s.lastActivity) > 30*time.Second {
					delete(s.c.pendingStore, id)
				}
			}
			c.pendingStoreMu.Unlock()
		case <-c.ctx.Done():
			return
		}
	}
}

func (j *Conn) handleCallback(ctx context.Context, req *jrpc2.Request) (result any, err error) {
	switch req.Method() {
	case "Signaling_ReceiveMessage_v1_0":
		defer func() {
			if err != nil {
				j.log.Error(err.Error())
				//j.Close()
				return
			}
		}()
		var batch []struct {
			From  uuid.UUID
			Inner string `json:"Message"`
			ID    string `json:"Id"`
		}
		if err := req.UnmarshalParams(&batch); err != nil {
			j.log.Error(err.Error())
			return nil, nil
			//return nil, fmt.Errorf("handle %q: decode parameters: %w", req.Method(), err)
		}
		for _, msg := range batch {
			var inner *jrpc2.ParsedRequest
			if err := json.Unmarshal([]byte(msg.Inner), &inner); err != nil {
				j.log.Error(err.Error())
				return nil, nil
				//return nil, fmt.Errorf("handle %q: decode inner message: %w", req.Method(), err)
			}
			if inner == nil {
				j.log.Error(fmt.Sprintf("handle %q: invalid batch message in params", req.Method()))
				return nil, nil
				//return nil, fmt.Errorf("handle %q: invalid batch message in params", req.Method())
			}

			switch inner.Method {
			case "Signaling_WebRtc_v1_0":
				var params struct {
					NetherNetID string `json:"netherNetId"` // ignored, just using their Messaging ID
					Data        string `json:"message"`
				}
				if err := json.Unmarshal(inner.Params, &params); err != nil {
					j.log.Error(err.Error())
					return nil, nil
					//return nil, fmt.Errorf("handle %q: decode parameters in inner message: %w", req.Method(), err)
				}
				if params.NetherNetID == "" || params.Data == "" {
					j.log.Error(fmt.Sprintf("handle %q: invalid inner message", req.Method()))
					return nil, nil
					//return nil, fmt.Errorf("handle %q: invalid inner message", req.Method())
				}

				signal := &nethernet.Signal{NetworkID: msg.From.String()}
				if err := signal.UnmarshalText([]byte(params.Data)); err != nil {
					j.log.Error(err.Error())
					return nil, nil
					//return nil, fmt.Errorf("handle %q: decode inner message data to signal: %w", req.Method(), err)
				}

				if signal.Type == nethernet.SignalTypeCandidate || signal.Type == nethernet.SignalTypeOffer {
					id := signal.ConnectionID
					j.pendingStoreMu.Lock()
					if _, ok := j.pendingStore[id]; !ok {
						j.pendingStore[id] = &pendingStore{c: j, lastActivity: time.Now()}
					}
					store := j.pendingStore[id]
					j.pendingStoreMu.Unlock()
					store.trySend(signal)
				} else {
					j.signals <- signal
				}

				b, _ := json.Marshal(map[string]any{
					"jsonrpc": "2.0",
					"method":  "Signaling_DeliveryNotification_V1_0",
					"params": map[string]any{
						"netherNetId": j.localNetworkID,
					},
				})
				resp, err := j.client.Call(ctx, "Signaling_SendClientMessage_v1_0", map[string]any{
					"toPlayerId": msg.From,
					"messageId":  msg.ID,
					"message":    string(b),
				})
				if err != nil {
					j.log.Error(err.Error())
					return nil, nil
					//return nil, fmt.Errorf("call Signaling_SendClientMessage_v1_0: %w", err)
				}
				if resp.Error() != nil {
					j.log.Error(resp.Error().Error())
					return nil, nil
					//return nil, resp.Error()
				}
				return nil, nil
			case "Signaling_DeliveryNotification_V1_0":
				continue
			default:
				//return nil, fmt.Errorf("handle %q: invalid inner message method: %q", req.Method(), inner.Method)
			}
		}
		return nil, nil
	case "System_Pong_v1_0":
		return nil, nil
	default:
		return nil, nil

		//return nil, fmt.Errorf("unknown JSONRPC method: %q", req.Method())
	}
}

func (j *Conn) Signal(ctx context.Context, signal *nethernet.Signal) error {
	// j.t.Logf("Signal(%s)", signal)

	// This is half-encoded JSONRPC 2.0 Message, but it isn't exported in the jrpc2 package.
	b, err := json.Marshal(map[string]any{
		"jsonrpc": "2.0",
		"method":  "Signaling_WebRtc_v1_0",
		"params": map[string]any{
			"netherNetId": j.localNetworkID,
			"message":     signal.String(),
		},
	})
	messagingID := uuid.MustParse(signal.NetworkID)
	resp, err := j.client.Call(context.Background(), "Signaling_SendClientMessage_v1_0", map[string]any{
		"toPlayerId": messagingID,
		"messageId":  uuid.New(), //< A unique ID associated to each message sent by the client.
		"message":    string(b),
	})
	if err != nil {
		return err
	}
	if resp.Error() != nil {
		return resp.Error()
	}
	return nil
}

func (j *Conn) Close() (err error) {
	j.once.Do(func() {
		err = j.conn.Close(websocket.StatusGoingAway, "")
		err = errors.Join(err, j.client.Close())
	})
	return err
}

func (c *Conn) ReadSignal() (*nethernet.Signal, error) {
	select {
	case s := <-c.signals:
		return s, nil
	case <-c.ctx.Done():
		return nil, context.Cause(c.ctx)
	}
}

func (c *Conn) Context() context.Context {
	return c.ctx
}

func (c *Conn) Notify(signals chan<- *nethernet.Signal) (stop func()) {
	go func() {
		for {
			sig, err := c.ReadSignal()
			if err != nil {
				close(signals)
				return
			}
			signals <- sig
		}
	}()
	return func() {
		c.Close()
	}
}

func (j *Conn) Credentials(ctx context.Context) (*nethernet.Credentials, error) {
	// j.t.Logf("Credentials(%#v)", ctx)

	j.credentialsMu.Lock()
	defer j.credentialsMu.Unlock()

	if j.credentials != nil {
		exp := j.credentialsTime.Add(time.Second * time.Duration(j.credentials.ExpirationInSeconds))
		if time.Now().Before(exp) {
			return j.credentials, nil
		}
	}

	var credentials *nethernet.Credentials
	if err := j.client.CallResult(ctx, "Signaling_TurnAuth_v1_0", map[string]any{}, &credentials); err != nil {
		return nil, fmt.Errorf("call Signaling_TurnAuth_v1_0: %w", err)
	}
	if credentials == nil || credentials.ExpirationInSeconds == 0 {
		return nil, errors.New("call Signaling_TurnAuth_v1_0: invalid result")
	}
	j.credentials, j.credentialsTime = credentials, time.Now()
	return j.credentials, nil
}

func (j *Conn) NetworkID() string {
	return j.localNetworkID
}

func (j *Conn) MessagingID() uuid.UUID {
	return j.messagingID
}

func (j *Conn) PongData([]byte) {}

type websocketChannel struct{ *websocket.Conn }

func (ch *websocketChannel) Send(b []byte) error {
	return ch.Write(context.Background(), websocket.MessageText, b)
}

func (ch *websocketChannel) Recv() ([]byte, error) {
	_, msg, err := ch.Read(context.Background())
	return msg, err
}

func (ch *websocketChannel) Close() error {
	return ch.Conn.Close(websocket.StatusGoingAway, "")
}
