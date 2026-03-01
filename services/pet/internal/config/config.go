


type Config struct {
	AppPort string

	PostgresUser     string
	PostgresPassword string
	PostgresDB       string
	PostgresHost     string
	PostgresPort     string

	ACLGRPCAddr string
}

func Load() *Config {
	return &Config{
		AppPort:          getEnv("APP_PORT", "8085"),
		PostgresUser:     getEnv("POSTGRES_USER", "pet_user"),
		PostgresPassword: getEnv("POSTGRES_PASSWORD", "supersecret"),
		PostgresDB:       getEnv("POSTGRES_DB", "pet_db"),
		PostgresHost:     getEnv("POSTGRES_HOST", "localhost"),
		PostgresPort:     getEnv("POSTGRES_PORT", "5435"),
		ACLGRPCAddr:      getEnv("ACL_GRPC_ADDR", "localhost:50057"),
	}
}

func getEnv(key, fallback string) string {
	if v, ok := os.LookupEnv(key); ok && v != "" {
		return v
	}
	return fallback
}