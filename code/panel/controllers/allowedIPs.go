package controllers

import (
	"bufio"
	"errors"
	"fmt"
	"net/http"
	"path/filepath"
	"slices"
	"strconv"
	"strings"

	ctx "github.com/CSPF-Founder/api-firewall-panel/api-protector/onpremise/panel/context"
	"github.com/CSPF-Founder/api-firewall-panel/api-protector/onpremise/panel/enums/flashtypes"
	"github.com/CSPF-Founder/api-firewall-panel/api-protector/onpremise/panel/internal/endpointcli"
	"github.com/CSPF-Founder/api-firewall-panel/api-protector/onpremise/panel/models"
	"github.com/CSPF-Founder/api-firewall-panel/api-protector/onpremise/panel/pkg/iputil"
	"github.com/CSPF-Founder/api-firewall-panel/api-protector/onpremise/panel/views"
	"github.com/go-chi/chi/v5"
)

var allowedIPListExt = []string{"txt"}

type allowedIPController struct {
	*App
}

func newAllowedIPController(app *App) *allowedIPController {
	return &allowedIPController{
		App: app,
	}
}

// RegisterRoutes registers the routes of allowedIPController
func (c *allowedIPController) registerRoutes() http.Handler {
	r := chi.NewRouter()
	r.Get("/", c.AllowedIPsByEndpoint)
	r.Post("/", c.AddAllowedIPHandler)
	r.Post("/import", c.ImportAllowedIPHandler)
	r.Delete("/{entryID}", c.DeleteAllowedIPHandler)
	r.Patch("/{entryID}", c.UpdateAllowedIPHandler)

	return r
}

func (c *allowedIPController) handleDeploy(endpoint models.Endpoint, newDeployment bool) (bool, error) {
	err := endpointcli.HandleDeploy(c.config.CliBinPath, c.config.WorkDir, endpoint, newDeployment)
	if err != nil {
		c.logger.Error("handleDeploy error", err)
		return false, errors.New("Unable to deploy")
	}
	return true, nil
}

func (c *allowedIPController) AllowedIPsByEndpoint(w http.ResponseWriter, r *http.Request) {
	inputID := chi.URLParam(r, "endpointID")
	endpointID, err := strconv.ParseUint(inputID, 10, 64)
	if err != nil {
		c.FlashAndGoBack(w, r, flashtypes.FlashWarning, "Invalid Endpoint ID")
		return
	}

	user := ctx.Get(r, "user").(models.User)
	endpoint, err := models.GetEndpointByIDAndUser(endpointID, user.ID)
	if err != nil {
		c.FlashAndGoBack(w, r, flashtypes.FlashWarning, "Invalid Endpoint ID")
		return
	}

	ipHeader, _ := models.GetConfigByEndpointIDAndKey(endpoint.ID, models.CLIENT_IP_HEADER)
	if ipHeader.ConfigValue == "" {
		c.Flash(w, r, flashtypes.FlashWarning, "IP Header not set for the endpoint. If it is intended, ignore this warning", false)
		c.Flash(w, r, flashtypes.FlashWarning, "Note: Without the IP Header set, it will default to the client IP from the HTTP request", false)
	}

	allowedIPs, err := models.GetAllowedIPByEndpointID(endpoint.ID)
	if err != nil {
		c.FlashAndGoBack(w, r, flashtypes.FlashWarning, "Error Fetching allowed ips")
		return
	}

	templateData := views.NewTemplateData(c.config, c.session, r)
	templateData.Title = fmt.Sprintf("Allowed IP List (%s)", endpoint.Label)
	templateData.Data = struct {
		Endpoint   models.Endpoint
		AllowedIPs []models.AllowedIP
	}{
		Endpoint:   endpoint,
		AllowedIPs: allowedIPs,
	}
	if err := views.RenderTemplate(w, "endpoints/allowed-ips/list", templateData); err != nil {
		c.logger.Error("Error rendering template: ", err)
		c.FlashAndGoBack(w, r, flashtypes.FlashWarning, "Unknown Error Occurred")
	}
}

// Import allowed ip's from file
func (c *allowedIPController) ImportAllowedIPHandler(w http.ResponseWriter, r *http.Request) {
	inputEndpointID := chi.URLParam(r, "endpointID")
	err := r.ParseMultipartForm(10 << 20)
	if err != nil {
		c.FlashAndGoBack(w, r, flashtypes.FlashWarning, "Invalid file")
		return
	}

	file, handler, err := r.FormFile("text_file")
	if err != nil {
		c.FlashAndGoBack(w, r, flashtypes.FlashWarning, "Invalid file")
		return
	}
	defer file.Close()

	fileName := handler.Filename
	fileType := filepath.Ext(fileName)[1:]

	if !slices.Contains(allowedIPListExt, fileType) {
		c.FlashAndGoBack(w, r, flashtypes.FlashWarning, "Invalid file format")
		return
	}

	fileSize := handler.Size
	// filesize max 10 mb
	if fileSize > 10000000 {
		c.FlashAndGoBack(w, r, flashtypes.FlashWarning, "File size should be less than 10 MB")
		return
	}

	restart := r.PostFormValue("restart")

	endpointID, err := strconv.ParseUint(inputEndpointID, 10, 64)
	if err != nil {
		c.FlashAndGoBack(w, r, flashtypes.FlashWarning, "Invalid Endpoint ID")
		return
	}

	user := ctx.Get(r, "user").(models.User)
	endpoint, err := models.GetEndpointByIDAndUser(endpointID, user.ID)
	if err != nil {
		c.FlashAndGoBack(w, r, flashtypes.FlashWarning, "Invalid Endpoint ID")
		return
	}

	var entries []string

	// Create a scanner to read the file line by line
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		// Check if the line is not empty after trimming whitespace
		if trimmed := strings.TrimSpace(line); trimmed != "" {
			entries = append(entries, trimmed)
		}
	}

	// Check for any errors during scanning
	if err := scanner.Err(); err != nil {
		c.FlashAndGoBack(w, r, flashtypes.FlashWarning, "Error reading File")
		return
	}

	addedEntries := 0
	existingEntries := 0
	totalCount := len(entries)
	invalidCount := 0

	// Iterate over array and execute insert statement
	if totalCount == 0 {
		c.FlashAndGoBack(w, r, flashtypes.FlashWarning, "No IPs Provided in the file")
		return
	}

	for _, entry := range entries {
		if !models.CheckAllowedIPExist(entry, endpoint.ID) {
			isValid := false
			isRange := 0
			if iputil.IsValidIP(w, entry) {
				isValid = true
			} else if iputil.IsValidIPRange(w, entry) {
				isValid = true
				isRange = 1
			}

			if isValid {
				allowedIPData := models.AllowedIP{EndpointID: endpoint.ID, IPData: entry, IsRange: isRange}
				err = models.SaveAllowedIP(&allowedIPData)
				if err == nil {
					addedEntries++
				}
			} else {
				invalidCount++
			}
		} else {
			existingEntries++
		}
	}

	if addedEntries == 0 {
		c.FlashAndGoBack(w, r, flashtypes.FlashWarning, "No IPs Added")
		return
	}

	if restart == "1" {
		isSuccess, err := c.handleDeploy(endpoint, false)
		if !isSuccess {
			if err != nil {
				c.logger.Error("ImportAllowedIPHandler", err)
				c.FlashAndGoBack(w, r, flashtypes.FlashWarning, "Saved the entries but unable to restart the endpoint.")
				return
			} else {
				c.FlashAndGoBack(w, r, flashtypes.FlashWarning, "Saved the entries but unable to restart the endpoint.")
				return
			}
		}
	}

	msg := fmt.Sprintf("Imported IPs: %d", addedEntries)

	msg = fmt.Sprintf("%s, Existing IPs: %d", msg, existingEntries)
	if invalidCount > 0 {
		msg = fmt.Sprintf("%s, Invalid IPs: %d", msg, invalidCount)
	}

	c.FlashAndGoBack(w, r, flashtypes.FlashSuccess, msg)
}

// Add Single Ip
func (c *allowedIPController) AddAllowedIPHandler(w http.ResponseWriter, r *http.Request) {

	inputID := chi.URLParam(r, "endpointID")

	allowedIP := r.PostFormValue("allowed_ip")
	restart := r.PostFormValue("restart")

	if allowedIP == "" {
		c.SendError(w, r, "No IP Provided")
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
		c.SendError(w, r, "Invalid Endpoint ID")
		return
	}

	if models.CheckAllowedIPExist(allowedIP, endpoint.ID) {
		c.SendError(w, r, "IP Already Exists")
		return
	}

	isValid := false
	isRange := 0
	if iputil.IsValidIP(w, allowedIP) {
		isValid = true
	} else if iputil.IsValidIPRange(w, allowedIP) {
		isValid = true
		isRange = 1
	}

	if !isValid {
		c.SendError(w, r, "Invalid IP")
		return
	}

	allowedIPData := models.AllowedIP{EndpointID: endpoint.ID, IPData: allowedIP, IsRange: isRange}
	dt_err := models.SaveAllowedIP(&allowedIPData)
	if dt_err != nil {
		c.SendError(w, r, "Error Adding IP")
		return
	}

	if restart == "1" {
		isSuccess, err := c.handleDeploy(endpoint, false)
		if !isSuccess {
			if err != nil {
				c.logger.Error("AddAllowedIPHandler", err)
				c.SendError(w, r, "Saved the entry but unable to restart the endpoint.")
				return
			} else {
				c.SendError(w, r, "Saved the entry but unable to restart the endpoint.")
				return
			}
		}
	}

	c.SendSuccess(w, r, "IP/IP Range added succesfully")
}

// Update Ip/Ip range route
func (c *allowedIPController) UpdateAllowedIPHandler(w http.ResponseWriter, r *http.Request) {
	inputEndpointID := chi.URLParam(r, "endpointID")
	entryID := chi.URLParam(r, "entryID")

	allowedIP := r.PostFormValue("allowed_ip")
	restart := r.PostFormValue("restart")

	if allowedIP == "" {
		c.SendError(w, r, "No IP Provided")
		return
	}
	endpointID, err := strconv.ParseUint(inputEndpointID, 10, 64)
	if err != nil {
		c.SendError(w, r, "Invalid Endpoint ID")
		return
	}

	dbID, err := strconv.ParseInt(entryID, 10, 64)
	if err != nil {
		c.SendError(w, r, "Invalid IP ID parameter")
		return
	}

	user := ctx.Get(r, "user").(models.User)
	endpoint, err := models.GetEndpointByIDAndUser(endpointID, user.ID)
	if err != nil {
		c.SendError(w, r, "Invalid Endpoint ID")
		return
	}

	isValid := false
	isRange := 0
	if iputil.IsValidIP(w, allowedIP) {
		isValid = true
	} else if iputil.IsValidIPRange(w, allowedIP) {
		isValid = true
		isRange = 1
	}

	if !isValid {
		c.SendError(w, r, "Invalid IP")
		return
	}

	dt_err := models.SaveAllowedIP(&models.AllowedIP{ID: dbID, EndpointID: endpoint.ID, IPData: allowedIP, IsRange: isRange})
	if dt_err != nil {
		c.SendError(w, r, "Error Updating IP Data")
		return
	}

	if restart == "1" {
		isSuccess, err := c.handleDeploy(endpoint, false)
		if !isSuccess {
			if err != nil {
				c.logger.Error("UpdateAllowedIPHandler", err)
				c.SendError(w, r, "Updated the entry but unable to restart the endpoint.")
				return
			} else {
				c.SendError(w, r, "Updated the entry but unable to restart the endpoint.")
				return
			}
		}
	}

	c.SendSuccess(w, r, "IP Updated succesfully")
}

// Delete Ip
func (c *allowedIPController) DeleteAllowedIPHandler(w http.ResponseWriter, r *http.Request) {
	inputEndpointID := chi.URLParam(r, "endpointID")
	entryID := chi.URLParam(r, "entryID")

	resp := DeleteResponse{
		Error:     "",
		IsRemoved: false,
	}

	endpointID, err := strconv.ParseUint(inputEndpointID, 10, 64)
	if err != nil {
		resp.Error = "Invalid Endpoint ID"
		c.sendDeleteResponse(w, resp)
		return
	}

	user := ctx.Get(r, "user").(models.User)
	endpoint, err := models.GetEndpointByIDAndUser(endpointID, user.ID)
	if err != nil {
		resp.Error = "Invalid Endpoint ID"
		c.sendDeleteResponse(w, resp)
		return
	}

	dbID, err := strconv.ParseInt(entryID, 10, 64)
	if err != nil {
		resp.Error = "Invalid ID to delete"
		c.sendDeleteResponse(w, resp)
		return
	}

	err = models.DeleteAllowedIP(endpoint.ID, dbID)
	if err != nil {
		resp.Error = "Error Removing IP/IP Range"
		c.sendDeleteResponse(w, resp)
		return
	}

	resp.IsRemoved = true

	isSuccess, err := c.handleDeploy(endpoint, false)
	if isSuccess {
		c.sendDeleteResponse(w, resp)
		return
	} else {
		if err != nil {
			c.logger.Error("DeleteAllowIPList", err)
			return
		}

		resp.Error = "Removed the entry but unable to restart the endpoint."
		c.sendDeleteResponse(w, resp)
		return
	}

}
