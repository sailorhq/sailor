package handlers

import (
	"encoding/json"
	"fmt"

	"github.com/golang-jwt/jwt/v5"
	v1 "github.com/sailorhq/sailor/pkg/core/v1"
	"github.com/valyala/fasthttp"
)

func (sc *SailorCore) GetKeyPairHandler(ctx *fasthttp.RequestCtx) {
	enc := json.NewEncoder(ctx)
	enc.SetEscapeHTML(false)

	params := extractSailorParams(ctx)

	if params.Ns == "" || params.App == "" {
		ctx.SetStatusCode(fasthttp.StatusBadRequest)
		enc.Encode(ResponseMessage{Message: "namespace or app should not be empty"})
		return
	}

	projectKey := fmt.Sprintf("%s_%s", params.Ns, params.App)
	if _, ok := sc.dbconns[projectKey]; !ok {
		ctx.SetStatusCode(fasthttp.StatusInternalServerError)
		enc.Encode(ResponseMessage{Message: "no such project in sailor"})
		return
	}

	// we identified that a token is passed, this is passed only when a client like CLI
	// is used to fetch a resource
	claims := ctx.UserValue("__sailor_claims").(jwt.MapClaims)
	if ak, ok := claims["access_key"].(string); !ok || ak == "" {
		ctx.SetStatusCode(fasthttp.StatusBadRequest)
		enc.Encode(ResponseMessage{"invalid token #5"})
		return
	}

	if sk, ok := claims["secret_key"].(string); !ok || sk == "" {
		ctx.SetStatusCode(fasthttp.StatusBadRequest)
		enc.Encode(ResponseMessage{"invalid token #6"})
		return
	}

	enc.Encode(v1.KeyPair{
		AccessKey: claims["access_key"].(string),
		SecretKey: claims["secret_key"].(string),
	})
}
