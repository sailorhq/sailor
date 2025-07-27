package handlers

import (
	"encoding/json"
	"net/http"

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
	OIDC *OIDCSetting `json:"oidc"`
}

func (sc *SailorCore) SailorSettingHandler(ctx *fasthttp.RequestCtx) {
	enc := json.NewEncoder(ctx)

	var ss SailorSetting
	if err := json.Unmarshal(ctx.Response.Body(), &ss); err != nil {
		ctx.SetStatusCode(http.StatusBadRequest)
		enc.Encode(ResponseMessage{Message: err.Error()})
		return
	}

	if ss.OIDC == nil || ss.OIDC.ClientID == "" ||
		ss.OIDC.IssuerURL == "" || ss.OIDC.RedirectURL == "" || len(ss.OIDC.Scopes) == 0 {
		ctx.SetStatusCode(http.StatusBadRequest)
		enc.Encode(ResponseMessage{Message: "oidc validations failed, some required fields are empty"})
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

	enc.Encode(ResponseMessage{Message: "ok"})
}
