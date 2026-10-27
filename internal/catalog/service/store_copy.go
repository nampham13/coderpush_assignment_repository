package service

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
)

func newID() (string, error) {
	var bytes [16]byte
	if _, err := rand.Read(bytes[:]); err != nil {
		return "", fmt.Errorf("generate ID: %w", err)
	}
	return hex.EncodeToString(bytes[:]), nil
}

func cloneUpdate(update *Update) *Update {
	if update == nil {
		return nil
	}
	raw, _ := json.Marshal(update)
	var copy Update
	_ = json.Unmarshal(raw, &copy)
	return &copy
}
