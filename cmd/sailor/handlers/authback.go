package handlers

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/coreos/go-oidc"
	"github.com/valyala/fasthttp"
	bolt "go.etcd.io/bbolt"
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

	if ss.OIDC == nil {
		ctx.SetStatusCode(http.StatusNotFound)
		enc.Encode(ResponseMessage{Message: "oidc sailor settings not present"})
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
		Scopes:       ss.OIDC.Scopes,
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

	// TODO :: we only marshal the things that we need
	var claims map[string]any
	if err := idToken.Claims(&claims); err != nil {
		ctx.SetStatusCode(http.StatusInternalServerError)
		enc.Encode(ResponseMessage{Message: "cannot parse claims from token"})
		return
	}

	err = sc.dbconns[BUCKET_ADMIN].Update(func(tx *bolt.Tx) error {
		userBucket := tx.Bucket([]byte(BUCKET_USERS))
		if userBucket == nil {
			// this case should not happen because while creating _admin sail
			// we create user bucket, so it is technically a unrecoverable error
			return errors.New("cannot find users volume to save into")
		}

		var (
			email string
			ok    bool
		)
		if email, ok = claims["email"].(string); !ok {
			sc.Log.Error("oidc:scopes must include email, because email serves of a sailor username for oidc based flows")
			return errors.New("")
		}
		userBytes := userBucket.Get([]byte(email))

		// this user has already been a part of sailor, so we only update
		// the claims for /token call
		var user DBUser
		if userBytes != nil {
			json.Unmarshal(userBytes, &user)
			claims["roles"] = user.Roles
			claims["permissions"] = user.Permissions
			claims["allowed_apps"] = user.AllowedApps

			token, err := jwtFromClaims(getMapClaims(claims), sc.setting.TokenKey)
			if err != nil {
				return errors.New("error while generating token")
			}
			user.Token = token
		} else {
			token, err := jwtFromClaims(getMapClaims(claims), sc.setting.TokenKey)
			if err != nil {
				return errors.New("error while generating token")
			}
			user = DBUser{
				Email:    email,
				Username: email,
				Token:    token,
			}
		}

		userBytes, err := json.Marshal(user)
		if err != nil {
			return err
		}
		return userBucket.Put([]byte(email), userBytes)
	})

	if err != nil {
		ctx.SetStatusCode(http.StatusInternalServerError)
		enc.Encode(ResponseMessage{Message: "error while updating sailor user"})
		return
	}

	enc.Encode(ResponseMessage{Message: "ok"})
}
