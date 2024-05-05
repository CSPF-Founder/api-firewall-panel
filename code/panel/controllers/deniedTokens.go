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
	"github.com/CSPF-Founder/api-firewall-panel/api-protector/onpremise/panel/views"
	"github.com/go-chi/chi/v5"
)

type deniedTokenController struct {
	*App
}

func newDeniedTokenController(app *App) *deniedTokenController {
	return &deniedTokenController{
		App: app,
	}
}

// registerRoutes registers all routes for deniedTokenController
func (c *deniedTokenController) registerRoutes() http.Handler {
	r := chi.NewRouter()
	r.Get("/", c.DenyTokensByEndpoint)
	r.Post("/", c.AddDenyTokenHandler)
	r.Post("/import", c.ImportDeniedTokenHandler)
	r.Patch("/{entryID}", c.UpdateDenyTokenHandler)
	r.Delete("/{entryID}", c.DeleteDenyTokenHandler)

	return r
}

// Handle Deploy function
func (c *deniedTokenController) handleDeploy(endpoint models.Endpoint, newDeployment bool) (bool, error) {
	err := endpointcli.HandleDeploy(c.config.CliBinPath, c.config.WorkDir, endpoint, newDeployment)
	if err != nil {
		c.logger.Error("handleDeploy error", err)
		return false, errors.New("Unable to deploy")
	}
	return true, nil
}

// Denytoken all Routes

// Deny Tokens Route
func (c *deniedTokenController) DenyTokensByEndpoint(w http.ResponseWriter, r *http.Request) {
	param_id := chi.URLParam(r, "endpointID")
	endpointID, err := strconv.ParseUint(param_id, 10, 64)
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

	denyTokenHeader, _ := models.GetConfigByEndpointIDAndKey(endpoint.ID, models.DENY_TOKEN_HEADER)
	if denyTokenHeader.ConfigValue == "" {
		c.Flash(w, r, flashtypes.FlashWarning, "Deny Token Header not set for the endpoint", false)
		c.Flash(w, r, flashtypes.FlashInfo, "Deny Token will not work without setting the header", false)
	}

	deniedTokens, err := models.GetDeniedTokensByEndpointID(endpoint.ID)
	if err != nil {
		c.FlashAndGoBack(w, r, flashtypes.FlashWarning, "Error Fetching Denied Tokens")
		return
	}

	templateData := views.NewTemplateData(c.config, c.session, r)

	templateData.Title = fmt.Sprintf("Denied Tokens (%s)", endpoint.Label)
	templateData.Data = struct {
		Endpoint     models.Endpoint
		DeniedTokens []models.DeniedToken
	}{
		Endpoint:     endpoint,
		DeniedTokens: deniedTokens,
	}

	if err := views.RenderTemplate(w, "endpoints/denied-tokens/list", templateData); err != nil {
		c.logger.Error("Error rendering template: ", err)
	}
}

// Import DenyToken route
func (c *deniedTokenController) ImportDeniedTokenHandler(w http.ResponseWriter, r *http.Request) {
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
		c.FlashAndGoBack(w, r, flashtypes.FlashWarning, "Error Reading File")
		return
	}

	addedEntries := 0
	existingEntries := 0
	totalCount := len(entries)
	invalidCount := 0

	// Iterate over array and execute insert statement
	if totalCount == 0 {
		c.FlashAndGoBack(w, r, flashtypes.FlashWarning, "No Tokens Provided in the file")
		return
	}

	for _, entry := range entries {
		if !models.CheckDeniedTokenExist(entry, endpoint.ID) {
			err := models.SaveDeniedToken(&models.DeniedToken{EndpointID: endpoint.ID, Token: entry})
			if err == nil {
				addedEntries++
			}
		} else {
			existingEntries++
		}
	}

	if addedEntries == 0 {
		c.FlashAndGoBack(w, r, flashtypes.FlashWarning, "No Tokens Added")
		return
	}

	if restart == "1" {
		isSuccess, err := c.handleDeploy(endpoint, false)
		if !isSuccess {
			if err != nil {
				c.logger.Error("ImportDenyTokens", err)
				c.FlashAndGoBack(w, r, flashtypes.FlashWarning, "Saved the entries but unable to restart the endpoint.")
				return
			} else {
				c.FlashAndGoBack(w, r, flashtypes.FlashWarning, "Saved the entries but unable to restart the endpoint.")
				return
			}
		}
	}

	msg := fmt.Sprintf("Imported Tokens: %d", addedEntries)

	msg = fmt.Sprintf("%s, Existing Tokens: %d", msg, existingEntries)
	if invalidCount > 0 {
		msg = fmt.Sprintf("%s, Invalid Tokens: %d", msg, invalidCount)
	}

	c.FlashAndGoBack(w, r, flashtypes.FlashSuccess, msg)

}

// Add Deny Token
func (c *deniedTokenController) AddDenyTokenHandler(w http.ResponseWriter, r *http.Request) {
	inputEndpointID := chi.URLParam(r, "endpointID")

	denyToken := r.PostFormValue("denied_token")
	restart := r.PostFormValue("restart")

	if denyToken == "" {
		c.SendError(w, r, "No Token Provided")
		return
	}

	endpointID, err := strconv.ParseUint(inputEndpointID, 10, 64)
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

	if models.CheckDeniedTokenExist(denyToken, endpoint.ID) {
		c.SendError(w, r, fmt.Sprintf("Token Already Exists: %s", denyToken))
		return
	}

	dt_err := models.SaveDeniedToken(&models.DeniedToken{EndpointID: endpoint.ID, Token: denyToken})
	if dt_err != nil {
		c.SendError(w, r, fmt.Sprintf("Error Adding Token: %s", denyToken))
		return
	}

	if restart == "1" {
		isSuccess, err := c.handleDeploy(endpoint, false)
		if !isSuccess {
			if err != nil {
				// c.SendError(w, r, err.Error())
				c.logger.Error("AddDenyTokenHandler", err)
				c.SendError(w, r, "Added the entry but unable to restart the endpoint.")
				return
			} else {
				c.SendError(w, r, "Added the entry but unable to restart the endpoint.")
				return
			}
		}
	}

	c.SendSuccess(w, r, "Denied Token Inserted Succesfully")
}

// Update Deny Token
func (c *deniedTokenController) UpdateDenyTokenHandler(w http.ResponseWriter, r *http.Request) {
	inputEndpointID := chi.URLParam(r, "endpointID")
	entryID := chi.URLParam(r, "entryID")

	denyToken := r.PostFormValue("denied_token")
	restart := r.PostFormValue("restart")

	if denyToken == "" {
		c.SendError(w, r, "No Token Provided")
		return
	}
	endpointID, err := strconv.ParseUint(inputEndpointID, 10, 64)
	if err != nil {
		c.SendError(w, r, "Invalid Endpoint ID")
		return
	}
	denyTokenID, err := strconv.ParseUint(entryID, 10, 64)
	if err != nil {
		c.SendError(w, r, "Invalid Denied Token ID parameter")
		return
	}

	user := ctx.Get(r, "user").(models.User)
	endpoint, err := models.GetEndpointByIDAndUser(endpointID, user.ID)
	if err != nil {
		c.SendError(w, r, "Invalid Endpoint ID")
		return
	}

	dt_err := models.SaveDeniedToken(&models.DeniedToken{ID: denyTokenID, EndpointID: endpoint.ID, Token: denyToken})
	if dt_err != nil {
		c.SendError(w, r, "Error Updating Token")
		return
	}

	if restart == "1" {
		isSuccess, err := c.handleDeploy(endpoint, false)
		if !isSuccess {
			if err != nil {
				c.logger.Error("UpdateDenyTokenHandler", err)
				c.SendError(w, r, "Updated the entry but unable to restart the endpoint.")
				return
			} else {
				c.SendError(w, r, "Updated the entry but unable to restart the endpoint.")
				return
			}
		}
	}

	c.SendSuccess(w, r, "Denied Token Updated Succesfully")
}

// Delete Deny Token
func (c *deniedTokenController) DeleteDenyTokenHandler(w http.ResponseWriter, r *http.Request) {
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

	dbID, err := strconv.ParseInt(entryID, 10, 64)
	if err != nil {
		resp.Error = "Invalid Denied Token ID parameter"
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

	err = models.DeleteDenyToken(endpoint.ID, dbID)
	if err != nil {
		resp.Error = "Error Removing Denied Token"
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
			c.logger.Error("DeleteDenyTokenHandler", err)
		}

		resp.Error = "Removed the entry but unable to restart the endpoint."
		c.sendDeleteResponse(w, resp)
		return
	}
}
