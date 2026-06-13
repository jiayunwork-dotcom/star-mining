package main

import (
	"log"
	"os"
	"strconv"

	"star-mining/internal/cache"
	"star-mining/internal/server"
)

func main() {
	port := getEnv("PORT", "8080")
	redisAddr := getEnv("REDIS_ADDR", "localhost:6379")
	redisPassword := getEnv("REDIS_PASSWORD", "")
	redisDB := getEnvInt("REDIS_DB", 0)
	staticDir := getEnv("STATIC_DIR", "../frontend/dist")

	redisCache, err := cache.NewRedisCache(redisAddr, redisPassword, redisDB)
	if err != nil {
		log.Printf("Warning: Failed to connect to Redis: %v", err)
		log.Println("Continuing without Redis cache...")
		redisCache = nil
	} else {
		defer redisCache.Close()
		log.Println("Redis connection established")
	}

	addr := ":" + port
	httpServer := server.NewHTTPServer(addr, staticDir, redisCache)

	log.Printf("Server starting on port %s", port)
	log.Printf("Static files directory: %s", staticDir)
	if err := httpServer.Start(); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}

func getEnv(key string, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}

func getEnvInt(key string, defaultValue int) int {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	intValue, err := strconv.Atoi(value)
	if err != nil {
		return defaultValue
	}
	return intValue
}
