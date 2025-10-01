package proxy

import (
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"net"
	"time"
	_ "unsafe"

	"github.com/gameparrot/netherconnect/auth"

	"github.com/gameparrot/netherconnect/proxy/newlogin"
	"github.com/go-jose/go-jose/v4"
	"github.com/go-jose/go-jose/v4/jwt"
	"github.com/sandertv/gophertunnel/minecraft/protocol/login"

	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
)

func (conn *ProxyConn) Login(clientData login.ClientData, authSession *auth.Session, protocol int32) (err error) {
	if err := conn.WritePacket(&packet.RequestNetworkSettings{ClientProtocol: protocol}); err != nil {
		return fmt.Errorf("send request network settings: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	key, _ := ecdsa.GenerateKey(elliptic.P384(), rand.Reader)

	xsts, err := authSession.LegacyMultiplayerXBL(context.Background())
	if err != nil {
		return err
	}

	var chainData string

	chainData, err = authChain(ctx, xsts, key)
	if err != nil {
		return &net.OpError{Op: "dial", Net: "minecraft", Err: err}
	}
	identityData, err := readChainIdentityData([]byte(chainData))
	if err != nil {
		return &net.OpError{Op: "dial", Net: "minecraft", Err: err}
	}

	conn.clientData = clientData
	conn.privateKey = key

	newTok, err := authSession.MultiplayerToken(context.Background(), key)
	if err != nil {
		return err
	}

	request := newlogin.Encode(chainData, newTok.SignedToken, conn.clientData, key, false)
	identityData, _, _, _ = newlogin.Parse(request)
	// If we got the identity data from Minecraft auth, we need to make sure we set it in the Conn too, as
	// we are not aware of the identity data ourselves yet.
	conn.identityData = identityData
	conn.loginPk = &packet.Login{ConnectionRequest: request, ClientProtocol: protocol}

	done := make(chan struct{})
	go func() {
		select {
		case <-ctx.Done():
			conn.conn.Close()
			err = ctx.Err()
		case <-done:
			return
		}
	}()
	err = conn.ReadLoop()
	done <- struct{}{}
	return err
}

// authChain requests the Minecraft auth JWT chain using the credentials passed. If successful, an encoded
// chain ready to be put in a login request is returned.
func authChain(ctx context.Context, xsts *auth.XBLToken, key *ecdsa.PrivateKey) (string, error) {
	// Obtain the raw chain data using the
	chain, err := auth.RequestMinecraftChain(ctx, xsts, key)
	if err != nil {
		return "", fmt.Errorf("request Minecraft auth chain: %w", err)
	}
	return chain, nil
}

// readChainIdentityData reads a login.IdentityData from the Mojang chain
// obtained through authentication.
func readChainIdentityData(chainData []byte) (login.IdentityData, error) {
	chain := struct{ Chain []string }{}
	if err := json.Unmarshal(chainData, &chain); err != nil {
		return login.IdentityData{}, fmt.Errorf("read chain: read json: %w", err)
	}
	data := chain.Chain[1]
	claims := struct {
		ExtraData login.IdentityData `json:"extraData"`
	}{}
	tok, err := jwt.ParseSigned(data, []jose.SignatureAlgorithm{jose.ES384})
	if err != nil {
		return login.IdentityData{}, fmt.Errorf("read chain: parse jwt: %w", err)
	}
	if err := tok.UnsafeClaimsWithoutVerification(&claims); err != nil {
		return login.IdentityData{}, fmt.Errorf("read chain: read claims: %w", err)
	}
	if claims.ExtraData.Identity == "" {
		return login.IdentityData{}, fmt.Errorf("read chain: no extra data found")
	}
	return claims.ExtraData, nil
}
