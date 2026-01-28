package signaling

import (
	"context"
	"fmt"
	"log/slog"
	"math/rand"
	"net/http"
	"net/url"
	"strconv"

	"github.com/df-mc/go-nethernet"
	"github.com/sandertv/gophertunnel/minecraft/protocol"
	"github.com/sandertv/gophertunnel/minecraft/service"
	"github.com/coder/websocket"
)

type Dialer struct {
	Options   *websocket.DialOptions
	NetworkID string
	Log       *slog.Logger
}

func (d Dialer) DialContext(ctx context.Context, tok *service.Token) (*Conn, error) {
	if d.Options == nil {
		d.Options = &websocket.DialOptions{}
	}
	if d.Options.HTTPClient == nil {
		d.Options.HTTPClient = &http.Client{}
	}
	if d.Options.HTTPHeader == nil {
		d.Options.HTTPHeader = make(http.Header) // TODO(lactyy): Move to *franchise.Transport
	}
	if d.NetworkID == "" {
		d.NetworkID = strconv.FormatUint(rand.Uint64(), 10)
	}
	if d.Log == nil {
		d.Log = slog.Default()
	}

	d.Options.HTTPHeader.Set("Authorization", tok.AuthorizationHeader)

	var env Environment

	discovery, err := service.Discover(ctx, service.ApplicationTypeMinecraftPE, protocol.CurrentVersion)
	if err != nil {
		return nil, fmt.Errorf("discover: %w", err)
	}

	if err := discovery.Environment(&env); err != nil {
		return nil, fmt.Errorf("decode environment: %w", err)
	}

	u, err := url.Parse(env.ServiceURI)
	if err != nil {
		return nil, fmt.Errorf("parse service URI: %w", err)
	}

	c, _, err := websocket.Dial(ctx, u.JoinPath("/ws/v1.0/signaling/", d.NetworkID).String(), d.Options)
	if err != nil {
		return nil, err
	}

	conn := &Conn{
		conn:    c,
		d:       d,
		signals: make(chan *nethernet.Signal),
		ready:   make(chan struct{}),
	}
	var cancel context.CancelCauseFunc
	conn.ctx, cancel = context.WithCancelCause(context.Background())

	go conn.read(cancel)
	go conn.ping()

	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case <-conn.ready:
		return conn, nil
	}
}
