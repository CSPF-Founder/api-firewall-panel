package controllers

import (
	"errors"
	"fmt"
	"net/http"
	"slices"
	"strconv"

	ctx "github.com/CSPF-Founder/api-firewall-panel/api-protector/onpremise/panel/context"
	"github.com/CSPF-Founder/api-firewall-panel/api-protector/onpremise/panel/internal/endpointcli"
	"github.com/CSPF-Founder/api-firewall-panel/api-protector/onpremise/panel/models"
	"github.com/CSPF-Founder/api-firewall-panel/api-protector/onpremise/panel/utils"
	"github.com/go-chi/chi/v5"
)

type ConfigTmplData struct {
	Endpoints  []models.Endpoint
	Headers    []string
	HeaderType string
}

type endpointConfigController struct {
	*App
}

func newEndpointConfigController(app *App) *endpointConfigController {
	return &endpointConfigController{
		App: app,
	}
}

// registerRoutes registers all routes for deniedTokenController
func (c *endpointConfigController) registerRoutes() http.Handler {
	r := chi.NewRouter()
	r.Post("/", c.AddConfigHandler)
	r.Patch("/", c.UpdateConfigHandler)
	r.Delete("/{config_key}", c.DeleteConfigHandler)

	return r
}

func (c *endpointConfigController) handleDeploy(endpoint models.Endpoint, newDeployment bool) (bool, error) {
	err := endpointcli.HandleDeploy(c.config.CliBinPath, c.config.WorkDir, endpoint, newDeployment)
	if err != nil {
		c.logger.Error("handleDeploy error", err)
		return false, errors.New("Unable to deploy")
	}
	return true, nil
}

// AddConfigHandler adds a new configuration for an endpoint
func (c *endpointConfigController) AddConfigHandler(w http.ResponseWriter, r *http.Request) {
	inputEndpointID := chi.URLParam(r, "endpointID")

	requiredParams := []string{"config_key", "config_value"}
	if !utils.CheckAllParamsExist(r, requiredParams) {
		c.SendError(w, r, "One or more required fields missing")
		return
	}

	restart := r.PostFormValue("restart")

	configName := r.PostFormValue("config_key")
	configInput := r.PostFormValue("config_value")
	customValue := r.PostFormValue("custom_value") // applicable for custom headers

	if !slices.Contains(models.ConfigKeys, configName) {
		c.SendError(w, r, "Invalid Request!")
		return
	}

	configValue := configInput
	// if custom header is selected
	if configInput == "custom" {
		if customValue == "" {
			c.SendError(w, r, "Custom Header not Provided!")
			return
		}
		configValue = customValue
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

	config := models.EndpointConfig{
		EndpointID:  endpoint.ID,
		ConfigKey:   configName,
		ConfigValue: configValue,
	}

	se_err := models.SaveEndpointConfig(&config)
	if se_err != nil {
		c.SendError(w, r, fmt.Sprintf("Error Adding %s", transformString(configName)))
		return
	}

	if restart == "1" {
		isSuccess, err := c.handleDeploy(endpoint, false)
		if !isSuccess {
			if err != nil {
				c.SendError(w, r, err.Error())
				return
			} else {
				c.SendError(w, r, "Unable to Deploy Configuration")
				return
			}
		}
	}

	c.SendSuccess(w, r, fmt.Sprintf("%s Created for Endpoint: %s", transformString(configName), endpoint.Label))
}

// UpdateConfigHandler updates a configuration for an endpoint
func (c *endpointConfigController) UpdateConfigHandler(w http.ResponseWriter, r *http.Request) {
	inputEndpointID := chi.URLParam(r, "endpointID")
	requiredParams := []string{"config_key", "config_value"}
	if !utils.CheckAllParamsExist(r, requiredParams) {
		c.SendError(w, r, "One or more required fields missing")
		return
	}
	restart := r.PostFormValue("restart")
	configName := r.PostFormValue("config_key")
	configInput := r.PostFormValue("config_value")
	customValue := r.PostFormValue("custom_value") // applicable for custom headers

	if !slices.Contains(models.ConfigKeys, configName) {
		c.SendError(w, r, "Invalid Request!")
		return
	}

	configValue := configInput
	// if custom header is selected
	if configInput == "custom" {
		if customValue == "" {
			c.SendError(w, r, "Custom Header not Provided!")
			return
		}
		configValue = customValue
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

	endpoint_config, err := models.GetConfigByEndpointIDAndKey(endpointID, configName)
	if err != nil {
		c.SendError(w, r, fmt.Sprintf("%s doesn't exist for Endpoint: %s", transformString(configName), endpoint.Label))
		return
	}

	endpoint_config.ConfigValue = configValue

	se_err := models.SaveEndpointConfig(&endpoint_config)
	if se_err != nil {
		c.SendError(w, r, fmt.Sprintf("Error Updating %s", transformString(configName)))
		return
	}

	if restart == "1" {
		isSuccess, err := c.handleDeploy(endpoint, false)
		if !isSuccess {
			if err != nil {
				c.SendError(w, r, err.Error())
				return
			} else {
				c.SendError(w, r, "Unable to Deploy Configuration")
				return
			}
		}
	}

	c.SendSuccess(w, r, fmt.Sprintf("%s updated for Endpoint: %s", transformString(configName), endpoint.Label))
}

// DeleteConfigHandler deletes a configuration for an endpoint
func (c *endpointConfigController) DeleteConfigHandler(w http.ResponseWriter, r *http.Request) {
	inputEndpointID := chi.URLParam(r, "endpointID")
	configKey := chi.URLParam(r, "config_key")

	if !slices.Contains(models.ConfigKeys, configKey) {
		c.SendError(w, r, "Invalid Request!")
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

	endpoint_config, err := models.GetConfigByEndpointIDAndKey(endpointID, configKey)
	if err != nil {
		c.SendError(w, r, fmt.Sprintf("%s doesn't exist for Endpoint: %s", transformString(configKey), endpoint.Label))
		return
	}

	se_err := models.DeleteEndpointConfig(endpoint_config.ID, configKey)
	if se_err != nil {
		c.SendError(w, r, fmt.Sprintf("Error Removing %s", transformString(configKey)))
		return
	}

	isSuccess, err := c.handleDeploy(endpoint, false)
	if isSuccess {
		c.SendSuccess(w, r, fmt.Sprintf("%s removed for Endpoint: %s", transformString(configKey), endpoint.Label))
		return
	} else {
		if err != nil {
			c.SendError(w, r, err.Error())
			return
		} else {
			c.SendError(w, r, "Unable to Deploy Configuration")
			return
		}
	}
}
