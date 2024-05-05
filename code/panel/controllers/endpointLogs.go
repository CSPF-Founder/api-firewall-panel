package controllers

import (
	"fmt"
	"io/fs"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	ctx "github.com/CSPF-Founder/api-firewall-panel/api-protector/onpremise/panel/context"
	"github.com/CSPF-Founder/api-firewall-panel/api-protector/onpremise/panel/enums/flashtypes"
	"github.com/CSPF-Founder/api-firewall-panel/api-protector/onpremise/panel/models"
	"github.com/CSPF-Founder/api-firewall-panel/api-protector/onpremise/panel/views"
	"github.com/go-chi/chi/v5"
)

type logsController struct {
	*App
}

type LogObject struct {
	Log        []map[string]string
	EndpointId uint64
}

type LogFile struct {
	Name string
	Path string
}

func newLogsController(app *App) *logsController {
	return &logsController{
		App: app,
	}
}

func (c *logsController) registerRoutes() http.Handler {
	router := chi.NewRouter()

	// Authenticated Routes
	router.Group(func(r chi.Router) {
		// r.Use(mid.RequireLogin)
		r.Get("/", c.List)     // List all Logs
		r.Get("/view", c.View) // List all Logs
		// r.Get("/view", c.DisplayAdd)                 // Display Log
		r.Post("/download", c.Download) // Download log
	})

	return router
}

func (c *logsController) List(w http.ResponseWriter, r *http.Request) {
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
	}
	targetDir := filepath.Join(c.config.WorkDir, fmt.Sprintf("docker/stack_%s/logs", endpoint.Label))
	logFiles, l_err := getLogFilesFromDir(targetDir)

	if l_err != nil {
		c.FlashAndGoBack(w, r, flashtypes.FlashWarning, "No Logs found for this Endpoint")
		return
	}

	if len(logFiles) < 1 {
		c.FlashAndGoBack(w, r, flashtypes.FlashWarning, "No Logs found for this Endpoint")
		return
	}

	templateData := views.NewTemplateData(c.config, c.session, r)
	templateData.Title = "API Protector Endpoints"
	templateData.Data = struct {
		Log        []LogFile
		EndpointId uint64
	}{logFiles, endpoint.ID}
	if err := views.RenderTemplate(w, "logs/list", templateData); err != nil {
		c.logger.Error("Error rendering template: ", err)
	}
}

func (c *logsController) View(w http.ResponseWriter, r *http.Request) {
	inputID := chi.URLParam(r, "endpointID")
	untrustedInput := r.URL.Query().Get("file")
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

	logsDir := models.GetDockerLogsDir(endpoint, c.config.WorkDir)
	logFiles, err := getLogFilesFromDir(logsDir)
	if err != nil || len(logFiles) == 0 {
		c.FlashAndGoBack(w, r, flashtypes.FlashWarning, "No Logs found for this Endpoint")
		return
	}

	var targetFilePath string
	var targetFileName string

	for _, logFile := range logFiles {
		if logFile.Name == untrustedInput {
			targetFilePath = logFile.Path
			targetFileName = logFile.Name
			break
		}
	}

	if targetFilePath == "" || targetFileName == "" {
		c.FlashAndGoBack(w, r, flashtypes.FlashWarning, "Invalid log file name")
		return
	}

	if _, err := os.Stat(targetFilePath); os.IsNotExist(err) {
		c.FlashAndGoBack(w, r, flashtypes.FlashWarning, "Log file does not exist")
		return
	}

	maxFileSize := int64(5 * 1024 * 1024) // 5MB in bytes
	file, err := os.Open(targetFilePath)
	if err != nil {
		c.FlashAndGoBack(w, r, flashtypes.FlashWarning, "Failed to open the log file")
		return
	}
	defer file.Close()

	fileInfo, err := file.Stat()
	if err != nil {
		c.FlashAndGoBack(w, r, flashtypes.FlashWarning, "Failed to get file information")
		return
	}

	fileSize := fileInfo.Size()
	var content []byte

	if fileSize > maxFileSize {
		c.FlashAndGoBack(w, r, flashtypes.FlashWarning, "Log file is too big, only showing last 5MB")
		content, err = readLastNBytesFromFile(file, maxFileSize)
		if err != nil {
			c.FlashAndGoBack(w, r, flashtypes.FlashWarning, "Failed to read last 5MB from file")
			return
		}
	} else {
		content, err = os.ReadFile(targetFilePath)
		if err != nil {
			c.FlashAndGoBack(w, r, flashtypes.FlashWarning, "Failed to read Log content")
			return
		}
	}

	templateData := views.NewTemplateData(c.config, c.session, r)
	templateData.Title = "API Protector Endpoints"
	templateData.Data = string(content)
	if err := views.RenderTemplate(w, "logs/view", templateData); err != nil {
		c.logger.Error("Error rendering template: ", err)
	}
}

func (c *logsController) Download(w http.ResponseWriter, r *http.Request) {
	inputID := chi.URLParam(r, "endpointID")
	untrustedInput := r.PostFormValue("file")
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

	logsDir := models.GetDockerLogsDir(endpoint, c.config.WorkDir)
	logFiles, err := getLogFilesFromDir(logsDir)
	if err != nil || len(logFiles) == 0 {
		c.FlashAndGoBack(w, r, flashtypes.FlashWarning, "No Logs found for this Endpoint")
		return
	}

	var targetFilePath string
	var targetFileName string

	for _, logFile := range logFiles {
		if logFile.Name == untrustedInput {
			targetFilePath = logFile.Path
			targetFileName = logFile.Name
			break
		}
	}

	if targetFilePath == "" || targetFileName == "" {
		c.FlashAndGoBack(w, r, flashtypes.FlashWarning, "Invalid log file name")
		return
	}

	if _, err := os.Stat(targetFilePath); os.IsNotExist(err) {
		c.FlashAndGoBack(w, r, flashtypes.FlashWarning, "Log file does not exist")
		return
	}

	// Open the file
	file, err := os.Open(targetFilePath)
	if err != nil {
		c.FlashAndGoBack(w, r, flashtypes.FlashWarning, "Failed to open file")
		return
	}
	defer file.Close()

	// Get file information
	fileInfo, err := file.Stat()
	if err != nil {
		c.FlashAndGoBack(w, r, flashtypes.FlashWarning, "Failed to get file information")
		return
	}

	// Set headers to initiate download
	w.Header().Set("Content-Type", "text/plain")
	w.Header().Set("Content-Disposition", "attachment; filename="+targetFileName)
	w.Header().Set("Content-Length", strconv.FormatInt(fileInfo.Size(), 10))
	w.Header().Set("Pragma", "no-cache")

	// Serve the file content
	http.ServeContent(w, r, targetFileName, fileInfo.ModTime(), file)

}

func getLogFilesFromDir(targetDir string) ([]LogFile, error) {
	var logFiles []LogFile

	// Walk through the directory
	err := filepath.WalkDir(targetDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		// Skip directories
		if d.IsDir() {
			return nil
		}

		// Check if the file starts with "firewall"
		if strings.HasPrefix(d.Name(), "firewall") {
			logFiles = append(logFiles, LogFile{Name: d.Name(), Path: path})
		}

		return nil
	})

	return logFiles, err
}

func readLastNBytesFromFile(file *os.File, n int64) ([]byte, error) {
	stat, err := file.Stat()
	if err != nil {
		return nil, err
	}
	fileSize := stat.Size()
	if fileSize < n {
		n = fileSize
	}

	buf := make([]byte, n)
	_, err = file.ReadAt(buf, fileSize-n)
	if err != nil {
		return nil, err
	}

	return buf, nil
}
