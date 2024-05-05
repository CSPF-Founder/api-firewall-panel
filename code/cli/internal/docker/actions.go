package docker

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/CSPF-Founder/api-firewall-panel/api-protector/cli/internal/app"
	"github.com/CSPF-Founder/api-firewall-panel/api-protector/cli/pkg/fileutil"
)

// func CopyOpenApiFile(args app.Args) {
// 	openApiDirPath := filepath.Join(app.Config.DockerDataDir, "stack_"+args.Label, "volumes")
// 	err := os.MkdirAll(openApiDirPath, 0700)
// 	if err != nil {
// 		app.ThrowDeployError(err)
// 	}

// 	openAPiFileName := "openapi.yaml"
// 	bytesRead, err := os.ReadFile(args.OpenApiFile)
// 	if err != nil {
// 		app.ThrowDeployError(err)
// 	}

// 	err = os.WriteFile(filepath.Join(openApiDirPath, openAPiFileName), bytesRead, 0644)

// 	if err != nil {
// 		app.ThrowDeployError(err)
// 	}
// }

func CopyToDockerVolume(containerLabel string, srcFile string) {

	fileName := filepath.Base(srcFile)
	dockerVolumeDir := filepath.Join(app.Config.DockerDataDir, "stack_"+containerLabel, "volumes")
	destFile := filepath.Join(dockerVolumeDir, fileName)

	err := fileutil.CopyFile(destFile, srcFile)
	if err != nil {
		app.ThrowDeployError(err)
	}
}

func StartDockerCompose(args app.Args) {
	DockerComposeFilePath := filepath.Join(app.Config.DockerDataDir, "stack_"+args.Label, "docker-compose.yml")

	cmd := exec.Command("/usr/bin/docker", "compose", "-f", DockerComposeFilePath, "--progress", "plain", "up", "-d")
	output, _ := cmd.CombinedOutput()
	outputString := string(output)
	outputString = strings.ToLower(outputString)

	if !strings.Contains(outputString, "created") && !strings.Contains(outputString, "recreated") && strings.Contains(outputString, "running") {
		// when there is no change in docker-compose file, the service will not restart
		// force restart the service

		cmd = exec.Command("/usr/bin/docker", "compose", "-f", DockerComposeFilePath, "--progress", "plain", "restart")
		output, _ = cmd.CombinedOutput()
		outputString = string(output)
		outputString = strings.ToLower(outputString)
	}

	if strings.Contains(outputString, "started") || strings.Contains(outputString, "running") {
		success_msg := app.DeploySuccessMsg{
			IsSuccess: true,
		}
		marshaled, err := json.MarshalIndent(success_msg, "", "   ")
		if err != nil {
			log.Fatalf("Unable to encode JSON: %s", err)
		}
		fmt.Println(string(marshaled))
		os.Exit(0)
	} else {
		app.ThrowDeployErrorMsg("Failed to start")
	}
}

func UnDeployDockerCompose(args app.Args) {
	DockerComposeFilePath := filepath.Join(app.Config.DockerDataDir, "stack_"+args.Label, "docker-compose.yml")

	cmd := exec.Command("/usr/bin/docker", "compose", "-f", DockerComposeFilePath, "down")
	output, _ := cmd.CombinedOutput()
	outputString := string(output)

	if strings.Contains(outputString, "Removed") || strings.Contains(outputString, "No resource found") {
		success_msg := app.DeploySuccessMsg{
			IsSuccess: true,
		}
		marshaled, err := json.MarshalIndent(success_msg, "", "   ")
		if err != nil {
			log.Fatalf("Unable to encode JSON: %s", err)
		}
		fmt.Println(string(marshaled))
		os.Exit(0)
	} else {
		app.ThrowDeployErrorMsg("Failed to stop")
	}
}
