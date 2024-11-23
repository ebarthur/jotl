package config

import (
	"fmt"
	"os"
	"path/filepath"
	"text/template"
)

type PostgresConfig struct {
	ContainerName string
	DBName        string
	User          string
	Password      string
	Port          string
	Volume        string
}

func DefaultPostgresConfig() PostgresConfig {
	return PostgresConfig{
		ContainerName: "jotl-postgres",
		DBName:        "jotl",
		User:          "jotl",
		Password:      "jotl_password",
		Port:          "5432",
		Volume:        "postgres_data",
	}
}

const dockerComposeTemplate = `version: '3.8'

services:
  postgres:
    image: postgres:latest
    container_name: {{.ContainerName}}
    environment:
      POSTGRES_DB: {{.DBName}}
      POSTGRES_USER: {{.User}}
      POSTGRES_PASSWORD: {{.Password}}
    ports:
      - "{{.Port}}:5432"
    volumes:
      - {{.Volume}}:/var/lib/postgresql/data
    restart: unless-stopped

volumes:
  {{.Volume}}:
    driver: local`

func (pc *PostgresConfig) CreateDockerCompose(currentDir string, config PostgresConfig) error {
	dockerComposePath := filepath.Join(currentDir, "jotl", "docker-compose.yml")

	// Parse template
	tmpl, err := template.New("docker-compose").Parse(dockerComposeTemplate)
	if err != nil {
		return fmt.Errorf("failed to parse docker-compose template: %w", err)
	}

	// Create file
	f, err := os.Create(dockerComposePath)
	if err != nil {
		return fmt.Errorf("failed to create docker-compose.yml: %w", err)
	}
	defer f.Close()

	// Execute template
	if err := tmpl.Execute(f, config); err != nil {
		return fmt.Errorf("failed to execute template: %w", err)
	}
	return nil
}
