package app

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/joho/godotenv"
)

type AppConfig struct {
	DockerDataDir  string
	ModuleName     string
	DockerImageTag string
}

var Config AppConfig

// use godot package to load/read the .env file and
// return the value of the key
func loadEnv() {
	//determin bin directory and load .env from there
	exe, err := os.Executable()
	if err != nil {
		log.Fatal(err)
	}
	binDir := filepath.Dir(exe)
	envPath := filepath.Join(binDir, ".env")
	if err := godotenv.Load(envPath); err == nil {
		return
	}

	// try to load .env from current directory
	envPath = ".env"
	if err := godotenv.Load(envPath); err == nil {
		return
	}

	panic("Unable to load .env file")
}

func getEnvValueOrError(key string) string {
	value := os.Getenv(key)
	if value == "" {
		panic(fmt.Sprintf("Environment variable %s not set", key))
	}
	return value
}
func InitApp() {
	loadEnv()

	Config = AppConfig{
		DockerImageTag: getEnvValueOrError("DOCKER_IMAGE_TAG"),
		DockerDataDir:  getEnvValueOrError("DOCKER_DATA_DIR"),
		ModuleName:     "api_protector",
	}
}
