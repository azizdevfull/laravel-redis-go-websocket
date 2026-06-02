package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/gorilla/websocket"
	"github.com/redis/go-redis/v9"
)

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func isOriginAllowed(origin string) bool {
	allowed := getEnv("ALLOWED_ORIGINS", "http://localhost:8000")
	for _, o := range strings.Split(allowed, ",") {
		if strings.TrimSpace(o) == origin {
			return true
		}
	}
	return false
}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		origin := r.Header.Get("Origin")
		// Non-browser clients (e.g. Go, Node) don't send Origin — allow them.
		// Browsers always send Origin, so we check against the allowlist.
		if origin == "" {
			return true
		}
		return isOriginAllowed(origin)
	},
}

func wsHandler(hub *Hub) http.HandlerFunc {
	secret := getEnv("WS_SECRET", "")

	return func(w http.ResponseWriter, r *http.Request) {
		token := r.URL.Query().Get("token")
		if !validateToken(token, secret) {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		channel := r.URL.Query().Get("channel")
		if channel == "" {
			http.Error(w, "channel required", http.StatusBadRequest)
			return
		}

		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Println("upgrade error:", err)
			return
		}
		defer conn.Close()

		hub.register(channel, conn)
		defer hub.unregister(channel, conn)

		log.Printf("+ client  channel=%s remote=%s", channel, r.RemoteAddr)

		// Read loop — detects disconnect and handles ping/pong
		for {
			if _, _, err := conn.ReadMessage(); err != nil {
				break
			}
		}

		log.Printf("- client  channel=%s remote=%s", channel, r.RemoteAddr)
	}
}

func startRedisSubscriber(ctx context.Context, rdb *redis.Client, hub *Hub) {
	prefix := getEnv("REDIS_PREFIX", "laravel-database-")
	pattern := prefix + "*"

	pubsub := rdb.PSubscribe(ctx, pattern)
	defer pubsub.Close()

	log.Printf("redis subscriber ready  pattern=%s", pattern)

	for msg := range pubsub.Channel() {
		channel := strings.TrimPrefix(msg.Channel, prefix)
		log.Printf("redis message  channel=%s len=%d", channel, len(msg.Payload))
		hub.broadcast(channel, []byte(msg.Payload))
	}
}

func main() {
	ctx := context.Background()

	redisPassword := getEnv("REDIS_PASSWORD", "")
	if redisPassword == "null" {
		redisPassword = ""
	}

	rdb := redis.NewClient(&redis.Options{
		Addr:     getEnv("REDIS_HOST", "127.0.0.1") + ":" + getEnv("REDIS_PORT", "6379"),
		Password: redisPassword,
	})

	if err := rdb.Ping(ctx).Err(); err != nil {
		log.Fatalf("redis connection failed: %v", err)
	}
	log.Println("redis connected")

	hub := newHub()
	go startRedisSubscriber(ctx, rdb, hub)

	http.HandleFunc("/ws", wsHandler(hub))
	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	addr := ":" + getEnv("PORT", "8080")
	log.Printf("WebSocket server listening on %s", addr)
	log.Fatal(http.ListenAndServe(addr, nil))
}
