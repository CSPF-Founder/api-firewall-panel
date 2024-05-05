package controllers

import "net/http"

type DeleteResponse struct {
	Error     string `json:"error"`
	IsRemoved bool   `json:"is_removed"`
}

func (c *App) sendDeleteResponse(w http.ResponseWriter, resp DeleteResponse) {
	if !resp.IsRemoved || resp.Error != "" {
		c.JSONResponse(w, resp, http.StatusUnprocessableEntity)
		return
	}

	c.JSONResponse(w, resp, http.StatusOK)
}
