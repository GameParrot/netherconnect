package messaging

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/coder/websocket"
	"github.com/go-jose/go-jose/v4"
	"github.com/go-jose/go-jose/v4/jwt"
	"github.com/google/uuid"
	"github.com/sandertv/gophertunnel/minecraft/protocol"
	"github.com/sandertv/gophertunnel/minecraft/service"
)

type Dialer struct {
	Options *websocket.DialOptions
	Log     *slog.Logger
}

func (d Dialer) DialContext(ctx context.Context, mct *service.Token) (*Conn, error) {
	if d.Log == nil {
		d.Log = slog.Default()
	}

	if d.Options == nil {
		d.Options = &websocket.DialOptions{}
	}
	if d.Options.HTTPClient == nil {
		d.Options.HTTPClient = &http.Client{}
	}
	if d.Options.HTTPHeader == nil {
		d.Options.HTTPHeader = make(http.Header) // TODO(lactyy): Move to *franchise.Transport
	}

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

	d.Options.HTTPHeader.Set("Authorization", mct.AuthorizationHeader)
	d.Options.HTTPHeader.Set("User-Agent", "libHttpClient/1.0.0.0")
	// TODO: session-id and request-id

	conn, _, err := websocket.Dial(context.Background(), u.JoinPath("/ws/v1.0/messaging/connect").String(), d.Options)
	if err != nil {
		return nil, fmt.Errorf("error connecting to signaling service: %w", err)
	}
	/*t.Cleanup(func() {
		if err := conn.Close(websocket.StatusNormalClosure, ""); err != nil {
			t.Fatalf("error closing websocket connection: %s", err)
		}
	})*/
	messagingID, err := claimMessagingID(mct)
	if err != nil {
		return nil, fmt.Errorf("obtain messaging id: %w", err)
	}

	return newConn(conn, messagingID, messagingID.String(), context.Background(), d.Log), nil

}

func claimMessagingID(mct *service.Token) (uuid.UUID, error) {
	token, err := jwt.ParseSigned(strings.TrimPrefix(mct.AuthorizationHeader, "MCToken "), []jose.SignatureAlgorithm{
		jose.RS256,
	})
	if err != nil {
		return uuid.UUID{}, fmt.Errorf("error parsing service token: %w", err)
	}
	var claims struct {
		jwt.Claims
		PlayerMessagingID uuid.UUID `json:"pmid"`
	}
	// actually the key is under the OpenID configuration, but we won't verify it for now
	if err := token.UnsafeClaimsWithoutVerification(&claims); err != nil {
		return uuid.UUID{}, fmt.Errorf("error extracting JWT claims: %w", err)
	}

	if err := claims.Validate(jwt.Expected{Time: time.Now()}); err != nil {
		return uuid.UUID{}, fmt.Errorf("error validating claims: %w", err)
	}
	return claims.PlayerMessagingID, nil
}
