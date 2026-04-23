package config

import (
	"os"

	"github.com/joho/godotenv"
)

func init() {
	_ = godotenv.Load()
}

type GateConfig struct {
	BaseURL  string
	Token    string
	User     string
	Pass     string
	CertFile string
	KeyFile  string
	Insecure bool
}

func LoadGateConfig() GateConfig {
	return GateConfig{
		BaseURL:  getEnv("GATE_URL", "http://localhost:8084"),
		Token:    getEnv("GATE_TOKEN", ""),
		User:     getEnv("GATE_USER", ""),
		Pass:     getEnv("GATE_PASS", ""),
		CertFile: getEnv("GATE_CERT_FILE", ""),
		KeyFile:  getEnv("GATE_KEY_FILE", ""),
		Insecure: getEnv("GATE_INSECURE", "") == "true",
	}
}

func getEnv(k, d string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return d
}
