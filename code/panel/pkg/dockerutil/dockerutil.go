package dockerutil

import (
	"encoding/json"
	"os/exec"
)

func GetDockerStatus(containerName string) (bool, error) {

	cmdArgs := []string{
		"container",
		"inspect",
		"--format='{{.State.Running}}'",
		containerName,
	}

	output, err := exec.Command("docker", cmdArgs...).CombinedOutput()
	if err != nil {
		return false, err
	}

	var result bool
	if err := json.Unmarshal(output, &result); err != nil {
		return result, err
	}

	return false, nil
}
