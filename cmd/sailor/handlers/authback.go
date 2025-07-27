package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/coreos/go-oidc"
	"github.com/valyala/fasthttp"
	"go.uber.org/zap"
	"golang.org/x/oauth2"
)

func (sc *SailorCore) AuthCallbackHandler(ctx *fasthttp.RequestCtx) {
	enc := json.NewEncoder(ctx)

	ss, err := getSailorSetting(sc.dbconns[BUCKET_ADMIN])
	if err != nil {
		ctx.SetStatusCode(http.StatusInternalServerError)
		enc.Encode(ResponseMessage{Message: err.Error()})
		return
	}

	oidcConfig := &oidc.Config{
		ClientID: ss.OIDC.ClientID,
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

	verifier := provider.Verifier(oidcConfig)

	if errMsg := ctx.QueryArgs().Peek("error"); string(errMsg) != "" {
		ctx.SetStatusCode(http.StatusBadRequest)
		enc.Encode(ResponseMessage{Message: string(errMsg)})
		return
	}

	code := ctx.QueryArgs().Peek("code")
	token, err := oauth2Config.Exchange(ctx, string(code))
	if err != nil {
		ctx.SetStatusCode(http.StatusInternalServerError)
		enc.Encode(ResponseMessage{Message: err.Error()})
		return
	}

	sc.Log.Debug("received token from callback", zap.Any("token", *token))

	rawIDToken, ok := token.Extra("id_token").(string)
	if !ok {
		ctx.SetStatusCode(http.StatusInternalServerError)
		enc.Encode(ResponseMessage{Message: "missing id_token"})
		return
	}

	idToken, err := verifier.Verify(ctx, rawIDToken)
	if err != nil {
		ctx.SetStatusCode(http.StatusInternalServerError)
		enc.Encode(ResponseMessage{Message: "token verification failed"})
		return
	}

	var claims map[string]any
	if err := idToken.Claims(&claims); err != nil {
		ctx.SetStatusCode(http.StatusInternalServerError)
		enc.Encode(ResponseMessage{Message: "cannot parse claims from token"})
		return
	}

	sc.Log.Info("claims", zap.Any("claims", claims))
}
