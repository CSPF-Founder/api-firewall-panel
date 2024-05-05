package controllers

import (
	"archive/zip"
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"time"

	ctx "github.com/CSPF-Founder/api-firewall-panel/api-protector/onpremise/panel/context"
	"github.com/CSPF-Founder/api-firewall-panel/api-protector/onpremise/panel/enums/flashtypes"
	"github.com/CSPF-Founder/api-firewall-panel/api-protector/onpremise/panel/internal/endpointcli"
	"github.com/CSPF-Founder/api-firewall-panel/api-protector/onpremise/panel/models"
	"github.com/CSPF-Founder/api-firewall-panel/api-protector/onpremise/panel/pkg/dockerutil"
	"github.com/CSPF-Founder/api-firewall-panel/api-protector/onpremise/panel/pkg/urlutil"
	"github.com/CSPF-Founder/api-firewall-panel/api-protector/onpremise/panel/utils"
	"github.com/CSPF-Founder/api-firewall-panel/api-protector/onpremise/panel/views"
	"github.com/go-chi/chi/v5"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

var allowedOpenAPIExt = []string{"yml", "yaml"}

type endpointsController struct {
	*App
}

func newEndpointsController(app *App) *endpointsController {
	return &endpointsController{
		App: app,
	}
}

func (c *endpointsController) registerRoutes() http.Handler {
	router := chi.NewRouter()

	// Authenticated Routes
	router.Group(func(r chi.Router) {
		// r.Use(mid.RequireLogin) //not needed here: already protected in main routes function

		// Routes for All Endpoints
		r.Get("/", c.ListEndpoints)             // List all Endpoints
		r.Post("/", c.AddEndPoint)              // Add a new endpoint
		r.Get("/add", c.DisplayAddEndpoint)     // Display add endpoint form
		r.Get("/list-edit", c.ListEditEndpoint) // List all Endpoints
		r.Get("/reconfig/list", c.ListReconfig) // List all Reconfig
		r.Get("/denied-tokens", c.DenyTokensPage)
		r.Get("/allowed-ips", c.AllowedIPsPage)

		// List all endpoints with configurations
		r.Route("/configs", func(r chi.Router) {
			r.Get("/ip-header", c.ListIPHeader)
			r.Get("/deny-token-header", c.ListTokenHeader)
		})

		// Routes for individual endpoints by ID
		r.Route("/{endpointID}", func(r chi.Router) {
			r.Post("/restart", c.RestartEndpoint)
			r.Get("/edit", c.DisplayEditEndpoint) // Display edit endpoint form
			r.Patch("/", c.UpdateEndpoint)        // Update an existing enpoint
			r.Delete("/", c.DeleteEndpoint)

			r.Patch("/mode", c.UpdateRequestMode)           // Update Request Mode
			r.Get("/reconfig/download", c.DownloadReconfig) // Download Reconfig

			// Deny Tokens Routes
			deniedTokenCtrl := newDeniedTokenController(c.App)
			r.Mount("/denied-tokens", deniedTokenCtrl.registerRoutes())

			//Allowed IP Routes
			allowedIPCtrl := newAllowedIPController(c.App)
			r.Mount("/allowed-ips", allowedIPCtrl.registerRoutes())

			epConfigCtrl := newEndpointConfigController(c.App)
			r.Mount("/configs", epConfigCtrl.registerRoutes())

			logsCtrl := newLogsController(c.App)
			r.Mount("/logs", logsCtrl.registerRoutes())
		})

	})

	return router
}

// Handle Deploy function
func (c *endpointsController) handleDeploy(endpoint models.Endpoint, newDeployment bool) (bool, error) {

	err := endpointcli.HandleDeploy(c.config.CliBinPath, c.config.WorkDir, endpoint, newDeployment)
	if err != nil {
		c.logger.Error("handleDeploy error", err)
		return false, errors.New("Unable to deploy")
	}

	return true, nil

}

// Transform String
func transformString(s string) string {
	parts := strings.Split(s, "_")
	for i, part := range parts {
		caser := cases.Title(language.English)
		parts[i] = caser.String(part)
	}
	return strings.Join(parts, " ")
}

// Show Add Endpoint Page
func (c *endpointsController) DisplayAddEndpoint(w http.ResponseWriter, r *http.Request) {
	templateData := views.NewTemplateData(c.config, c.session, r)
	templateData.Title = "Add Api Protector"
	if err := views.RenderTemplate(w, "endpoints/add", templateData); err != nil {
		c.logger.Error("Error rendering template: ", err)
	}
}

// Add Endpoint Handler
func (c *endpointsController) AddEndPoint(w http.ResponseWriter, r *http.Request) {
	err := r.ParseMultipartForm(10 << 20) // 10 MB
	if err != nil {
		c.SendJSONError(w, "Invalid file")
		return
	}

	file, handler, err := r.FormFile("yaml_file")
	if err != nil {
		c.SendJSONError(w, "Invalid file")
		return
	}
	defer file.Close()

	fileName := strings.ToLower(handler.Filename)
	fileType := filepath.Ext(fileName)[1:]

	if !slices.Contains(allowedOpenAPIExt, fileType) {
		c.SendJSONError(w, "Invalid file format")
		return
	}

	fileSize := handler.Size
	// filesize max 10 mb
	if fileSize > 10000000 {
		c.SendJSONError(w, "File size should be less than 10 MB")
		return
	}

	tmpFileName := filepath.Join(c.config.TempUploadsDir, fmt.Sprintf("%s.yaml", utils.GetRandomString(32)))
	out, err := os.Create(tmpFileName)
	if err != nil {
		c.logger.Error("Error creating file:", err)
		c.SendJSONError(w, "Unable to upload file")
		return
	}
	defer out.Close()

	_, err = io.Copy(out, file)
	if err != nil {
		c.logger.Error("Error copying file:", err)
		c.SendJSONError(w, "Invalid File")
		return
	}

	defer os.Remove(tmpFileName)

	c.handleEndpointAddition(w, r, tmpFileName)
}

// Show Edit Endpoint Page
func (c *endpointsController) DisplayEditEndpoint(w http.ResponseWriter, r *http.Request) {
	inputID := chi.URLParam(r, "endpointID")

	endpointID, err := strconv.ParseUint(inputID, 10, 64)
	if err != nil {
		c.FlashAndGoBack(w, r, flashtypes.FlashWarning, "Invalid Endpoint ID")
		return
	}
	user := ctx.Get(r, "user").(models.User)
	endpoint, err := models.GetEndpointByIDAndUser(endpointID, user.ID)
	if err != nil {
		c.FlashAndGoBack(w, r, flashtypes.FlashWarning, err.Error())
		return
	}
	templateData := views.NewTemplateData(c.config, c.session, r)
	templateData.Title = "Edit Endpoint"
	templateData.Data = endpoint
	if err := views.RenderTemplate(w, "endpoints/edit", templateData); err != nil {
		c.logger.Error("Error rendering template: ", err)
	}
}

// Update Endpoint Handler
func (c *endpointsController) UpdateEndpoint(w http.ResponseWriter, r *http.Request) {
	inputID := chi.URLParam(r, "endpointID")
	apiUrl := r.PostFormValue("api_url")
	requestMode := r.PostFormValue("request_mode")

	endpointID, err := strconv.ParseUint(inputID, 10, 64)
	if err != nil {
		c.SendJSONError(w, "Invalid Endpoint ID")
		return
	}

	user := ctx.Get(r, "user").(models.User)
	endpoint, err := models.GetEndpointByIDAndUser(endpointID, user.ID)
	if err != nil {
		c.SendJSONError(w, "Invalid Endpoint ID")
		return
	}

	perr := r.ParseMultipartForm(10 << 20)
	if perr != nil {
		c.SendJSONError(w, "Invalid file")
		return
	}

	tempFilePath, err := c.validateOpenApiFile(r)
	if err != nil {
		c.SendJSONError(w, err.Error())
		return
	}

	apiURLChanged := false
	requestModeChanged := false
	updatedItems := 0

	if apiUrl == endpoint.ApiUrl && endpoint.RequestMode == requestMode {
		c.SendJSONError(w, "No changes to update")
		return
	}

	if apiUrl != "" && apiUrl != endpoint.ApiUrl {
		if err := urlutil.IsServerReachable(apiUrl); err != nil {
			c.SendJSONError(w, fmt.Sprintf("UnableToReach API Server: %s", err.Error()))
			return
		}
		apiURLChanged = true
		endpoint.ApiUrl = apiUrl
	}

	if requestMode != "" && requestMode != endpoint.RequestMode {
		requestModeChanged = true
		endpoint.RequestMode = requestMode
	}

	defer os.Remove(tempFilePath)

	if tempFilePath != "" {
		targetDir, destinationFile := models.GetOpenApiFilePath(endpoint, c.config.WorkDir)

		if err := os.MkdirAll(targetDir, os.ModePerm); err != nil {
			// if err := os.MkdirAll(targetDir, 0777); err != nil {
			c.logger.Error("HandleJobCreation:", err)
			c.SendJSONError(w, "Unable to create target directory for OpenAPI file")
			return
		}

		if err := os.Rename(tempFilePath, destinationFile); err != nil {
			c.logger.Error("HandleJobCreation:", err)
			c.SendJSONError(w, "Unable to update the OpenAPI file")
			return
		}
		updatedItems++
	}

	if apiURLChanged || requestModeChanged {
		err := models.SaveEndpoint(&endpoint)
		if err != nil {
			c.SendJSONError(w, "Unable to update the endpoint")
			return
		}
		updatedItems++
	}

	if updatedItems > 0 {
		isSuccess, err := c.handleDeploy(endpoint, false)
		if !isSuccess {
			if err != nil {
				c.SendJSONError(w, err.Error())
				return
			} else {
				c.SendJSONError(w, "Unable to start the endpoint")
				return
			}
		}

		data := map[string]interface{}{
			"success":      "Successfully updated the API Protector",
			"api_url":      endpoint.ApiUrl,
			"request_mode": endpoint.RequestMode,
		}
		c.SendJSONSuccess(w, data)
		return
	}
	c.SendJSONError(w, "No changes to update")
}

// Validate OpenApi File
func (c *endpointsController) validateOpenApiFile(r *http.Request) (string, error) {
	file, handler, err := r.FormFile("yaml_file")
	if err != nil {
		return "", errors.New("Invalid file")
	}
	defer file.Close()

	fileName := handler.Filename
	fileType := filepath.Ext(fileName)[1:]

	if !slices.Contains(allowedOpenAPIExt, fileType) {
		return "", errors.New("Invalid file format")
	}

	fileSize := handler.Size
	// filesize max 10 mb
	if fileSize > 10000000 {
		return "", errors.New("File size should be less than 10 MB")
	}

	// tmpFileName := "/tmp/" + utils.GetRandomString(32) + ".yaml"
	tmpFileName := filepath.Join(c.config.TempUploadsDir, fmt.Sprintf("%s.yaml", utils.GetRandomString(32)))
	out, err := os.Create(tmpFileName)
	if err != nil {
		c.logger.Error("Error creating file:", err)
		return "", errors.New("Unable to upload file")
	}

	defer out.Close()

	_, err = io.Copy(out, file)
	if err != nil {
		c.logger.Error("Error copying file:", err)
		return "", errors.New("Invalid File")
	}

	cmdArgs := []string{
		"--module",
		"validate",
		"--api-file",
		tmpFileName,
	}
	output, err := endpointcli.RunCmd(c.config.CliBinPath, cmdArgs)
	if err != nil {
		c.logger.Error("Error validating file:", err)
		return "", errors.New("Error validating file")
	}

	if output == nil {
		c.logger.Error("Error validating file: output is nil", nil)
		return "", errors.New("Error validating file")
	}

	if !output.IsValid {
		if output.ErrorMessage == "" {
			return "", errors.New("Error validating file")
		} else {
			return "", errors.New(output.ErrorMessage)
		}
	}

	return tmpFileName, nil

}

// RestartEndpoint is used to restart the endpoint
func (c *endpointsController) RestartEndpoint(w http.ResponseWriter, r *http.Request) {
	inputID := chi.URLParam(r, "endpointID")

	endpointID, err := strconv.ParseUint(inputID, 10, 64)
	if err != nil {
		c.SendError(w, r, "Invalid Endpoint ID")
		return
	}

	user := ctx.Get(r, "user").(models.User)
	endpoint, err := models.GetEndpointByIDAndUser(endpointID, user.ID)
	if err != nil {
		c.SendError(w, r, "Invalid Endpoint ID")
		return
	}

	isSuccess, err := c.handleDeploy(endpoint, false)
	if !isSuccess {
		if err != nil {
			c.SendError(w, r, err.Error())
			return
		} else {
			c.SendError(w, r, "Unable to start the endpoint")
			return
		}
	}

	c.SendSuccess(w, r, "Successfully restarted the endpoint")
}

// Show Endpoints List
func (c *endpointsController) ListEndpoints(w http.ResponseWriter, r *http.Request) {
	user := ctx.Get(r, "user").(models.User)
	data, err := models.GetEndpoints(&user)
	if err != nil {
		c.SendJSONError(w, err.Error())
	}
	templateData := views.NewTemplateData(c.config, c.session, r)
	templateData.Title = "API Protector Endpoints"
	templateData.Data = data
	if err := views.RenderTemplate(w, "endpoints/list", templateData); err != nil {
		c.logger.Error("Error rendering template: ", err)
	}
}

// Show List edit Endpoint
func (c *endpointsController) ListEditEndpoint(w http.ResponseWriter, r *http.Request) {
	user := ctx.Get(r, "user").(models.User)
	data, err := models.GetEndpoints(&user)
	if err != nil {
		c.SendJSONError(w, err.Error())
	}
	templateData := views.NewTemplateData(c.config, c.session, r)
	templateData.Title = "Edit API Protector"
	templateData.Data = data
	if err := views.RenderTemplate(w, "endpoints/api_list", templateData); err != nil {
		c.logger.Error("Error rendering template: ", err)
	}
}

// Delete Endpoint Handler
func (c *endpointsController) DeleteEndpoint(w http.ResponseWriter, r *http.Request) {
	inputID := chi.URLParam(r, "endpointID")

	endpointID, err := strconv.ParseUint(inputID, 10, 64)
	if err != nil {
		c.SendError(w, r, "Invalid Endpoint ID")
		return
	}

	user := ctx.Get(r, "user").(models.User)
	endpoint, err := models.GetEndpointByIDAndUser(endpointID, user.ID)
	if err != nil {
		c.SendError(w, r, "There is some error deleting endpoint!")
		return
	}

	containerID := dockerutil.GetClient().GetIDFromName(context.TODO(), endpoint.Label)

	if containerID != "" {
		// Run undeploy command only if the container is running
		cmdArgs := []string{
			"--module",
			"undeploy",
			"--label",
			endpoint.Label,
		}

		output, err := endpointcli.RunCmd(c.config.CliBinPath, cmdArgs)

		if err != nil {
			c.logger.Error("UnDeploy error: ", err)
			c.SendError(w, r, "Unable to undeploy the endpoint")
			return
		}

		if output == nil {
			c.logger.Error("UnDeploy error: output is nil", nil)
			c.SendError(w, r, "Unable to undeploy the endpoint")
			return
		}

		if !output.IsSuccess {
			if output.ErrorMessage == "" {
				c.SendError(w, r, "Unable to undeploy the endpoint")
				return
			} else {
				c.SendError(w, r, output.ErrorMessage)
				return

			}
		}
	}

	targetDirPath, _ := models.GetOpenApiFilePath(endpoint, c.config.WorkDir)
	// Check if the target directory exists before removing it
	if _, err := os.Stat(targetDirPath); err == nil {
		err := os.RemoveAll(targetDirPath)
		if err != nil {
			c.logger.Error("Unable to remove directories", err)
			c.SendError(w, r, "Unable to remove directories")
			return
		}
	}
	dockerDirPath := models.GetDockerStackDir(endpoint, c.config.WorkDir)
	// Check if the target directory exists before removing it
	if _, err := os.Stat(dockerDirPath); err == nil {

		err := os.RemoveAll(dockerDirPath)
		if err != nil {
			c.logger.Error("Unable to remove directories", err)
			c.SendError(w, r, "Unable to remove directories")
			return
		}
	}

	error := models.DeleteEndpoint(endpoint.ID)
	if error != nil {
		c.SendError(w, r, "There is some issue deleting endpoint!")
		return
	}

	c.SendSuccess(w, r, "Successfully deleted the endpoint")
}

// Update Request Mode for Endpoint
func (c *endpointsController) UpdateRequestMode(w http.ResponseWriter, r *http.Request) {
	inputID := chi.URLParam(r, "endpointID")

	requiredParams := []string{"request_mode"}
	if !utils.CheckAllParamsExist(r, requiredParams) {
		c.SendError(w, r, "Invalid Request")
		return
	}

	validModes := []string{"block", "monitor"}
	requestMode := r.PostFormValue("request_mode")

	isValid := false
	for _, mode := range validModes {
		if mode == requestMode {
			isValid = true
			break
		}
	}

	if !isValid {
		c.SendError(w, r, "Invalid Request Mode")
		return
	}

	endpointID, err := strconv.ParseUint(inputID, 10, 64)
	if err != nil {
		c.SendError(w, r, "Invalid Endpoint ID")
		return
	}

	user := ctx.Get(r, "user").(models.User)
	endpoint, err := models.GetEndpointByIDAndUser(endpointID, user.ID)

	if err != nil {
		c.SendError(w, r, "There is some error updating request mode!")
		return
	}

	if endpoint.RequestMode == requestMode {
		c.SendError(w, r, fmt.Sprintf("No change as request mode is already %s", requestMode))
		return
	}

	endpoint.RequestMode = requestMode
	save := models.SaveEndpoint(&endpoint)
	if save != nil {
		c.SendError(w, r, "Unable to update the Request mode")
		return
	}

	isSuccess, err := c.handleDeploy(endpoint, false)
	if isSuccess {
		c.SendSuccess(w, r, "Request Mode updated!")
		return
	} else {
		if err != nil {
			c.SendError(w, r, err.Error())
			return
		} else {
			c.SendError(w, r, "Unable to start the endpoint")
			return
		}
	}
}

// Download Config Routes
func (c *endpointsController) ListReconfig(w http.ResponseWriter, r *http.Request) {
	user := ctx.Get(r, "user").(models.User)
	data, err := models.GetEndpoints(&user)
	if err != nil {
		c.SendJSONError(w, err.Error())
	}
	templateData := views.NewTemplateData(c.config, c.session, r)
	templateData.Title = "Download Reconfig"
	templateData.Data = data
	if err := views.RenderTemplate(w, "endpoints/download-reconfig", templateData); err != nil {
		c.logger.Error("Error rendering template: ", err)
	}
}

// Download Reconfig Function
func (c *endpointsController) DownloadReconfig(w http.ResponseWriter, r *http.Request) {
	inputID := chi.URLParam(r, "endpointID")
	endpointID, err := strconv.ParseUint(inputID, 10, 64)
	if err != nil {
		c.FlashAndGoBack(w, r, flashtypes.FlashWarning, "Invalid Endpoint ID")
		return
	}

	user := ctx.Get(r, "user").(models.User)
	data, err := models.GetEndpointByIDAndUser(endpointID, user.ID)
	if err != nil {
		c.SendJSONError(w, err.Error())
		return
	}

	dockerDir := models.GetDockerLogsDir(data, c.config.WorkDir)
	targetDir, _ := models.GetOpenApiFilePath(data, c.config.WorkDir)
	zipFileName := fmt.Sprintf("%s_reconfig.zip", data.Label)
	c.zipFolderAndDownload(r, w, targetDir, dockerDir, zipFileName)
}

// Other required functions

// Handle zip creation for download of reconfig files
func (c *endpointsController) zipFolderAndDownload(r *http.Request, w http.ResponseWriter, openApiSource, firewallSource, zipFilename string) {
	// Create a buffer to store the zip file contents
	buf := new(bytes.Buffer)

	// Create a new zip writer using the buffer
	zipWriter := zip.NewWriter(buf)

	// Function to add files to zip
	addFileToZip := func(filePath, relativePath string) error {
		fileToZip, err := os.Open(filePath)
		if err != nil {
			return err
		}
		defer fileToZip.Close()

		// Create a new file in the zip archive
		fileInZip, err := zipWriter.Create(relativePath)
		if err != nil {
			return err
		}

		// Copy the file contents to the zip file
		_, err = io.Copy(fileInZip, fileToZip)
		return err
	}

	// Function to walk through directory and add files to zip
	walkAndAddFiles := func(sourceDir string) error {
		return filepath.Walk(sourceDir, func(filePath string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}

			if !info.IsDir() {
				// Get the relative path of the file
				relativePath, err := filepath.Rel(sourceDir, filePath)
				if err != nil {
					return err
				}

				// Add the file to the zip archive
				if err := addFileToZip(filePath, relativePath); err != nil {
					return err
				}
			}
			return nil
		})
	}

	// Add files from openApiSource
	if err := walkAndAddFiles(openApiSource); err != nil {
		c.FlashAndGoBack(w, r, flashtypes.FlashWarning, "Error Fetching Open Source Api files")
		return
	}

	// Add files from firewallSource
	if err := walkAndAddFiles(firewallSource); err != nil {
		c.FlashAndGoBack(w, r, flashtypes.FlashWarning, "Error Fetching Firewall Source files")
		return
	}

	// Close the zip writer
	if err := zipWriter.Close(); err != nil {
		c.FlashAndGoBack(w, r, flashtypes.FlashWarning, "Error Creating a zip file")
		return
	}

	// Set content type and attachment header for zip file
	w.Header().Set("Content-Type", "application/zip")
	w.Header().Set("Content-Disposition", "attachment; filename="+zipFilename)

	// Write the buffer contents to the response writer
	if _, err := io.Copy(w, buf); err != nil {
		c.FlashAndGoBack(w, r, flashtypes.FlashWarning, "Error Dowloading a zip file")
		return
	}

	defer os.Remove(zipFilename)
}

// Handle Endpoint Addition
func (c *endpointsController) handleEndpointAddition(w http.ResponseWriter, r *http.Request, tmpFileName string) {

	cmdArgs := []string{
		"--module",
		"validate",
		"--api-file",
		tmpFileName,
	}

	output, err := endpointcli.RunCmd(c.config.CliBinPath, cmdArgs)
	if err != nil {
		c.logger.Error("handleEndpointAddition error", err)
		c.SendJSONError(w, "Unable to validate the file")
		return
	}

	if output == nil {
		c.logger.Error("handleEndpointAddition error: output is nil", nil)
		c.SendJSONError(w, "Unable to validate the file")
		return
	}

	if !output.IsValid {
		if output.ErrorMessage == "" {
			c.SendJSONError(w, "Unable to validate the file")
			return
		} else {
			c.SendJSONError(w, output.ErrorMessage)
			return
		}
	}

	apiUrl := r.FormValue("api_url")
	label := strings.ToLower(r.FormValue("label"))
	portOption := r.FormValue("select_port_option")

	if err := urlutil.IsServerReachable(apiUrl); err != nil {
		c.SendJSONError(w, fmt.Sprintf("UnableToReach API Server: %s", err.Error()))
		return
	}

	var listeningPort int

	_, err = models.GetEndpointByLabel(label)
	if err == nil {
		c.SendJSONError(w, "Label already exists, please choose another one")
		return
	}

	if portOption == "custom" {
		portInput := r.FormValue("port_input")
		if portInput == "" {
			c.SendJSONError(w, "Invalid custom Port")
			return
		}

		customPort, err := strconv.Atoi(portInput)
		if err != nil {
			c.SendJSONError(w, "Invalid custom Port")
			return
		}

		if customPort <= models.MIN_CUSTOM_PORT || customPort >= models.MAX_CUSTOM_PORT {
			c.SendJSONError(w, fmt.Sprintf("Custom port is out of range, value between %d and %d is allowed", models.MIN_CUSTOM_PORT, models.MAX_CUSTOM_PORT))
			return
		}

		if models.CheckPortExist(customPort) {
			c.SendJSONError(w, fmt.Sprintf("Port %d already in use\n", customPort))
			return
		}

		listeningPort = customPort
	} else {
		listeningPort = models.GetAvailablePort()
		if listeningPort == 0 {
			c.SendJSONError(w, "No Port available right now, use custom port!")
			return
		}
	}

	if listeningPort == 0 {
		c.SendJSONError(w, "Invalid listening port")
	}

	healthPort := models.GetAvailableHealthPort()
	if healthPort == 0 {
		c.SendJSONError(w, "Unable to assign Health port")
		return
	}

	endpoint := models.Endpoint{
		Label:         label,
		ApiUrl:        apiUrl,
		ListeningPort: listeningPort,
		HealthPort:    healthPort,
		RequestMode:   models.DEFAULT_REQUEST_MODE,
		CreatedAt:     time.Now(),
		UserID:        ctx.Get(r, "user").(models.User).ID,
	}

	if err := models.SaveEndpoint(&endpoint); err != nil {
		c.logger.Error("HandleJobCreation:", err)
		c.SendJSONError(w, "Unable to add the Endpoint")
		return
	}

	targetDir, destinationFile := models.GetOpenApiFilePath(endpoint, c.config.WorkDir)

	if err := os.MkdirAll(targetDir, os.ModePerm); err != nil {
		c.logger.Error("HandleJobCreation:", err)
		c.SendJSONError(w, "Unable to create target directory")
		return
	}

	if err := os.Rename(tmpFileName, destinationFile); err != nil {
		_ = os.RemoveAll(targetDir)
		c.logger.Error("HandleJobCreation:", err)
		c.SendJSONError(w, "Unable to upload the yaml file")
		return
	}

	isSuccess, err := c.handleDeploy(endpoint, true)
	if !isSuccess {
		_ = os.RemoveAll(targetDir)
		if err != nil {
			c.SendJSONError(w, err.Error())
			return
		} else {
			c.SendJSONError(w, "Unable to deploy the endpoint")
			return
		}
	}

	c.SendJSONSuccess(w, "Successfully added the Api Protector")
}

func (c *endpointsController) DenyTokensPage(w http.ResponseWriter, r *http.Request) {
	templateData := views.NewTemplateData(c.config, c.session, r)

	user := ctx.Get(r, "user").(models.User)
	data, err := models.GetEndpoints(&user)
	if err != nil {
		c.SendJSONError(w, err.Error())
	}
	templateData.Title = "Endpoints"
	templateData.Data = data
	if err := views.RenderTemplate(w, "endpoints/denied-tokens/index", templateData); err != nil {
		c.logger.Error("Error rendering template: ", err)
	}

}

func (c *endpointsController) AllowedIPsPage(w http.ResponseWriter, r *http.Request) {
	templateData := views.NewTemplateData(c.config, c.session, r)
	user := ctx.Get(r, "user").(models.User)
	data, err := models.GetEndpoints(&user)
	if err != nil {
		c.SendJSONError(w, err.Error())
	}
	templateData.Title = "Endpoints"
	templateData.Data = data
	if err := views.RenderTemplate(w, "endpoints/allowed-ips/index", templateData); err != nil {
		c.logger.Error("Error rendering template: ", err)
	}
}

// Ip Header Route
func (c *endpointsController) ListIPHeader(w http.ResponseWriter, r *http.Request) {
	user := ctx.Get(r, "user").(models.User)

	data, err := models.GetEndpointListWithConfig(&user, models.CLIENT_IP_HEADER)
	if err != nil {
		c.SendJSONError(w, err.Error())
	}
	templateData := views.NewTemplateData(c.config, c.session, r)
	templateData.Title = "IP Header"
	templateData.Data = ConfigTmplData{data, models.IPHeaders, models.CLIENT_IP_HEADER}
	if err := views.RenderTemplate(w, "endpoints/configs/ip-header", templateData); err != nil {
		c.logger.Error("Error rendering template: ", err)
	}
}

// Token Header Route
func (c *endpointsController) ListTokenHeader(w http.ResponseWriter, r *http.Request) {
	user := ctx.Get(r, "user").(models.User)
	data, err := models.GetEndpointListWithConfig(&user, models.DENY_TOKEN_HEADER)
	if err != nil {
		c.SendJSONError(w, err.Error())
	}
	templateData := views.NewTemplateData(c.config, c.session, r)
	templateData.Title = "Deny Token Header"
	templateData.Data = ConfigTmplData{data, models.TokenHeaders, models.DENY_TOKEN_HEADER}
	if err := views.RenderTemplate(w, "endpoints/configs/token-header", templateData); err != nil {
		c.logger.Error("Error rendering template: ", err)
	}
}
