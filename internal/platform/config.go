package platform

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