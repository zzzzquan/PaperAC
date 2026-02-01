package config

import "os"

func getBindAddress() string {
	port := os.Getenv("PORT")
	if port != "" {
		return ":" + port
	}
	return getEnv("BIND_ADDR", ":8080")
}
