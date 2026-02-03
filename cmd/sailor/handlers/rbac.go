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
	"errors"
	"fmt"
	"slices"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	v1 "github.com/sailorhq/sailor/pkg/core/v1"
	"github.com/valyala/fasthttp"
	bolt "go.etcd.io/bbolt"
)

func (sc *SailorCore) AuthRBACHandler(ctx *fasthttp.RequestCtx) {
	enc := json.NewEncoder(ctx)

	userKey := string(ctx.QueryArgs().Peek("user"))

	// TODO :: take the email value in some constant
	if strings.EqualFold(userKey, "admin@super.sailor") {
		ctx.SetStatusCode(fasthttp.StatusBadRequest)
		enc.Encode(ResponseMessage{Message: "cannot update RBAC for super admin"})
		return
	}

	var req v1.RBACRequest
	if err := json.Unmarshal(ctx.Request.Body(), &req); err != nil {
		ctx.SetStatusCode(fasthttp.StatusBadRequest)
		enc.Encode(ResponseMessage{Message: "cannot parse constraints"})
		return
	}

	var updated v1.RBACConstraints
	err := sc.dbconns[BUCKET_ADMIN].Update(func(tx *bolt.Tx) error {
		userBucket := tx.Bucket([]byte(BUCKET_USERS))
		userBytes := userBucket.Get([]byte(userKey))
		if userBytes == nil {
			return errors.New("no such sailor user")
		}
		var user DBUser
		json.Unmarshal(userBytes, &user)

		userConstraints := v1.RBACConstraints{
			Roles:       user.Roles,
			Permissions: user.Permissions,
			AllowedApps: user.AllowedApps,
		}

		updated = updateRBACConstraints(userConstraints, req)

		user.Roles = updated.Roles
		user.Permissions = updated.Permissions
		user.AllowedApps = updated.AllowedApps

		userBytes, err := json.Marshal(&user)
		if err != nil {
			return err
		}

		return userBucket.Put([]byte(userKey), userBytes)
	})

	if err != nil {
		ctx.SetStatusCode(fasthttp.StatusInternalServerError)
		enc.Encode(ResponseMessage{Message: err.Error()})
		return
	}

	claims := ctx.UserValue("__sailor_claims").(jwt.MapClaims)
	go sc.addAuditEvent(&v1.AuditEvent{
		Username:  claims["email"].(string),
		Action:    "update_rbac",
		Timestamp: time.Now(),
		Details: map[string]any{
			"for_user": userKey,
			"final":    updated,
		},
	})

	enc.Encode(ResponseMessage{Message: "ok"})
}

func (sc *SailorCore) GetRBACHandler(ctx *fasthttp.RequestCtx) {
	enc := json.NewEncoder(ctx)

	claims := ctx.UserValue("__sailor_claims").(jwt.MapClaims)

	var rbac v1.RBACConstraints
	var (
		constraints []any
		ok          bool
	)
	if constraints, ok = claims["roles"].([]any); !ok {
		ctx.SetStatusCode(fasthttp.StatusBadRequest)
		enc.Encode(ResponseMessage{Message: "invalid token #5"})
		return
	}
	rbac.Roles = anyToStringSlice(constraints)

	if constraints, ok = claims["permissions"].([]any); !ok {
		ctx.SetStatusCode(fasthttp.StatusBadRequest)
		enc.Encode(ResponseMessage{Message: "invalid token #6"})
		return
	}
	rbac.Permissions = anyToStringSlice(constraints)

	if constraints, ok = claims["allowed_apps"].([]any); !ok && !slices.Contains(rbac.Roles, "admin") {
		ctx.SetStatusCode(fasthttp.StatusBadRequest)
		enc.Encode(ResponseMessage{Message: "invalid token #7"})
		return
	}
	rbac.AllowedApps = anyToStringSlice(constraints)

	enc.Encode(rbac)
}

func updateRBACConstraints(current v1.RBACConstraints, patch v1.RBACRequest) v1.RBACConstraints {
	// addition
	for _, pr := range patch.Addition.Roles {
		if !slices.Contains(current.Roles, pr) {
			current.Roles = append(current.Roles, pr)
		}
	}

	for _, pp := range patch.Addition.Permissions {
		if !slices.Contains(current.Permissions, pp) {
			current.Permissions = append(current.Permissions, pp)
		}
	}

	for _, paa := range patch.Addition.AllowedApps {
		if !slices.Contains(current.AllowedApps, paa) {
			current.AllowedApps = append(current.AllowedApps, paa)
		}
	}

	// deletion
	for _, pr := range patch.Deletion.Roles {
		current.Roles = removeRBACElement(current.Roles, pr)
	}

	for _, pp := range patch.Deletion.Permissions {
		current.Permissions = removeRBACElement(current.Permissions, pp)
	}

	for _, paa := range patch.Deletion.AllowedApps {
		current.AllowedApps = removeRBACElement(current.AllowedApps, paa)
	}

	return current

}

func removeRBACElement(slice []string, target string) []string {
	for index, elem := range slice {
		if strings.EqualFold(elem, target) {
			if index+1 >= len(slice) {
				slice = slice[:index]
			} else {
				slice = append(slice[:index], slice[index+1:]...)
			}
		}
	}

	return slice
}

func anyToStringSlice(data []any) []string {
	// Create a new string slice with the same length as the original slice
	stringSlice := make([]string, len(data))

	for i, v := range data {
		// Type assertion: this will panic if v is not a string
		s, ok := v.(string)
		if !ok {
			// Handle the error, e.g., return an error or a default value
			fmt.Printf("Element at index %d is not a string, it is a %T\n", i, v)
			// Depending on requirements, you might want to return an error from this function
			continue
		}
		stringSlice[i] = s
	}

	return stringSlice
}
