package handlers

import (
	"encoding/json"
	"net/http"
	"slices"

	"github.com/valyala/fasthttp"
	bolt "go.etcd.io/bbolt"
)

type OIDCSetting struct {
	ClientID     string   `json:"client_id"`
	ClientSecret string   `json:"client_secret"`
	IssuerURL    string   `json:"issuer_url"`
	Scopes       []string `json:"scopes"`
	RedirectURL  string   `json:"redirect_url"`
}

type SailorSetting struct {
	OIDC        *OIDCSetting `json:"oidc"`
	TokenKey    string       `json:"token_key"`
	OIDC     *OIDCSetting `json:"oidc"`
	TokenKey string       `json:"token_key"`
}

func (sc *SailorCore) SailorSettingHandler(ctx *fasthttp.RequestCtx) {
	enc := json.NewEncoder(ctx)

	var ss SailorSetting
	if err := json.Unmarshal(ctx.Request.Body(), &ss); err != nil {
		ctx.SetStatusCode(http.StatusBadRequest)
		enc.Encode(ResponseMessage{Message: err.Error()})
		return
	}

	if ss.OIDC == nil {
		ctx.SetStatusCode(http.StatusNotFound)
		enc.Encode(ResponseMessage{Message: "oidc sailor settings not present"})
	}

	if ss.OIDC.ClientID == "" || ss.OIDC.ClientSecret == "" ||
		ss.OIDC.IssuerURL == "" || ss.OIDC.RedirectURL == "" || len(ss.OIDC.Scopes) == 0 {
		ctx.SetStatusCode(http.StatusBadRequest)
		enc.Encode(ResponseMessage{Message: "oidc validations failed, some required fields are empty"})
		return
	}

	if !slices.Contains(ss.OIDC.Scopes, "email") {
		ctx.SetStatusCode(http.StatusBadRequest)
		enc.Encode(ResponseMessage{Message: "oidc:scopes must contain email"})
		return
	}

	err := sc.dbconns[BUCKET_ADMIN].Update(func(tx *bolt.Tx) error {
		b, err := json.Marshal(&ss)
		if err != nil {
			return err
		}
		return tx.Bucket([]byte(BUCKET_SETTING)).Put([]byte(KEY_SETTING), b)
	})

	if err != nil {
		ctx.SetStatusCode(http.StatusInternalServerError)
		enc.Encode(ResponseMessage{Message: err.Error()})
		return
	}

	// we update the sailor settings inside our own core memory lookup
	sc.setting = &ss

	enc.Encode(ResponseMessage{Message: "ok"})
}

func (sc *SailorCore) GetSailorSettingHandler(ctx *fasthttp.RequestCtx) {
	enc := json.NewEncoder(ctx)
	ctx.Response.Header.Set("Content-Type", "application/json")
	enc.Encode(sc.setting)
}
