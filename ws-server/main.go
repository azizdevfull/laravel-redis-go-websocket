package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"sync/atomic"
	"syscall"
	"time"

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
		if origin == "" {
			return true
		}
		return isOriginAllowed(origin)
	},
}

var activeConns atomic.Int64

func wsHandler(hub *Hub) http.HandlerFunc {
	secret := getEnv("WS_SECRET", "")
	maxConns, _ := strconv.ParseInt(getEnv("MAX_CONNECTIONS", "10000"), 10, 64)

	return func(w http.ResponseWriter, r *http.Request) {
		if activeConns.Load() >= maxConns {
			http.Error(w, "too many connections", http.StatusServiceUnavailable)
			return
		}

		channel := r.URL.Query().Get("channel")
		if channel == "" {
			http.Error(w, "channel required", http.StatusBadRequest)
			return
		}

		token := r.URL.Query().Get("token")
		if !validateToken(token, channel, secret) {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Println("upgrade error:", err)
			return
		}

		client := &Client{
			hub:     hub,
			conn:    conn,
			channel: channel,
			send:    make(chan []byte, sendBufferSize),
		}

		activeConns.Add(1)
		hub.register(client)

		go client.writePump()
		client.readPump() // blocks until disconnect

		activeConns.Add(-1)
	}
}

func startRedisSubscriber(ctx context.Context, rdb *redis.Client, hub *Hub) {
	prefix := getEnv("REDIS_PREFIX", "laravel-database-")
	pattern := prefix + "*"

	pubsub := rdb.PSubscribe(ctx, pattern)
	defer pubsub.Close()

	log.Printf("redis subscriber ready  pattern=%s", pattern)

	for {
		select {
		case <-ctx.Done():
			return
		case msg := <-pubsub.Channel():
			channel := strings.TrimPrefix(msg.Channel, prefix)
			hub.broadcast(channel, []byte(msg.Payload))
		}
	}
}

func metricsHandler(hub *Hub) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		fmt.Fprintf(w, "ws_connections_active %d\n", activeConns.Load())
		fmt.Fprintf(w, "ws_connections_by_channel %d\n", hub.ConnectionCount())
	}
}

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

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

	mux := http.NewServeMux()
	mux.HandleFunc("/ws", wsHandler(hub))
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	mux.HandleFunc("/metrics", metricsHandler(hub))

	server := &http.Server{
		Addr:    ":" + getEnv("PORT", "8080"),
		Handler: mux,
	}

	go func() {
		<-ctx.Done()
		log.Println("shutting down...")
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		server.Shutdown(shutdownCtx)
	}()

	log.Printf("WebSocket server listening on %s", server.Addr)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("server error: %v", err)
	}
	log.Println("server stopped")
}
