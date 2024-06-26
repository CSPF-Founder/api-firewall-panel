package controllers

import (
	"encoding/json"
	"net/http"
)

func (app *App) JSONResponse(w http.ResponseWriter, data any, c int) {
	jsonString, err := json.Marshal(data)
	if err != nil {
		http.Error(w, "Error creating JSON response", http.StatusInternalServerError)
		app.logger.Error("Error writing JSON response", err)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	// Respond with the JSON data
	w.WriteHeader(c)
	if _, err := w.Write(jsonString); err != nil {
		app.logger.Error("Error writing JSON response", err)
	}
}

// SendJSONError displays JSON error messages and exits the program.
func (app *App) SendJSONError(w http.ResponseWriter, message string, statusCodeOption ...int) {
	if message == "" {
		message = "Error Occurred"
	}

	statusCode := http.StatusUnprocessableEntity
	if len(statusCodeOption) > 0 {
		statusCode = statusCodeOption[0]
	}

	jsonOutput := map[string]interface{}{"error": message}
	app.JSONResponse(w, jsonOutput, statusCode)
}

// SendJSONSuccess displays JSON success messages and exits the program.
func (app *App) SendJSONSuccess(w http.ResponseWriter, message interface{}, statusCodeOption ...int) {
	if message == "" {
		message = "Error Occurred"
	}

	statusCode := http.StatusOK
	if len(statusCodeOption) > 0 {
		statusCode = statusCodeOption[0]
	}

	var jsonOutput map[string]interface{}

	//Check if message is string or map, and output json accrodingly
	//Implemented this to pass multiple JsonSuccess data
	switch v := message.(type) {
	case string:
		jsonOutput = map[string]interface{}{"success": message}
	case map[string]interface{}:
		jsonOutput = v
	default:
		jsonOutput = map[string]interface{}{"success": message}
	}

	app.JSONResponse(w, jsonOutput, statusCode)
}
