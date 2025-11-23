package handlers

import "github.com/valyala/fasthttp"

func (sc *SailorCore) K8sAdmissionHookHandler(ctx *fasthttp.RequestCtx) {
	ctx.WriteString("k8s admission hook handler")
}
