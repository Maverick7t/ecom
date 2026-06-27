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
