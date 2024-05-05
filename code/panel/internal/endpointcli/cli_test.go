package endpointcli

import (
	"strings"
	"testing"
)

func TestRunCmdInvalid(t *testing.T) {
	cmdToRun := "invalid"
	args := []string{"-invalid"}

	_, err := RunCmd(cmdToRun, args)
	if err == nil {
		t.Errorf("Expected error, got nil")
	}

	if !strings.Contains(err.Error(), "executable file not found") {
		t.Errorf("Expected error message: Invalid command, got %s", err.Error())
	}
}
