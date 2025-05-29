package config

import (
	"fmt"
	"os"
	"strings"

	"github.com/joho/godotenv"
)

type Cors struct {
	AllowedOrigins []string
}

type ServerConfig struct {
	Port string
	Cors
}

type DatabaseConfig struct {
	ConnString string
}

type Config struct {
	DatabaseConfig
	ServerConfig
}

func InitConfig() (c Config, e error) {

	err := godotenv.Load()
	if !(os.Getenv("OPTIONAL_LOAD_ENV_FILE") == "TRUE") && err != nil {
		return Config{}, fmt.Errorf("error loading .env file. OPTIONAL_LOAD_ENV_FILE=%s.\n", os.Getenv("OPTIONAL_LOAD_ENV_FILE"))
	}
	c.ServerConfig.Port = os.Getenv("LISTENING_PORT")
	c.ServerConfig.Cors.AllowedOrigins = strings.Split(os.Getenv("CORS_ALLOWED_ORIGINS"), ",")

	// database
	dbUrl := os.Getenv("DATABASE_URL")
	c.DatabaseConfig.ConnString = dbUrl

	// server
	return c, nil
}
