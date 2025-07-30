package handlers

import (
	"errors"
	"net/http"

	"github.com/coreos/go-oidc"
	"github.com/valyala/fasthttp"
	bolt "go.etcd.io/bbolt"
	"golang.org/x/oauth2"
	"k8s.io/apimachinery/pkg/util/json"
)

func (sc *SailorCore) AuthOIDCHandler(ctx *fasthttp.RequestCtx) {
	enc := json.NewEncoder(ctx)

	ss, err := getSailorSetting(sc.dbconns[BUCKET_ADMIN])
	if err != nil {
		ctx.SetStatusCode(http.StatusInternalServerError)
		enc.Encode(ResponseMessage{Message: err.Error()})
		return
	}

	provider, err := oidc.NewProvider(ctx, ss.OIDC.IssuerURL)
	if err != nil {
		ctx.SetStatusCode(http.StatusInternalServerError)
		enc.Encode(ResponseMessage{Message: err.Error()})
		return
	}

	oauth2Config := oauth2.Config{
		ClientID:     ss.OIDC.ClientID,
		ClientSecret: ss.OIDC.ClientSecret,
		Endpoint:     provider.Endpoint(),
		RedirectURL:  ss.OIDC.RedirectURL,
		Scopes:       []string{oidc.ScopeOpenID, "profile", "email"},
	}

	// TODO :: think about the state which should be passed as part of OIDC request
	ctx.Redirect(oauth2Config.AuthCodeURL("state-ash"), http.StatusFound)
}

func getSailorSetting(adminDB *bolt.DB) (*SailorSetting, error) {
	var ss SailorSetting
	err := adminDB.View(func(tx *bolt.Tx) error {
		settingBytes := tx.Bucket([]byte(BUCKET_SETTING)).Get([]byte(KEY_SETTING))
		if settingBytes == nil {
			return errors.New("sailor settings not found.")
		}

		if err := json.Unmarshal(settingBytes, &ss); err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return &ss, nil
}
