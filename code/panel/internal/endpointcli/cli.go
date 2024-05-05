package endpointcli

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"os/exec"
)

type CliOutput struct {
	IsSuccess    bool     `json:"success"`
	IsValid      bool     `json:"is_valid"`
	ErrorMessage string   `json:"error"`
	Messages     []string `json:"messages"` // Additional messages
}

// RunCmd runs a command and returns the output
func RunCmd(cmdToRun string, args []string) (*CliOutput, error) {
	cmd := exec.Command(cmdToRun, args...)

	var outbuf, errbuf bytes.Buffer
	cmd.Stdout = &outbuf
	cmd.Stderr = &errbuf
	err := cmd.Run()
	if err != nil {
		if errbuf.Len() > 0 {
			var errOutput CliOutput
			if err := json.Unmarshal(errbuf.Bytes(), &errOutput); err != nil {
				return nil, fmt.Errorf("Error running command: %s", errbuf.String())
			} else {
				return &errOutput, nil
			}
		}
		return nil, errors.New("Error running command")
	}

	if errbuf.Len() > 0 {
		return nil, fmt.Errorf("Error running command: %s", errbuf.String())
	}

	var cliOutput CliOutput
	if err := json.Unmarshal(outbuf.Bytes(), &cliOutput); err != nil {
		return nil, err
	}

	return &cliOutput, nil

}
