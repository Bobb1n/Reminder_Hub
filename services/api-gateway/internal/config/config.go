package config

import "os"

type Config struct {
    ServerPort        string
    AuthServiceURL    string
    CoreServiceURL    string
    CollectorServiceURL string
    InternalToken     string
    JWTSecret         string
}

func Load() *Config {
    return &Config{
        ServerPort:        getEnv("SERVER_PORT", "8080"),
        AuthServiceURL:    getEnv("AUTH_SERVICE_URL", "http://auth-service:8081"),
        CoreServiceURL:    getEnv("CORE_SERVICE_URL", "http://core-service:8082"),
        CollectorServiceURL: getEnv("COLLECTOR_SERVICE_URL", "http://collector-service:8084"),
        InternalToken:     getEnv("INTERNAL_API_TOKEN", "gateway-secret-token"),
        JWTSecret:         getEnv("JWT_SECRET", "your-jwt-secret-key"),
    }
}

func getEnv(key, defaultValue string) string {
    if value := os.Getenv(key); value != "" {
        return value
    }
    return defaultValue
}