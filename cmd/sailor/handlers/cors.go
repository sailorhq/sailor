package handlers

import (
	"slices"

	"github.com/valyala/fasthttp"
	"go.uber.org/zap"
)

func (sc *SailorCore) WithCors(handler fasthttp.RequestHandler) fasthttp.RequestHandler {
	return func(ctx *fasthttp.RequestCtx) {
		origin := string(ctx.Request.Header.Peek("Origin"))
		allowed := slices.Contains(sc.setting.Console.Hosts, origin)
		sc.Log.Info("origin check with allowed status", zap.String("origin", origin), zap.Bool("isAllowed", allowed))
		if allowed {
			setCorsHeaders(ctx, origin)
		}

		// Handle preflight requests
		if string(ctx.Method()) == "OPTIONS" {
			ctx.SetStatusCode(fasthttp.StatusOK)
			return
		}

		handler(ctx)
	}
}

func setCorsHeaders(ctx *fasthttp.RequestCtx, origin string) {
	ctx.Response.Header.Set("Access-Control-Allow-Origin", origin)
	ctx.Response.Header.Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
	ctx.Response.Header.Set("Access-Control-Allow-Headers", "Content-Type, Authorization, x-token, x-user, x-pass")
	ctx.Response.Header.Set("Access-Control-Expose-Headers", "x-resource-version")
	ctx.Response.Header.Set("Access-Control-Allow-Credentials", "true")
}
