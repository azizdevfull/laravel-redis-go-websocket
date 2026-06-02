package main

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"strconv"
	"strings"
	"time"
)

const tokenTTL = 60 // seconds

// validateToken checks "userId:timestamp:hmac_sha256(userId:timestamp, secret)"
func validateToken(token, secret string) bool {
	if secret == "" {
		return false
	}

	parts := strings.SplitN(token, ":", 3)
	if len(parts) != 3 {
		return false
	}

	ts, err := strconv.ParseInt(parts[1], 10, 64)
	if err != nil {
		return false
	}

	if time.Now().Unix()-ts > tokenTTL {
		return false
	}

	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(parts[0] + ":" + parts[1]))
	expected := hex.EncodeToString(mac.Sum(nil))

	return hmac.Equal([]byte(expected), []byte(parts[2]))
}
