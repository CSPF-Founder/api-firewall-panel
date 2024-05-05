package validators

import (
	"encoding/json"
	"fmt"
	"log"
	"net/url"
	"os"
	"regexp"

	"github.com/CSPF-Founder/api-firewall-panel/api-protector/cli/internal/app"
	"github.com/getkin/kin-openapi/openapi3"
)

func checkOpenAPIFile(filePath string) error {
	_, err := os.Stat(filePath)
	if err != nil {
		return err
	}

	loader := openapi3.NewLoader()
	swagger, err := loader.LoadFromFile(filePath)
	if err != nil {
		return err
	}

	// Validate the swagger
	err = swagger.Validate(loader.Context)
	if err != nil {
		return err
	}

	return nil
}

func ValidateOpenAPI(args app.Args) {
	err := checkOpenAPIFile(args.OpenApiFile)
	if err != nil {
		if args.Module == "validate" {
			app.ThrowValidationError(err)
		} else if args.Module == "deploy" {
			app.ThrowDeployError(err)
		}
	} else {
		// Set validation to true
		valid := app.ValidationTrue{
			IsValid: true,
		}
		marshaled, err := json.MarshalIndent(valid, "", "   ")
		if err != nil {
			log.Fatalf("Unable to encode JSON: %s", err)
		}
		if args.Module == "validate" {
			fmt.Println(string(marshaled))
			os.Exit(0)
		}
	}
}

func ValidateUrl(apiUrl string) {
	_, err := url.ParseRequestURI(apiUrl)
	if err != nil {
		app.ThrowDeployError(err)
	}
}

func ValidatePort(lisenPort uint, healthPort uint) {
	if lisenPort < 1024 || lisenPort > 65535 || healthPort < 1024 || healthPort > 65535 || healthPort == lisenPort {
		app.ThrowDeployErrorMsg("This is not a valid allowed port")
	}
}

func ValidateLabel(label string) {
	alphanumeric := regexp.MustCompile("^[a-z0-9-_]*$")
	if len(label) > 64 || !alphanumeric.MatchString(label) {
		app.ThrowDeployErrorMsg("Invalid Label")
	}
}

func ValidateRequestMode(requestMode string) {
	if requestMode != "monitor" {
		if requestMode != "block" {
			app.ThrowDeployErrorMsg("Invalid Mode Given")
		}
	}
}
