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
package main

import (
	"os"

	"github.com/sailorhq/sailor/cmd/sailor/handlers"
	"github.com/sailorhq/sailor/cmd/sailor/routes"
	"go.uber.org/zap"

	"github.com/fasthttp/router"
	"github.com/valyala/fasthttp"
)

// var staticConsoleFS embed.FS

var Version string

func main() {
	core := handlers.NewSailorCore()

	if core == nil {
		panic("core of sailor was unable to start! check for errors...")
	}

	core.Log.Info("sailor core initialized")

	r := router.New()
	apiV1 := r.Group("/api/v1")

	core.Log.Info("initializing core routes")

	// Register all route groups
	routes.RegisterSettingRoutes(apiV1, core)
	routes.RegisterAuthRoutes(apiV1, core)
	routes.RegisterProjectRoutes(apiV1, core)
	routes.RegisterResourceRoutes(apiV1, core)
	routes.RegisterAuditRoutes(apiV1, core)

	hooksK8s := r.Group("/hooks/k8s")
	routes.RegisterK8sHookRoutes(hooksK8s, core)

	// it is best to take SAILOR_PORT through ENV because people have their own
	// deployment flow
	port := os.Getenv("SAILOR_PORT")
	if port == "" {
		port = ":7766"
	}
	core.Log.Info("[🐧] sailor core: ", zap.String("version", Version))
	core.Log.Info("[🐧] starting core sailor server", zap.String("port", port))

	var coreHandler fasthttp.RequestHandler = r.Handler
	consoleHosts := core.GetConsoleHosts()
	if len(consoleHosts) > 0 {
		// we need to add CORS for console hosts
		core.Log.Info("enabling CORS for console hosts", zap.Strings("hosts", consoleHosts))
		coreHandler = core.WithCors(r.Handler)
	}

	if err := fasthttp.ListenAndServe(port, coreHandler); err != nil {
		core.Log.Error("unable to start sailor core", zap.Error(err))
	}
}
