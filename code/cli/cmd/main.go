package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/CSPF-Founder/api-firewall-panel/api-protector/cli/internal/app"
	"github.com/CSPF-Founder/api-firewall-panel/api-protector/cli/internal/docker"
	"github.com/CSPF-Founder/api-firewall-panel/api-protector/cli/internal/validators"
)

func main() {
	app.InitApp()

	cliArgs, errorMessages := app.ParseCli()
	if errorMessages != nil {
		errJSON := app.ErrorMsg{
			ErrorMessage: "Invalid Inputs",
			Messages:     errorMessages,
		}

		jsonByte, err := json.MarshalIndent(errJSON, "", "   ")
		if err != nil {
			log.Fatalf("Unable to encode JSON : %s", err)
		}

		fmt.Fprint(os.Stderr, string(jsonByte))
		os.Exit(1)

	}

	switch cliArgs.Module {
	case "validate":
		validators.ValidateOpenAPI(cliArgs)
	case "undeploy":
		validators.ValidateLabel(cliArgs.Label)
		docker.UnDeployDockerCompose(cliArgs)
	case "deploy":
		handleDeploy(cliArgs)
	default:
		errorMessages = append(errorMessages, "Invalid Action")
		errJSON := app.ErrorMsg{
			ErrorMessage: "Invalid Inputs",
			Messages:     errorMessages,
		}
		jsonByte, err := json.MarshalIndent(errJSON, "", "   ")
		if err != nil {
			log.Fatalf("Unable to encode JSON : %s", err)
		}

		fmt.Fprint(os.Stderr, string(jsonByte))
		os.Exit(1)
	}
}

func handleDeploy(cliArgs app.Args) {
	validators.ValidateOpenAPI(cliArgs)
	validators.ValidateUrl(cliArgs.ApiUrl)
	validators.ValidatePort(cliArgs.ListenPort, cliArgs.HealthPort)
	validators.ValidateLabel(cliArgs.Label)
	validators.ValidateRequestMode(cliArgs.RequestMode)

	// create docker volume dir:
	dockerVolumeDir := filepath.Join(app.Config.DockerDataDir, fmt.Sprintf("stack_%s", cliArgs.Label), "volumes")
	err := os.MkdirAll(dockerVolumeDir, 0700)
	if err != nil {
		app.ThrowDeployError(err)
	}
	docker.CopyToDockerVolume(cliArgs.Label, cliArgs.OpenApiFile)

	if cliArgs.DenyToken != nil && cliArgs.DenyToken.FilePath != "" {
		docker.CopyToDockerVolume(cliArgs.Label, cliArgs.DenyToken.FilePath)
	}

	if cliArgs.AllowIP != nil && cliArgs.AllowIP.FilePath != "" {
		docker.CopyToDockerVolume(cliArgs.Label, cliArgs.AllowIP.FilePath)
	}

	err = docker.CreateDockerCompose(cliArgs)
	if err != nil {
		app.ThrowDeployError(err)
	}

	docker.StartDockerCompose(cliArgs)
}
