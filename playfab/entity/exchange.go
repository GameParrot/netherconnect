package entity

import (
	"github.com/gameparrot/netherconnect/playfab/internal"
	"github.com/gameparrot/netherconnect/playfab/title"
)

func (tok *Token) Exchange(t title.Title, id string) (_ *Token, err error) {
	r := exchange{
		Entity: Key{
			Type: TypeMasterPlayerAccount,
			ID:   id,
		},
	}

	return internal.Post[*Token](t, "/Authentication/GetEntityToken", r, tok.SetAuthHeader)
}

type exchange struct {
	CustomTags map[string]any `json:"CustomTags,omitempty"`
	Entity     Key            `json:"Entity,omitempty"`
}
