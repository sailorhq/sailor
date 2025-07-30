package handlers

import (
	"bytes"
	"fmt"
	"net/http"
	"slices"
	"strings"

	"github.com/golang-jwt/jwt/v5"
	"github.com/valyala/fasthttp"
	bolt "go.etcd.io/bbolt"
	"go.uber.org/zap"
	"k8s.io/apimachinery/pkg/util/json"
)

type RBACConstraints struct {
	Permissions []string `json:"permissions"`
	Roles       []string `json:"roles"`
	AllowedApps []string `json:"allowed_apps"`
}

func (sc *SailorCore) Authenticated(next fasthttp.RequestHandler, pnr RBACConstraints) fasthttp.RequestHandler {
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
		if params.ProjectKey != "" {
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
		enc := json.NewEncoder(ctx)
		params := extractSailorParams(ctx)
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
