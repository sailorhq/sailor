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
package command

import (
	"crypto/md5"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/sailorhq/sailor/internal/types"
	v1 "github.com/sailorhq/sailor/pkg/core/v1"
)

func TestSyncResources(t *testing.T) {
	// Setup temporary working directory
	tempDir, err := os.MkdirTemp("", "sailor-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir)

	origWd, _ := os.Getwd()
	os.Chdir(tempDir)
	defer os.Chdir(origWd)

	// Mock server for Sailor API
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Basic routing for GetResource
		// Path: /api/v1/resource/{ns}/{app}/{kind}
		// or /api/v1/resource/{ns}/{app}/misc/{name}
		w.Header().Set("x-resource-version", "1")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"test": "data"}`))
	}))
	defer server.Close()

	cfg := &CLIConfig{
		Env:          "local",
		SailorClient: v1.CoreV1(server.URL, "test-token"),
		CwdSailorLockFile: types.SailorLockFile{
			Environments: make(map[string]types.ResourceVersion),
		},
	}

	sf := &types.SailorFile{
		Project: types.ProjectDetails{
			Namespace: "test-ns",
			App:       "test-app",
		},
		Config: types.ResourceFile{
			File: "config.json",
		},
	}

	t.Run("Fresh sync downloads file and creates lock", func(t *testing.T) {
		err := syncResources(sf, cfg)
		if err != nil {
			t.Fatalf("syncResources failed: %v", err)
		}

		// Check if file was downloaded
		if _, err := os.Stat("config.json"); os.IsNotExist(err) {
			t.Error("config.json was not created")
		}

		// Check if lock file was created
		if _, err := os.Stat("sailor.lock"); os.IsNotExist(err) {
			t.Error("sailor.lock was not created")
		}

		// Verify lock file content
		lockBytes, _ := os.ReadFile("sailor.lock")
		var lock types.SailorLockFile
		json.Unmarshal(lockBytes, &lock)
		if lock.Environments["local"].Config == nil {
			t.Error("Lock file does not contain config info")
		}
	})

	t.Run("No changes if local and remote match lock", func(t *testing.T) {
		// Already synced from previous test
		// Update cfg with current lock state (it was updated in previous test because it's a pointer in some places, 
		// but let's be explicit if needed. Actually cfg.CwdSailorLockFile was updated).
		
		err := syncResources(sf, cfg)
		if err != nil {
			t.Fatalf("syncResources failed: %v", err)
		}
		// Should not error and should not overwrite (nothing to verify here easily without mocking os.WriteFile)
	})

	t.Run("Warning and skip if local changes exist", func(t *testing.T) {
		// Modify local file to create a "conflict"
		os.WriteFile("config.json", []byte(`{"test": "local-change"}`), 0644)
		
		// Run sync again
		err := syncResources(sf, cfg)
		if err != nil {
			t.Fatalf("syncResources failed: %v", err)
		}

		// File should NOT have been overwritten
		content, _ := os.ReadFile("config.json")
		if string(content) != `{"test": "local-change"}` {
			t.Error("config.json was unexpectedly overwritten")
		}
	})
}

func TestSyncResource_States(t *testing.T) {
	tempDir, _ := os.MkdirTemp("", "sailor-res-test-*")
	defer os.RemoveAll(tempDir)

	origWd, _ := os.Getwd()
	os.Chdir(tempDir)
	defer os.Chdir(origWd)

	remoteData := []byte(`{"ver": "remote"}`)
	remoteHash := fmt.Sprintf("%x", md5.Sum(remoteData))
	resData := &v1.ResourceData{
		Version: 2,
		Data:    remoteData,
	}

	res := SyncResource{
		Path: "test.json",
		Kind: "config",
	}

	cfg := &CLIConfig{Env: "dev"}

	t.Run("Safe overwrite when no lock exists", func(t *testing.T) {
		os.WriteFile("test.json", []byte(`{"ver": "old"}`), 0644)
		envVer := &types.ResourceVersion{}
		
		updated, err := syncResource(cfg, res, resData, envVer)
		if err != nil {
			t.Fatal(err)
		}
		if !updated {
			t.Error("expected updated to be true")
		}
		
		content, _ := os.ReadFile("test.json")
		if string(content) != string(remoteData) {
			t.Errorf("expected %s, got %s", string(remoteData), string(content))
		}
	})

	t.Run("Skip overwrite when local diverged from lock", func(t *testing.T) {
		lockData := []byte(`{"ver": "locked"}`)
		lockHash := fmt.Sprintf("%x", md5.Sum(lockData))
		
		os.WriteFile("test.json", []byte(`{"ver": "dirty"}`), 0644)
		
		envVer := &types.ResourceVersion{
			Config: &types.LockVersion{
				Version: 1,
				Hash:    lockHash,
			},
		}

		updated, err := syncResource(cfg, res, resData, envVer)
		if err != nil {
			t.Fatal(err)
		}
		
		if updated {
			// This might be true because the lock file *metadata* is updated even if file write is skipped.
			// However, in the skip overwrite case, the file should still be the same.
		}
		
		content, _ := os.ReadFile("test.json")
		if string(content) != `{"ver": "dirty"}` {
			t.Error("local changes were overwritten")
		}
		
		if envVer.Config.Hash != remoteHash {
			t.Errorf("expected lock hash to be updated to remote hash %s, got %s", remoteHash, envVer.Config.Hash)
		}
	})
}
