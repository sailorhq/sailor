package handlers

import (
	"encoding/json"
	"errors"
	"slices"
	"strings"

	v1 "github.com/codekidx/sailor/pkg/core/v1"
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

		updated := updateRBACConstraints(userConstraints, req)

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

	enc.Encode(ResponseMessage{Message: "ok"})
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
