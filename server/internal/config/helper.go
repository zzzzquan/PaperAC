package config

import "os"

func getBindAddress() string {
	port := os.Getenv("PORT")
	if port != "" {
		return "0.0.0.0:" + port
	}
	return getEnv("BIND_ADDR", "0.0.0.0:8080")
}
