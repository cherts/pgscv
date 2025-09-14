package service

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"github.com/cherts/pgscv/internal/log"
	"os"
	"slices"
	"sync"
)

// trustedScriptHashes contains pre-approved SHA256 hashes of trusted discovery scripts.
// Scripts with hashes in this list will be executed without security warnings.
// Format: slice of SHA256 hex strings (64 characters each).
// Example:
//
//	var trustedScriptHashes = []string{
//		"e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855",
//	}
var trustedScriptHashes = []string{}

var scriptHashesCache = make(map[string]string)
var scriptHashesMutex sync.RWMutex

// validateScriptHash validates the SHA256 hash of the discovery script file against
// security requirements. This function provides protection against unauthorized script
// modifications and execution of untrusted code.
//
// The function performs the following checks:
//  1. Calculates SHA256 hash of the script file content
//  2. Checks if the hash exists in the trustedScriptHashes list (logs warning if not)
//  3. Compares against cached hashes to detect script modifications
//  4. Blocks execution if script content has changed from previous runs
//
// Security behavior:
//   - First execution: Script runs regardless of trust status, but logs warnings for untrusted hashes
//   - Subsequent executions: Script runs only if hash matches previous execution
//   - Hash changes: Execution is blocked and warning is logged
//
// Returns:
//   - error: if script cannot be read, hash calculation fails, or script content has changed
//   - nil: if script is trusted or first execution of untrusted script
//
// Note: This is a security-critical function that helps prevent execution of modified
// or unauthorized scripts in the discovery pipeline.
func (s *ScriptDiscovery) validateScriptHash() error {
	scriptPath := s.config.scriptPath

	// G304 (CWE-22): Potential file inclusion via variable (Confidence: HIGH, Severity: MEDIUM)
	data, err := os.ReadFile(scriptPath) // #nosec G304
	if err != nil {
		return fmt.Errorf("[Script SD] failed to read script file %s: %w", scriptPath, err)
	}

	hash := sha256.Sum256(data)
	currentHash := hex.EncodeToString(hash[:])

	scriptHashesMutex.RLock()

	cachedHash, exists := scriptHashesCache[scriptPath]

	scriptHashesMutex.RUnlock()

	if exists && cachedHash != currentHash {
		log.Warnf("[Script SD] Script hash changed for %s. Previous: %s, Current: %s. Execution blocked.",
			scriptPath, cachedHash[:16]+"...", currentHash[:16]+"...")

		return fmt.Errorf("script hash changed, execution blocked")
	}

	isTrusted := slices.Contains(trustedScriptHashes, currentHash)

	if !isTrusted {
		log.Warnf("[Script SD] Untrusted script hash for %s: %s. Script will be executed but consider adding this hash to trusted list.",
			scriptPath, currentHash)
	} else {
		log.Debugf("[Script SD] Script hash verified for %s: %s", scriptPath, currentHash[:16]+"...")
	}

	scriptHashesMutex.Lock()
	scriptHashesCache[scriptPath] = currentHash
	scriptHashesMutex.Unlock()

	return nil
}
