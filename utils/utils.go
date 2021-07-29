package utils

import (
	"github.com/joho/godotenv"
	"os"
)

func init() {
	err := godotenv.Load("../.env")
	if err != nil {
		customLog.Error("load env error:", err)
	}
}

func GetEnv(name string) string {
	if val := os.Getenv(name); val != "" {
		return val
	}
	customLog.Panicf("Missing env: %s", name)
	return ""
}
