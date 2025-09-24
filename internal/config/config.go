package config

import (
	"bufio"
	"errors"
	"io/fs"
	"os"
	"strings"
)

// Config contains application configuration sourced from environment variables.
type Config struct {
	GoogleAPIKey string
	ServerPort   string
}

// LoadEnvFile loads key=value pairs from the provided file into the process environment.
func LoadEnvFile(path string) error {
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			return errors.New("invalid line in env file: " + line)
		}
		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])
		if strings.HasPrefix(value, "\"") && strings.HasSuffix(value, "\"") && len(value) >= 2 {
			value = strings.Trim(value, "\"")
		}
		if err := os.Setenv(key, value); err != nil {
			return err
		}
	}

	if err := scanner.Err(); err != nil {
		return err
	}

	return nil
}

// Load reads environment variables to build a Config value.
func Load() (Config, error) {
	cfg := Config{
		GoogleAPIKey: os.Getenv("GOOGLE_MAPS_API_KEY"),
		ServerPort:   os.Getenv("PORT"),
	}

	if cfg.ServerPort == "" {
		cfg.ServerPort = "8080"
	}

	if cfg.GoogleAPIKey == "" {
		return Config{}, errors.New("GOOGLE_MAPS_API_KEY is required")
	}

	return cfg, nil
}

// LoadFromEnvFile first attempts to read an env file and ignores missing file errors.
func LoadFromEnvFile(path string) error {
	err := LoadEnvFile(path)
	if err != nil && !errors.Is(err, fs.ErrNotExist) {
		return err
	}
	return nil
}
