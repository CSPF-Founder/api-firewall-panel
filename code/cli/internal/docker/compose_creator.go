package docker

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/CSPF-Founder/api-firewall-panel/api-protector/cli/internal/app"
	"gopkg.in/yaml.v3"
)

type Config struct {
	Version  string             `yaml:"version"`
	Services map[string]Service `yaml:"services"`
}

type Service struct {
	ContainerName   string            `yaml:"container_name"`
	Image           string            `yaml:"image"`
	Restart         string            `yaml:"restart"`
	Environment     map[string]string `yaml:"environment"`
	Volumes         []string          `yaml:"volumes"`
	StopGracePeriod time.Duration     `yaml:"stop_grace_period"`
	NetworkMode     string            `yaml:"network_mode"`
	Logging         Logging           `yaml:"logging"`
}

type Logging struct {
	Driver  string            `yaml:"driver"`
	Options map[string]string `yaml:"options"`
}

func CreateDockerCompose(args app.Args) error {
	dockerVolumeDir := filepath.Join(app.Config.DockerDataDir, "stack_"+args.Label, "volumes")
	logsDirPath := filepath.Join(app.Config.DockerDataDir, "stack_"+args.Label, "logs")
	err := os.MkdirAll(logsDirPath, 0777)
	if err != nil {
		return err
	}

	dockerComposeFilePath := filepath.Join(app.Config.DockerDataDir, "stack_"+args.Label, "docker-compose.yml")

	requestMode := "LOG_ONLY"
	if args.RequestMode == "block" {
		requestMode = "BLOCK"
	}

	containerEnv := map[string]string{
		"APIFW_API_SPECS":                  "/opt/resources/openapi.yaml",
		"APIFW_HEALTH_HOST":                fmt.Sprintf("0.0.0.0:%d", args.HealthPort),
		"APIFW_LOG_FORMAT":                 "JSON",
		"APIFW_LOG_LEVEL":                  "INFO",
		"APIFW_REQUEST_VALIDATION":         requestMode,
		"APIFW_RESPONSE_VALIDATION":        "DISABLE",
		"APIFW_SERVER_DIAL_TIMEOUT":        "20s",
		"APIFW_SERVER_INSECURE_CONNECTION": "true",
		"APIFW_SERVER_MAX_CONNS_PER_HOST":  "512",
		"APIFW_SERVER_READ_TIMEOUT":        "20s",
		"APIFW_SERVER_WRITE_TIMEOUT":       "20s",
		"APIFW_SERVER_URL":                 args.ApiUrl,
		"APIFW_URL":                        fmt.Sprintf("http://0.0.0.0:%d", args.ListenPort),
	}

	if args.DenyToken != nil {
		fileName := filepath.Base(args.DenyToken.FilePath)
		targetPath := filepath.Join("/opt/resources", fileName)
		containerEnv["APIFW_DENYLIST_TOKENS_HEADER_NAME"] = args.DenyToken.HeaderName
		containerEnv["APIFW_DENYLIST_TOKENS_FILE"] = targetPath
	}

	if args.ClientIPHeader != "" {
		containerEnv["APIFW_ALLOW_IP_HEADER_NAME"] = args.ClientIPHeader
	}

	if args.AllowIP != nil {
		fileName := filepath.Base(args.AllowIP.FilePath)
		targetPath := filepath.Join("/opt/resources", fileName)
		containerEnv["APIFW_ALLOW_IP_FILE"] = targetPath
	}

	cfg := Config{
		Version: "3.8",
		Services: map[string]Service{
			args.Label: {
				ContainerName: args.Label,
				Image:         app.Config.DockerImageTag,
				Restart:       "on-failure",
				Environment:   containerEnv,
				Volumes: []string{
					dockerVolumeDir + ":/opt/resources:ro",
					logsDirPath + ":/var/tmp/logs/:rw",
				},
				StopGracePeriod: 1 * time.Second,
				NetworkMode:     "host",
				Logging: Logging{
					Driver: "json-file",
					Options: map[string]string{
						"max-size": "5m",
						"max-file": "5",
					},
				},
			},
		},
	}

	yamlData, err := yaml.Marshal(cfg)
	if err != nil {
		return err
	}

	err = os.WriteFile(dockerComposeFilePath, yamlData, 0700)
	if err != nil {
		return err
	}

	return nil
}
