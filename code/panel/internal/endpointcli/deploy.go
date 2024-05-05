package endpointcli

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/CSPF-Founder/api-firewall-panel/api-protector/onpremise/panel/internal/endpointinput"
	"github.com/CSPF-Founder/api-firewall-panel/api-protector/onpremise/panel/models"
)

// Handle Deploy function
func HandleDeploy(cliBinPath string, workDir string, endpoint models.Endpoint, newDeployment bool) error {

	_, openapiPath := models.GetOpenApiFilePath(endpoint, workDir)

	cmdArgs, err := BuildDeployArgs(workDir, openapiPath, endpoint)
	if err != nil {
		return err
	}

	output, err := RunCmd(cliBinPath, cmdArgs)
	if err != nil {
		return err
	}

	if output == nil {
		return errors.New("output is nil")
	}

	if output.IsSuccess {
		return nil
	}

	errorMsg := errors.New("Unable to deploy")
	if output.ErrorMessage != "" {
		errorMsg = errors.New(output.ErrorMessage)
	}

	// error path and perform cleanup here
	if newDeployment {
		// delete the file
		defer func() {
			os.Remove(openapiPath)
			_ = os.RemoveAll(filepath.Dir(openapiPath))
		}()
		//Delete the endpoint
		err := models.DeleteEndpoint(endpoint.ID)
		if err != nil {
			if output.ErrorMessage != "" {
				return fmt.Errorf("%s, %s", output.ErrorMessage, err.Error())
			} else {
				return err
			}
		}
	}

	return errorMsg
}

// Function to build deploy command dynamically
func BuildDeployArgs(workDir string, openapiPath string, endpoint models.Endpoint) (cmdArgs []string, err error) {

	cmdArgs = []string{
		"--module",
		"deploy",
		"--api-file",
		openapiPath,
		"--api-url",
		endpoint.ApiUrl,
		"--label",
		endpoint.Label,
		"--listen-port",
		strconv.Itoa(endpoint.ListeningPort),
		"--request-mode",
		endpoint.RequestMode,
		"--health-port",
		strconv.Itoa(endpoint.HealthPort),
	}

	epConfigs, err := models.GetConfigsByEndpointID(endpoint.ID)
	if err != nil {
		return nil, fmt.Errorf("Error fetching endpoint configurations")
	}

	ipFile, err := endpointinput.GenerateAllowedIpFile(workDir, endpoint)
	if err != nil && !strings.Contains(err.Error(), "Error getting allowed IPs") {
		// ignore error if it is due to no allowed IPs
		// return error if it is due to other reasons
		return nil, fmt.Errorf("Error generating allowed ip file")
	} else if ipFile != "" {
		cmdArgs = append(cmdArgs, "--allow-ip-file", ipFile)
	}

	// Check if Configutaion is not nil, if so, add it to the command
	if len(epConfigs) > 0 {
		for _, config := range epConfigs {
			switch config.ConfigKey {
			case models.CLIENT_IP_HEADER:
				cmdArgs = append(cmdArgs, "--client-ip-header", config.ConfigValue)
			case models.DENY_TOKEN_HEADER:
				deniedFile, err := endpointinput.GenerateDeniedTokenFile(workDir, endpoint)
				if err != nil {
					return nil, fmt.Errorf("Error generating denied token file")
				} else if deniedFile != "" {
					cmdArgs = append(cmdArgs, "--deny-token-header", config.ConfigValue)
					cmdArgs = append(cmdArgs, "--deny-token-file", deniedFile)
				}
			}
		}
	}

	return cmdArgs, nil
}
