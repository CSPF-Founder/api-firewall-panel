package controllers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"strings"
	"testing"

	"github.com/CSPF-Founder/api-firewall-panel/api-protector/onpremise/panel/models"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/PuerkitoBio/goquery"
)

// Tests for the Endpoints controller

func (ctx *testContext) mockGetEndpointByID(endpoint models.Endpoint) {
	endpointRow := sqlmock.NewRows([]string{"id", "label", "request_mode", "created_at", "user_id"}).
		AddRow(endpoint.ID, endpoint.Label, endpoint.RequestMode, endpoint.CreatedAt, endpoint.UserID)

	ctx.mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM `endpoints` WHERE id = ? ORDER BY `endpoints`.`id` LIMIT 1")).
		WithArgs(endpoint.ID).
		WillReturnRows(endpointRow)
}

func (ctx *testContext) mockEmptyGetEndpointByID(job models.Endpoint) {
	endpointRow := sqlmock.NewRows([]string{"id", "label", "request_mode", "created_at", "user_id"})

	ctx.mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM `endpoints` WHERE id = ? ORDER BY `endpoints`.`id` LIMIT 1")).
		WithArgs(job.ID).
		WillReturnRows(endpointRow)
}

func TestDeleteEndpointWithMissingCSRF(t *testing.T) {
	testUser := models.User{
		ID:       1,
		Username: "test",
		Email:    "test@example.com",
	}
	ctx, resp := loggedSessionForTest(t, testUser)

	testEndpoint := models.Endpoint{
		ID:     1,
		UserID: 1,
	}

	body := struct {
		Key uint64 `json:"endpoint_delete_id"`
	}{
		Key: testEndpoint.ID,
	}

	out, err := json.Marshal(body)
	if err != nil {
		t.Fatalf("error Creating json for request body: %v", err)
	}

	ctx.mockGetByUserID(testUser)

	deleteURL := fmt.Sprintf("%s/endpoints/delete", ctx.server.URL)
	req, err := http.NewRequest("POST", deleteURL, bytes.NewBuffer(out))
	if err != nil {
		t.Fatalf("error creating new /users/login request: %v", err)
	}

	req.Header.Set("Cookie", resp.Header.Get("Set-Cookie"))

	client := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	resp, err = client.Do(req)
	if err != nil {
		t.Fatalf("error requesting the /users/login endpoint: %v", err)
	}
	got := resp.StatusCode
	expected := http.StatusForbidden
	if got != expected {
		t.Fatalf("invalid status code received. expected %d got %d", expected, got)
	}

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		t.Fatalf("error parsing /login response body")
	}

	if !strings.Contains(doc.Text(), InvalidCSRFTokenError) {
		t.Fatalf("Expected %s in response body but not found", InvalidCSRFTokenError)
	}

}

func TestDeleteInvalidEndpointID(t *testing.T) {
	testUser := models.User{
		ID:       1,
		Username: "test",
		Email:    "test@example.com",
	}
	ctx, resp := loggedSessionForTest(t, testUser)

	endpointJob := models.Endpoint{
		ID:     1,
		UserID: 1,
	}

	ctx.mockGetByUserID(testUser)
	ctx.mockEmptyGetEndpointByID(endpointJob)

	deleteURL := fmt.Sprintf("%s/endpoints/%d", ctx.server.URL, endpointJob.ID)
	req, err := http.NewRequest("DELETE", deleteURL, nil)
	if err != nil {
		t.Fatalf("error creating new /users/login request: %v", err)
	}

	req.Header.Set("Cookie", resp.Header.Get("Set-Cookie"))

	client := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	resp, err = client.Do(req)
	if err != nil {
		t.Fatalf("error requesting the /users/login endpoint: %v", err)
	}
	got := resp.StatusCode
	expected := http.StatusForbidden
	if got != expected {
		t.Fatalf("invalid status code received. expected %d got %d", expected, got)
	}

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		t.Fatalf("error parsing /login response body")
	}

	if !strings.Contains(doc.Text(), InvalidCSRFTokenError) {
		t.Fatalf("Expected %s in response body but not found", InvalidCSRFTokenError)
	}

}

func TestDeleteValidEndpointID(t *testing.T) {
	testUser := models.User{
		ID:       1,
		Username: "test",
		Email:    "test@example.com",
	}
	ctx, resp := loggedSessionForTest(t, testUser)

	testEndpoint := models.Endpoint{
		ID:     1,
		UserID: 1,
	}

	body := struct {
		Key uint64 `json:"endpoint_delete_id"`
	}{
		Key: testEndpoint.ID,
	}

	out, err := json.Marshal(body)
	if err != nil {
		t.Fatalf("error Creating json for request body: %v", err)
	}

	ctx.mockGetByUserID(testUser)
	ctx.mockGetEndpointByID(testEndpoint)

	deleteURL := fmt.Sprintf("%s/endpoints/delete", ctx.server.URL)
	req, err := http.NewRequest("POST", deleteURL, bytes.NewBuffer(out))
	if err != nil {
		t.Fatalf("error creating new /users/login request: %v", err)
	}

	req.Header.Set("Cookie", resp.Header.Get("Set-Cookie"))

	client := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	resp, err = client.Do(req)
	if err != nil {
		t.Fatalf("error requesting the /users/login endpoint: %v", err)
	}
	got := resp.StatusCode
	expected := http.StatusForbidden
	if got != expected {
		t.Fatalf("invalid status code received. expected %d got %d", expected, got)
	}

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		t.Fatalf("error parsing /login response body")
	}

	if !strings.Contains(doc.Text(), InvalidCSRFTokenError) {
		t.Fatalf("Expected %s in response body but not found", InvalidCSRFTokenError)
	}

}

func TestDeleteEndpointWithInvalidSession(t *testing.T) {
	testUser := models.User{
		ID:       1,
		Username: "test",
		Email:    "test@example.com",
	}
	// ctx, resp := loggedSessionForTest(t, testUser)
	ctx := setupTest(t)

	testEndpoint := models.Endpoint{
		ID:     1,
		UserID: 1,
	}

	body := struct {
		Key uint64 `json:"endpoint_delete_id"`
	}{
		Key: testEndpoint.ID,
	}

	out, err := json.Marshal(body)
	if err != nil {
		t.Fatalf("error Creating json for request body: %v", err)
	}

	ctx.mockGetByUserID(testUser)
	ctx.mockGetEndpointByID(testEndpoint)

	deleteURL := fmt.Sprintf("%s/endpoint/delete", ctx.server.URL)
	req, err := http.NewRequest("POST", deleteURL, bytes.NewBuffer(out))
	if err != nil {
		t.Fatalf("error creating new /users/login request: %v", err)
	}

	client := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("error requesting the /users/login endpoint: %v", err)
	}
	got := resp.StatusCode
	expected := http.StatusForbidden
	if got != expected {
		t.Fatalf("invalid status code received. expected %d got %d", expected, got)
	}

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		t.Fatalf("error parsing /login response body")
	}

	if !strings.Contains(doc.Text(), InvalidCSRFTokenError) {
		t.Fatalf("Expected %s in response body but not found", InvalidCSRFTokenError)
	}

}
