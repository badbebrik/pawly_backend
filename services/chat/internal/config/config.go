package config

import "pawly/pkg/configenv"

type Config struct {
	AppPort           string
	JWTSecret         string
	PostgresUser      string
	PostgresPassword  string
	PostgresDB        string
	PostgresHost      string
	PostgresPort      string
	ACLGRPCAddr       string
	ProfileGRPCAddr   string
	PetGRPCAddr       string
	RedisAddr         string
	RedisPassword     string
	RedisDB           int
	RedisChannel      string
	PresenceTTL       int
	PresenceHeartbeat int
}

func Load() (*Config, error) {
	redisDB, err := configenv.Int("REDIS_DB", 0)
	if err != nil {
		return nil, err
	}
	presenceTTL, err := configenv.Int("PRESENCE_TTL_SECONDS", 45)
	if err != nil {
		return nil, err
	}
	presenceHeartbeat, err := configenv.Int("PRESENCE_HEARTBEAT_SECONDS", 15)
	if err != nil {
		return nil, err
	}

	return &Config{
		AppPort:           configenv.String("APP_PORT", "8090"),
		JWTSecret:         configenv.String("JWT_SECRET", ""),
		PostgresUser:      configenv.String("POSTGRES_USER", ""),
		PostgresPassword:  configenv.String("POSTGRES_PASSWORD", ""),
		PostgresDB:        configenv.String("POSTGRES_DB", ""),
		PostgresHost:      configenv.String("POSTGRES_HOST", "localhost"),
		PostgresPort:      configenv.String("POSTGRES_PORT", "5432"),
		ACLGRPCAddr:       configenv.String("ACL_GRPC_ADDR", ""),
		ProfileGRPCAddr:   configenv.String("PROFILE_GRPC_ADDR", ""),
		PetGRPCAddr:       configenv.String("PET_GRPC_ADDR", ""),
		RedisAddr:         configenv.String("REDIS_ADDR", "localhost:6379"),
		RedisPassword:     configenv.String("REDIS_PASSWORD", ""),
		RedisDB:           redisDB,
		RedisChannel:      configenv.String("REDIS_CHANNEL", "chat.events"),
		PresenceTTL:       presenceTTL,
		PresenceHeartbeat: presenceHeartbeat,
	}, nil
}
