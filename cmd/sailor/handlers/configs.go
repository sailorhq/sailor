package handlers

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"

	bolt "go.etcd.io/bbolt"

	diffmod "github.com/sergi/go-diff/diffmatchpatch"

	"github.com/go-playground/validator/v10"
)

func ConfigHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPatch:
		patchConfig(w, r)
	case http.MethodGet:
		getConfig(w, r)
	default:
		return
	}
}

func patchConfig(w http.ResponseWriter, r *http.Request) {
	ns := r.URL.Query().Get("ns")
	app := r.URL.Query().Get("app")
	_ = r.URL.Query().Get("key")

	enc := json.NewEncoder(w)

	if ns == "" || app == "" {
		enc.Encode(ResponseMessage{Message: "namespace or app is empty"})
		return
	}

	dbpath := fmt.Sprintf("./configs/%s-%s.db", ns, app)
	if f, _ := os.Stat(dbpath); f == nil {
		enc.Encode(ResponseMessage{Message: "no such app in this namespace"})
		return
	}

	db, err := bolt.Open(dbpath, 0600, nil)
	if err != nil {
		panic(err)
	}
	defer db.Close()

	b, _ := io.ReadAll(r.Body)
	defer r.Body.Close()

	if validJson := json.Valid(b); !validJson {
		w.WriteHeader(http.StatusInternalServerError)
		enc.Encode(ResponseMessage{Message: err.Error()})
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

		if err := hasRuleForAllKeys(data, rules); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			enc.Encode(ResponseMessage{Message: err.Error()})
			return err
		}
		if err := validateWithRules(data, rules); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			enc.Encode(ResponseMessage{Message: err.Error()})
			return err
		}

		max := metaBucket.Get([]byte(KEY_DEPLOYED_VERSION))
		fmt.Println("max", string(max))
		next, err = strconv.Atoi(string(max))
		next += 1

		var buf bytes.Buffer
		json.Compact(&buf, b)
		newConfigJson := buf.String()
		fmt.Println("newConfigJson: ", configJson, newConfigJson)
		// TODO :: convert it into a function
		diff := differ.DiffMain(configJson, newConfigJson, true)
		patchList := differ.PatchMake(configJson, newConfigJson, diff)
		patchh := differ.PatchToText(patchList)

		fmt.Println("patchh: ", patchh)

		diffBucket := tx.Bucket([]byte(BUCKET_DIFFS))

		return diffBucket.Put(fmt.Append(nil, next), []byte(patchh))
	})

	if err != nil {
		enc.Encode(err.Error())
	} else {
		enc.Encode(fmt.Sprintf("new deployment: %d created", next))

	}
}

func hasRuleForAllKeys(mainMap, subMap map[string]any) error {
	for key := range mainMap {
		if _, ok := subMap[key]; !ok {
			return fmt.Errorf("rule for %s not present", key)
		} else if nestedMap, ok := mainMap[key].(map[string]any); ok {
			return hasRuleForAllKeys(nestedMap, subMap[key].(map[string]any))
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

func getConfig(w http.ResponseWriter, r *http.Request) {
	ns := r.URL.Query().Get("ns")
	app := r.URL.Query().Get("app")
	_ = r.URL.Query().Get("key")

	enc := json.NewEncoder(w)

	if ns == "" || app == "" {
		enc.Encode(ResponseMessage{Message: "namespace or app is empty"})
		return
	}

	dbpath := fmt.Sprintf("./configs/%s-%s.db", ns, app)
	if f, _ := os.Stat(dbpath); f == nil {
		enc.Encode(ResponseMessage{Message: "no such app in this namespace"})
		return
	}

	db, err := bolt.Open(dbpath, 0600, nil)
	if err != nil {
		panic(err)
	}
	defer db.Close()

	configJson := buildConfig(db)
	enc.Encode(configJson)
}

func buildConfig(db *bolt.DB) string {
	var configJson string
	db.View(func(tx *bolt.Tx) error {
		diffBuck := tx.Bucket([]byte(BUCKET_DIFFS))
		metaBucket := tx.Bucket([]byte(BUCKET_META))

		// TODO ::
		min := []byte("0")
		max := metaBucket.Get([]byte(KEY_DEPLOYED_VERSION))
		cur := diffBuck.Cursor()

		differ := diffmod.New()
		var applied []bool

		for diff_ver, diff := cur.Seek(min); diff_ver != nil && bytes.Compare(diff_ver, max) <= 0; diff_ver, diff = cur.Next() {
			fmt.Println("diff_ver", string(diff_ver), string(max))
			p, _ := differ.PatchFromText(string(diff))
			configJson, applied = differ.PatchApply(p, configJson)
			fmt.Println("applied: ", applied, "patch: ", string(diff))
		}

		return nil
	})

	return configJson
}
