package app

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
)

type ErrorMsg struct {
	ErrorMessage string   `json:"error"`
	Messages     []string `json:"messages"`
}

type ValidationErrorMsg struct {
	IsValid      bool   `json:"is_valid"`
	ErrorMessage string `json:"error"`
}

type DeployErrorMsg struct {
	IsSuccess    bool   `json:"success"`
	ErrorMessage string `json:"error"`
}

type DeploySuccessMsg struct {
	IsSuccess bool `json:"success"`
}

type ValidationTrue struct {
	IsValid bool `json:"is_valid"`
}

/**
 * Throw Error and Exit the process
 */
func ThrowDeployError(err error) {
	ThrowDeployErrorMsg(fmt.Sprintf("Deploy Error: %s", err.Error()))
}

func ThrowDeployErrorMsg(message string) {
	// defer os.Exit(1)
	errOutput := DeployErrorMsg{
		IsSuccess:    false,
		ErrorMessage: message,
	}

	jsonByte, err := json.MarshalIndent(errOutput, "", "   ")
	if err != nil {
		log.Fatalf("Unable to encode JSON: %s", err)
	}

	fmt.Fprint(os.Stderr, string(jsonByte))
	os.Exit(1)
}

func ThrowValidationError(err error) {
	ThrowValidationErrorMsg(fmt.Sprintf("OpenAPI Validation Error: %s", err.Error()))
}

func ThrowValidationErrorMsg(message string) {
	// defer os.Exit(1)
	errOutput := ValidationErrorMsg{
		IsValid:      false,
		ErrorMessage: message,
	}

	jsonByte, err := json.MarshalIndent(errOutput, "", "   ")
	if err != nil {
		log.Fatalf("Unable to encode JSON: %s", err)
	}

	fmt.Fprint(os.Stderr, string(jsonByte))
	os.Exit(1)
}
