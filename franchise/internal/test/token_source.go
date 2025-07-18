package test

import (
	"encoding/json"
	"fmt"
	"golang.org/x/oauth2"
	"os"
	"testing"
)

func TokenSource(t *testing.T, path string, src oauth2.TokenSource, hooks ...RefreshTokenFunc) *oauth2.Token {
	tok, err := readTokenSource(path, src)
	if err != nil {
		t.Fatalf("error reading token: %s", err)
	}
	for _, h := range hooks {
		tok, err = h(tok)
		if err != nil {
			t.Fatalf("error refreshing token: %s", err)
		}
	}
	return tok
}

type RefreshTokenFunc func(old *oauth2.Token) (new *oauth2.Token, err error)

func readTokenSource(path string, src oauth2.TokenSource) (t *oauth2.Token, err error) {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		t, err = src.Token()
		if err != nil {
			return nil, fmt.Errorf("obtain token: %w", err)
		}
		f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, os.ModePerm)
		if err != nil {
			return nil, err
		}
		defer f.Close()
		if err := json.NewEncoder(f).Encode(t); err != nil {
			return nil, fmt.Errorf("encode: %w", err)
		}
		return t, nil
	} else if err != nil {
		return nil, fmt.Errorf("stat: %w", err)
	}
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	if err := json.NewDecoder(f).Decode(&t); err != nil {
		return nil, fmt.Errorf("decode: %w", err)
	}
	return t, nil
}
