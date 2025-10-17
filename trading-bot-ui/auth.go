package main

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// AuthManager handles simple PIN authentication
type AuthManager struct {
	pinHash     string
	pinFilePath string
	isLocked    bool
}

// NewAuthManager creates a new auth manager
func NewAuthManager() *AuthManager {
	// Get user config directory
	configDir, err := os.UserConfigDir()
	if err != nil {
		configDir = "."
	}

	pinPath := filepath.Join(configDir, "trading-bot", "auth.pin")

	return &AuthManager{
		pinFilePath: pinPath,
		isLocked:    true,
	}
}

// Initialize sets up auth (creates PIN on first run)
func (a *AuthManager) Initialize() error {
	// Check if PIN file exists
	if _, err := os.Stat(a.pinFilePath); os.IsNotExist(err) {
		// No PIN set - app is unlocked by default on first run
		a.isLocked = false
		return nil
	}

	// Load existing PIN hash
	data, err := os.ReadFile(a.pinFilePath)
	if err != nil {
		return fmt.Errorf("failed to read PIN file: %w", err)
	}

	a.pinHash = strings.TrimSpace(string(data))
	a.isLocked = true
	return nil
}

// SetPIN creates a new PIN (only callable when unlocked)
func (a *AuthManager) SetPIN(pin string) error {
	if a.isLocked {
		return fmt.Errorf("must unlock before setting PIN")
	}

	if len(pin) < 4 {
		return fmt.Errorf("PIN must be at least 4 digits")
	}

	// Hash the PIN
	hash := sha256.Sum256([]byte(pin))
	a.pinHash = hex.EncodeToString(hash[:])

	// Create config directory if it doesn't exist
	dir := filepath.Dir(a.pinFilePath)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	// Save hashed PIN to file (0600 = read/write for owner only)
	if err := os.WriteFile(a.pinFilePath, []byte(a.pinHash), 0600); err != nil {
		return fmt.Errorf("failed to save PIN: %w", err)
	}

	return nil
}

// Unlock verifies PIN and unlocks the app
func (a *AuthManager) Unlock(pin string) error {
	if !a.isLocked {
		return nil // Already unlocked
	}

	// No PIN set yet
	if a.pinHash == "" {
		a.isLocked = false
		return nil
	}

	// Hash provided PIN
	hash := sha256.Sum256([]byte(pin))
	providedHash := hex.EncodeToString(hash[:])

	// Compare hashes
	if providedHash != a.pinHash {
		return fmt.Errorf("incorrect PIN")
	}

	a.isLocked = false
	return nil
}

// Lock locks the app
func (a *AuthManager) Lock() {
	a.isLocked = true
}

// IsLocked returns lock status
func (a *AuthManager) IsLocked() bool {
	return a.isLocked
}

// HasPIN returns true if PIN is set
func (a *AuthManager) HasPIN() bool {
	return a.pinHash != ""
}

// ChangePIN changes existing PIN (must be unlocked)
func (a *AuthManager) ChangePIN(oldPIN, newPIN string) error {
	if a.isLocked {
		return fmt.Errorf("must unlock before changing PIN")
	}

	// Verify old PIN first
	hash := sha256.Sum256([]byte(oldPIN))
	oldHash := hex.EncodeToString(hash[:])

	if oldHash != a.pinHash {
		return fmt.Errorf("incorrect old PIN")
	}

	// Set new PIN
	return a.SetPIN(newPIN)
}

// RemovePIN removes PIN protection (must be unlocked)
func (a *AuthManager) RemovePIN() error {
	if a.isLocked {
		return fmt.Errorf("must unlock before removing PIN")
	}

	// Delete PIN file
	if err := os.Remove(a.pinFilePath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to remove PIN: %w", err)
	}

	a.pinHash = ""
	return nil
}
