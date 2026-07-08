package config

import (
	"fmt"
	"os"
	"strconv"
)

type Config struct {
	AppEnv  string
	AppPort string

	DatabaseURL            string
	DBMaxConns             int32
	DBConnTimeout          int
	SupabaseURL            string
	SupabaseServiceRoleKey string
	SupabaseJWTSecret      string

	NIMAPIKey  string
	NIMBaseURL string

	R2AccountID     string
	R2AccessKey     string
	R2SecretKey     string
	R2Bucket        string
	R2PublicBaseURL string

	OTELEndpoint    string
	OTELServiceName string

	RateLimitRPM int
}

func Load() (*Config, error) {
	cfg := &Config{
		AppEnv:          getEnv("APP_ENV", "dev"),
		AppPort:         getEnv("APP_PORT", "8080"),
		DBMaxConns:      int32(getEnvInt("DB_MAX_CONNS", 10)),
		DBConnTimeout:   getEnvInt("DB_CONN_TIMEOUT_SEC", 5),
		NIMBaseURL:      getEnv("NIM_BASE_URL", "https://integrate.api.nvidia.com/v1"),
		OTELServiceName: getEnv("OTEL_SERVICE_NAME", "product-intelligence"),
		RateLimitRPM:    getEnvInt("RATE_LIMIT_RPM", 100),
	}

	required := map[string]*string{
		"DATABASE_URL":              &cfg.DatabaseURL,
		"SUPABASE_URL":              &cfg.SupabaseURL,
		"SUPABASE_SERVICE_ROLE_KEY": &cfg.SupabaseServiceRoleKey,
		"SUPABASE_JWT_SECRET":       &cfg.SupabaseJWTSecret,
		"NVIDIA_NIM_API_KEY":        &cfg.NIMAPIKey,
		"R2_ACCOUNT_ID":             &cfg.R2AccountID,
		"R2_ACCESS_KEY":             &cfg.R2AccessKey,
		"R2_SECRET_KEY":             &cfg.R2SecretKey,
		"R2_BUCKET":                 &cfg.R2Bucket,
		"OTEL_ENDPOINT":             &cfg.OTELEndpoint,
	}

	missing := []string{}
	for key, dest := range required {
		val := os.Getenv(key)
		if val == "" {
			missing = append(missing, key)
			continue
		}
		*dest = val
	}

	cfg.R2PublicBaseURL = os.Getenv("R2_PUBLIC_BASE_URL")

	if len(missing) > 0 {
		return nil, fmt.Errorf("missing required environment variables: %v", missing)
	}

	return cfg, nil
}

func (c *Config) IsProd() bool { return c.AppEnv == "prod" }
func (c *Config) IsDev() bool  { return c.AppEnv == "dev" }

func getEnv(key, defaultVal string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultVal
}

func getEnvInt(key string, defaultVal int) int {
	val := os.Getenv(key)
	if val == "" {
		return defaultVal
	}
	n, err := strconv.Atoi(val)
	if err != nil {
		return defaultVal
	}
	return n
}
