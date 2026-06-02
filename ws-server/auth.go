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

// validateToken checks "userId:timestamp:channel:hmac_sha256(userId:timestamp:channel, secret)"
// and verifies the token's embedded channel matches the requested channel.
func validateToken(token, channel, secret string) bool {
	if secret == "" {
		return false
	}

	parts := strings.SplitN(token, ":", 4)
	if len(parts) != 4 {
		return false
	}

	ts, err := strconv.ParseInt(parts[1], 10, 64)
	if err != nil {
		return false
	}

	if time.Now().Unix()-ts > tokenTTL {
		return false
	}

	if parts[2] != channel {
		return false
	}

	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(parts[0] + ":" + parts[1] + ":" + parts[2]))
	expected := hex.EncodeToString(mac.Sum(nil))

	return hmac.Equal([]byte(expected), []byte(parts[3]))
}
