// sailor
// Copyright (C) 2025 SailorHQ and Ashish Shekar (codekidX)

// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.

// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.

// You should have received a copy of the GNU General Public License
// along with this program.  If not, see <http://www.gnu.org/licenses/>.
package handlers

import (
	"bytes"
	"fmt"
	"net/http"
	"slices"
	"strings"

	v1 "github.com/codekidx/sailor/pkg/core/v1"
	"github.com/golang-jwt/jwt/v5"
	"github.com/valyala/fasthttp"
	bolt "go.etcd.io/bbolt"
	"go.uber.org/zap"
	"k8s.io/apimachinery/pkg/util/json"
)

func (sc *SailorCore) Authenticated(next fasthttp.RequestHandler, pnr v1.RBACConstraints) fasthttp.RequestHandler {
	return func(ctx *fasthttp.RequestCtx) {
		enc := json.NewEncoder(ctx)
		token := string(ctx.Request.Header.Peek("x-token"))
		if strings.TrimSpace(string(token)) == "" {
			ctx.SetStatusCode(http.StatusUnauthorized)
			enc.Encode(map[string]any{
				"message": "unable to authenticate",
			})
			return
		}

		jwtToken, err := jwt.Parse(token, func(token *jwt.Token) (any, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("invalid signing method")
			}
			return []byte(sc.setting.TokenKey), nil
		})
		if err != nil {
			sc.Log.Error("error while parsing token", zap.Error(err))

			ctx.SetStatusCode(http.StatusBadRequest)
			enc.Encode(map[string]any{
				"message": "invalid token",
			})
			return
		}

		if !jwtToken.Valid {
			ctx.SetStatusCode(http.StatusForbidden)
			enc.Encode(map[string]any{
				"message": "provided sailor token is not valid",
			})
			return
		}

		mc := jwtToken.Claims.(jwt.MapClaims)
		if issuer, err := mc.GetIssuer(); err != nil || issuer != "sailor" {
			sc.Log.Warn("a token with unknown issuer was used",
				zap.String("iss", issuer),
				zap.String("from_ip", ctx.RemoteIP().String()))

			ctx.SetStatusCode(fasthttp.StatusForbidden)
			enc.Encode(ResponseMessage{"unknown token"})
			return
		}

		// we enter the realm of RBAC validations

		hasValidRole := validateRBAC(mc, "roles", pnr.Roles)
		hasValidPermissions := validateRBAC(mc, "permissions", pnr.Permissions)

		email := mc["email"].(string)
		if !hasValidRole || !hasValidPermissions {
			sc.Log.Error("invalid permissions",
				zap.String("email", email),
				zap.Bool("role_check_passed", hasValidRole),
				zap.Bool("permission_check_passed", hasValidPermissions),
			)
			ctx.SetStatusCode(http.StatusUnauthorized)
			enc.Encode(map[string]any{
				"message": "insufficient permissions",
			})
			return
		}

		sc.Log.Info("valid roles and permissions",
			zap.String("email", email),
			zap.String("req_path", ctx.Request.URI().String()),
			zap.String("req_method", string(ctx.Request.Header.Method())),
		)

		params := extractSailorParams(ctx)
		isProjectCreationPath := strings.HasPrefix(string(ctx.Path()), "/api/v1/project")
		if params.ProjectKey != "" && !isProjectCreationPath {
			// we will do validations only if we got the project key
			// if we don't have it we let the main handler throw an error
			isAllowedToPerformAction := validateRBAC(mc, "allowed_apps", []string{
				params.ProjectKey,
				fmt.Sprintf("%s-*", params.Ns), // namespace wildcard permission
			})

			if !isAllowedToPerformAction {
				sc.Log.Error("invalid project access",
					zap.String("email", email),
					zap.String("failed_for_ns", params.Ns),
				)
				ctx.SetStatusCode(http.StatusUnauthorized)
				enc.Encode(map[string]any{
					"message": "insufficient access",
				})
				return
			}

			msg := fmt.Sprintf("access provided to %s for %s", email, params.ProjectKey)
			sc.Log.Info(msg,
				zap.String("req_path", ctx.Request.URI().String()),
				zap.String("req_method", string(ctx.Request.Header.Method())),
			)
		}

		ctx.SetUserValue("__sailor_claims", mc)
		next(ctx)
	}
}

func validateRBAC(claims jwt.MapClaims, key string, target []string) bool {
	if len(target) == 0 {
		// this just means that there is no permission needed to perform this
		// action.
		return true
	}

	var (
		dst []string
		ok  bool
	)

	// cannot find the required RBAC key
	if _, ok = claims[key].([]any); !ok {
		return false
	}

	// when getting claims from JWT it doesn't know if the array has string
	// values, but we know it for sure.
	for _, v := range claims[key].([]any) {
		dst = append(dst, v.(string))
	}

	for _, t := range target {
		if slices.Contains(dst, t) {
			return true
		}
	}

	return false
}

func (sc *SailorCore) ClientCallable(next fasthttp.RequestHandler) fasthttp.RequestHandler {
	return func(ctx *fasthttp.RequestCtx) {
		// we identified that a token is passed, this is passed only when a client like CLI
		// or dashboard asks for resource, we check if the token is valid and user is
		// allowed to access this namespace and app
		params := extractSailorParams(ctx)
		if params.ProjectKey != "" && params.Token != "" {
			sc.Authenticated(next, v1.RBACConstraints{
				Roles:       []string{"user"}, // TODO :: pick from constant
				AllowedApps: []string{params.ProjectKey, fmt.Sprintf("%s-*", params.Ns)},
			})(ctx)
			return
		}

		enc := json.NewEncoder(ctx)
		if params.ProjectKey == "" {
			// this case will be handled by the main API handler
			// we cannot check for access-key and secret-key without
			// projectKey
			next(ctx)
			return
		}

		if _, ok := sc.dbconns[params.ProjectKey]; !ok {
			next(ctx)
			return
		}

		var accessKey []byte
		sc.dbconns[params.ProjectKey].View(func(tx *bolt.Tx) error {
			metaBucket := tx.Bucket([]byte(BUCKET_META))
			accessKey = metaBucket.Get([]byte(KEY_ACCESS_KEY))
			return nil
		})

		if !bytes.Equal([]byte(params.AccessKey), accessKey) {
			ctx.SetStatusCode(http.StatusForbidden)
			enc.Encode(map[string]any{
				"message": "not allowed to access this resource",
			})
			return
		}

		next(ctx)
	}
}
