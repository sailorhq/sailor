package handlers

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"

	bolt "go.etcd.io/bbolt"

	"github.com/codekidx/sailor/internal/types"
	diffmod "github.com/sergi/go-diff/diffmatchpatch"

	"github.com/go-playground/validator/v10"
)

func (sc *SailorCore) ConfigHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPatch:
		sc.patchConfig(w, r)
	case http.MethodGet:
		sc.getConfig(w, r)
	default:
		return
	}
}

func (sc *SailorCore) patchConfig(w http.ResponseWriter, r *http.Request) {
	enc := json.NewEncoder(w)

	params, err := sc.extractSailorParams(r)
	if err != nil {
		// TODO: log here!
		enc.Encode(ResponseMessage{Message: err.Error()})
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	db, err := sc.getDBConn(params)
	if err != nil {
		enc.Encode(ResponseMessage{Message: err.Error()})
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	b, _ := io.ReadAll(r.Body)
	defer r.Body.Close()

	if validJson := json.Valid(b); !validJson {
		w.WriteHeader(http.StatusInternalServerError)
		enc.Encode(ResponseMessage{Message: "not a valid json!"})
		return
	}

	configJson := buildConfig(db)
	differ := diffmod.New()

	var next int

	err = db.Update(func(tx *bolt.Tx) error {
		metaBucket := tx.Bucket([]byte(BUCKET_META))
		ruleBytes := metaBucket.Get([]byte(KEY_RULES))

		var data map[string]any
		json.Unmarshal(b, &data)

		var rules map[string]any
		json.Unmarshal(ruleBytes, &rules)

		if err := hasRuleForAllKeys(data, rules, "$root"); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			enc.Encode(ResponseMessage{Message: err.Error()})
			return err
		}
		if err := validateWithRules(data, rules); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			enc.Encode(ResponseMessage{Message: err.Error()})
			return err
		}

		newConfigJson := string(b)
		fmt.Println("newConfigJson: ", configJson, newConfigJson)
		// TODO :: convert it into a function
		diff := differ.DiffMain(configJson, newConfigJson, true)
		patchList := differ.PatchMake(configJson, newConfigJson, diff)
		patchh := differ.PatchToText(patchList)

		fmt.Println("patchh: ", patchh)

		diffBucket := tx.Bucket([]byte(BUCKET_DIFFS))

		max, _ := diffBucket.Cursor().Last()
		fmt.Println("max", string(max))

		next, _ = strconv.Atoi(string(max))
		next += 1

		return diffBucket.Put(fmt.Append(nil, next), []byte(patchh))
	})

	if err != nil {
		enc.Encode(err.Error())
	} else {
		enc.Encode(fmt.Sprintf("new deployment: %d created", next))

	}
}

// TODO :: we need to move rules checking to the frontend
func hasRuleForAllKeys(mainMap, subMap map[string]any, parent string) error {
	// OPT :: instead of throwing error one by one , we can get all the missing rule keys
	// and then form a single error at one time
	for key := range mainMap {
		keyPath := fmt.Sprintf("%s.%s", parent, key)
		fmt.Println("checking key: ", keyPath)
		if _, ok := subMap[key]; !ok {
			return fmt.Errorf("rule for %s not present", keyPath)
		} else if nestedMap, ok := mainMap[key].(map[string]any); ok {
			return hasRuleForAllKeys(nestedMap,
				subMap[key].(map[string]any),
				keyPath)
		}
	}
	return nil
}

func validateWithRules(data, rules map[string]any) error {
	validate := validator.New()

	fmt.Println("rules: ", rules)

	errMap := validate.ValidateMap(data, rules)
	if len(errMap) == 0 {
		return nil
	}

	var message string
	for _, v := range errMap {
		message += fmt.Sprintf("%s\n", v)
	}

	return errors.New(message)
}

func (sc *SailorCore) getConfig(w http.ResponseWriter, r *http.Request) {
	enc := json.NewEncoder(w)

	params, err := sc.extractSailorParams(r)
	if err != nil {
		// TODO: log here!
		enc.Encode(ResponseMessage{Message: err.Error()})
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	db, err := sc.getDBConn(params)
	if err != nil {
		enc.Encode(ResponseMessage{Message: err.Error()})
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	configJson := buildConfig(db)
	var builtConfig map[string]any
	json.Unmarshal([]byte(configJson), &builtConfig)
	w.Header().Set("Content-Type", "application/json")
	enc.Encode(builtConfig)

}

func buildConfig(db *bolt.DB) string {
	var configJson string
	db.View(func(tx *bolt.Tx) error {
		deploymentBucket := tx.Bucket([]byte(BUCKET_DEPLOYMENT))
		metaBucket := tx.Bucket([]byte(BUCKET_META))

		min := []byte("0")
		max := metaBucket.Get([]byte(KEY_DEPLOYED_VERSION))
		cur := deploymentBucket.Cursor()

		differ := diffmod.New()

		for depVer, depBytes := cur.Seek(min); depVer != nil && bytes.Compare(depVer, max) <= 0; depVer, depBytes = cur.Next() {
			var deployment types.Deployment
			json.Unmarshal(depBytes, &deployment)
			// fmt.Println("diff_ver", string(diff_ver), string(max))
			p, _ := differ.PatchFromText(string(deployment.Diff))
			configJson, _ = differ.PatchApply(p, configJson)
		}

		return nil
	})

	return configJson
}
