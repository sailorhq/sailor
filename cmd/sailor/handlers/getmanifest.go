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
	"encoding/json"

	"github.com/valyala/fasthttp"
)

func (sc *SailorCore) GetManifestHandler(ctx *fasthttp.RequestCtx) {
	enc := json.NewEncoder(ctx)
	ss, err := getSailorSetting(sc.dbconns[BUCKET_ADMIN])
	if err != nil {
		ctx.SetStatusCode(fasthttp.StatusInternalServerError)
		enc.Encode(ResponseMessage{"sailor settings not found"})
		return
	}

	enc.Encode(ss.Manifest)
}
