package app

import (
	"flag"
)

type denyToken struct {
	HeaderName string
	FilePath   string
}

type allowIP struct {
	FilePath string
}

type Args struct {
	OpenApiFile string
	Label       string
	ListenPort  uint
	HealthPort  uint
	ApiUrl      string
	RequestMode string
	Module      string

	DenyToken      *denyToken
	AllowIP        *allowIP
	ClientIPHeader string
}

func ParseCli() (cliArgs Args, errorMessages []string) {

	flag.StringVar(&cliArgs.OpenApiFile, "api-file", "", "OpenAPI File Path")
	flag.StringVar(&cliArgs.ApiUrl, "api-url", "", "Full URL for remote api")
	flag.StringVar(&cliArgs.Label, "label", "", "Easy to remember name to reference the instence")
	flag.UintVar(&cliArgs.ListenPort, "listen-port", 0, "Port in which the secured API instence will be available")
	flag.UintVar(&cliArgs.HealthPort, "health-port", 0, "Port in which the health check API will be available")
	flag.StringVar(&cliArgs.RequestMode, "request-mode", "", "Request blocked or monitor")
	flag.StringVar(&cliArgs.Module, "module", "", "Module validate or deploy")

	var denyTokenHeaderName string
	var denyTokenFilePath string
	flag.StringVar(&denyTokenHeaderName, "deny-token-header", "", "Deny Token Header Name")
	flag.StringVar(&denyTokenFilePath, "deny-token-file", "", "Deny Token File Path")

	flag.StringVar(&cliArgs.ClientIPHeader, "client-ip-header", "", "Client IP Header Name")
	var allowIPFilePath string
	flag.StringVar(&allowIPFilePath, "allow-ip-file", "", "Allow IP File Path")

	flag.Parse()

	if allowIPFilePath != "" {
		cliArgs.AllowIP = &allowIP{
			FilePath: allowIPFilePath,
		}
	}

	if denyTokenHeaderName != "" && denyTokenFilePath != "" {
		cliArgs.DenyToken = &denyToken{
			HeaderName: denyTokenHeaderName,
			FilePath:   denyTokenFilePath,
		}
	}

	errorMessages = validateCliArgs(cliArgs, errorMessages)

	return cliArgs, errorMessages
}

func validateCliArgs(cliArgs Args, errorMessages []string) []string {
	switch cliArgs.Module {
	case "validate":
		if cliArgs.OpenApiFile == "" {
			errorMessages = append(errorMessages, "OpenAPI File not set")
		}
	case "undeploy":
		if cliArgs.Label == "" {
			errorMessages = append(errorMessages, "Label Input missing")
		}
	case "deploy":
		errorMessages = validateCliArgsForDeploy(cliArgs, errorMessages)
	default:
		errorMessages = append(errorMessages, "Invalid Action")

	}

	return errorMessages

}

func validateCliArgsForDeploy(cliArgs Args, errorMessages []string) []string {
	if cliArgs.OpenApiFile == "" {
		errorMessages = append(errorMessages, "OpenAPI File not set")
	}

	if cliArgs.ApiUrl == "" {
		errorMessages = append(errorMessages, "Remote API URL not set")
	}

	if cliArgs.Label == "" {
		errorMessages = append(errorMessages, "Label Input missing")
	}

	if cliArgs.ListenPort == 0 {
		errorMessages = append(errorMessages, "ListenPort not specified")
	}

	if cliArgs.HealthPort == 0 {
		errorMessages = append(errorMessages, "HealthPort not specified")
	}

	if cliArgs.RequestMode == "" {
		errorMessages = append(errorMessages, "Request mode not specified")
	}
	return errorMessages

}
