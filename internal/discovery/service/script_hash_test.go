package service

import (
	"crypto/sha256"
	"encoding/hex"
	"testing"

	"os"
	"path/filepath"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func addTrustedScriptHash(hash string) {
	trustedScriptHashes = append(trustedScriptHashes, hash)
}

func TestScriptDiscovery_validateScriptHash(t *testing.T) {
	sd := NewScriptDiscovery()

	tmpDir := t.TempDir()
	scriptPath := filepath.Join(tmpDir, "test_script.sh")
	scriptContent := "#!/bin/sh\necho 'test'"
	err := os.WriteFile(scriptPath, []byte(scriptContent), 0755)
	require.NoError(t, err)

	config := scriptConfig{
		Script:           scriptPath,
		ExecutionTimeout: "10s",
		RefreshInterval:  "30s",
	}

	err = sd.Init(config)
	require.NoError(t, err)

	err = sd.validateScriptHash()
	assert.NoError(t, err)

	err = sd.validateScriptHash()
	assert.NoError(t, err)

	err = os.WriteFile(scriptPath, []byte("#!/bin/sh\necho 'modified'"), 0755)
	require.NoError(t, err)

	err = sd.validateScriptHash()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "script hash changed")
}

func TestScriptDiscovery_validateScriptHash_Trusted(t *testing.T) {
	testContent := "#!/bin/sh\necho 'trusted'"
	hash := sha256.Sum256([]byte(testContent))
	trustedHash := hex.EncodeToString(hash[:])

	addTrustedScriptHash(trustedHash)

	sd := NewScriptDiscovery()

	tmpDir := t.TempDir()
	scriptPath := filepath.Join(tmpDir, "trusted_script.sh")
	err := os.WriteFile(scriptPath, []byte(testContent), 0755)
	require.NoError(t, err)

	config := scriptConfig{
		Script:           scriptPath,
		ExecutionTimeout: "10s",
		RefreshInterval:  "30s",
	}

	err = sd.Init(config)
	require.NoError(t, err)

	err = sd.validateScriptHash()
	assert.NoError(t, err)
}
