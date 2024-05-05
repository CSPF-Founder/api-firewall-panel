package endpointinput

import (
	"bufio"
	"errors"
	"fmt"
	"os"

	"github.com/CSPF-Founder/api-firewall-panel/api-protector/onpremise/panel/models"
	"github.com/CSPF-Founder/api-firewall-panel/api-protector/onpremise/panel/utils"
)

func GenerateDeniedTokenFile(workDir string, endpoint models.Endpoint) (string, error) {
	data, err := models.GetDeniedTokensByEndpointID(endpoint.ID)
	if err != nil {
		return "", errors.New("Error getting denied tokens")
	}

	if len(data) == 0 {
		return "", nil
	}

	// Check if the file exists
	inputDir, _ := models.GetOpenApiFilePath(endpoint, workDir)
	filename := fmt.Sprintf("%s/denied-tokens.txt", inputDir)
	exists := utils.FileExists(filename)

	var file *os.File

	// If the file already existed, truncate it to remove old data
	if exists {
		// Open the file in append mode
		file, err = os.OpenFile(filename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			return "", errors.New("Error opening file")
		}
		if err := file.Truncate(0); err != nil {
			return "", errors.New("Error truncating file")
		}
	} else {
		// Create a new text file
		file, err = os.Create(filename)
		if err != nil {
			return "", errors.New("Error creating file")
		}
	}
	defer file.Close()

	// Write data to the file
	writer := bufio.NewWriter(file)
	for _, item := range data {
		_, err := writer.WriteString(item.Token + "\n")
		if err != nil {
			return "", errors.New("Error writing to file")
		}
	}
	writer.Flush()

	return filename, nil
}

func GenerateAllowedIpFile(workDir string, endpoint models.Endpoint) (string, error) {
	data, err := models.GetAllowedIPByEndpointID(endpoint.ID)
	if err != nil {
		return "", errors.New("Error getting allowed IPs")
	}

	if len(data) == 0 {
		return "", nil
	}

	// Check if the file exists
	inputDir, _ := models.GetOpenApiFilePath(endpoint, workDir)
	filename := fmt.Sprintf("%s/allowed-ip.txt", inputDir)
	exists := utils.FileExists(filename)

	var file *os.File

	// If the file already existed, truncate it to remove old data
	if exists {
		// Open the file in append mode
		file, err = os.OpenFile(filename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			return "", errors.New("Error opening file")
		}
		if err := file.Truncate(0); err != nil {
			return "", errors.New("Error truncating file")
		}
	} else {
		// Create a new text file
		file, err = os.Create(filename)
		if err != nil {
			return "", errors.New("Error creating file")
		}
	}
	defer file.Close()

	// Write data to the file
	writer := bufio.NewWriter(file)
	for _, item := range data {
		_, err := writer.WriteString(item.IPData + "\n")
		if err != nil {
			return "", errors.New("Error writing to file")
		}
	}
	writer.Flush()

	return filename, nil
}
