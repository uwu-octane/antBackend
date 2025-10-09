package envloader

import (
	"log"
	"os"
	"path/filepath"

	"github.com/joho/godotenv"
)

func Load() {
	cwd, err := os.Getwd()
	if err != nil {
		log.Fatalf("Failed to get current working directory: %v", err)
	}

	rootEnv := filepath.Join(cwd, "..", ".env")
	localEnv := filepath.Join(cwd, ".env")

	loadIfExists(rootEnv)
	loadIfExists(localEnv)
}

func loadIfExists(path string) {
	if _, err := os.Stat(path); err == nil {
		if err := godotenv.Load(path); err != nil {
			log.Fatalf("Failed to load environment file %s: %v", path, err)
		} else {
			log.Printf("Loaded environment file %s", path)
		}
	}
}
