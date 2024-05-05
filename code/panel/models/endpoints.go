package models

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/CSPF-Founder/api-firewall-panel/api-protector/onpremise/panel/pkg/dockerutil"
	"gorm.io/gorm"
)

const DateTimeFormat = "2006-01-02 3:4 PM"

const (
	MIN_HEALTH_PORT      = 60001
	MAX_HEALTH_PORT      = 119001
	MIN_AUTO_PORT        = 3000
	MAX_AUTO_PORT        = 3100
	MIN_CUSTOM_PORT      = 1000
	MAX_CUSTOM_PORT      = 60000
	DEFAULT_REQUEST_MODE = "monitor"
)

// Endpoint represents the endpoint model.
type Endpoint struct {
	ID              uint64    `gorm:"primaryKey;default:uuid_short()"`
	StatusText      string    `gorm:"-"`
	Label           string    `json:"label" sql:"null"`
	ApiUrl          string    `json:"api_url" sql:"null"`
	ListeningPort   int       `json:"listening_port" sql:"null"`
	HealthPort      int       `json:"health_port" sql:"null"`
	RequestMode     string    `json:"request_mode" sql:"null"`
	CreatedAt       time.Time `json:"created_at" sql:"null"`
	CreatedAtString string    `gorm:"-"`
	UserID          int64     `json:"user_id" sql:"not null"`

	// for db relationship
	Config EndpointConfig
}

func (e *Endpoint) AfterFind(tx *gorm.DB) (err error) {
	e.CreatedAtString = e.CreatedAt.Format(DateTimeFormat)
	status, dockerErr := dockerutil.GetClient().GetContainerStatusByName(context.TODO(), e.Label)
	if dockerErr == nil && strings.ToLower(status) == "running" {
		e.StatusText = "Running"
	} else {
		e.StatusText = "Not Running"
	}
	return nil
}

func GetEndpointByIDAndUser(id uint64, userID int64) (Endpoint, error) {
	var endpoint Endpoint
	err := db.Where("id=? AND user_id=?", id, userID).First(&endpoint).Error
	return endpoint, err
}

// GetEndpointsWithConfig returns the endpoints with configurations
func GetEndpointListWithConfig(u *User, configType string) ([]Endpoint, error) {
	us := []Endpoint{}
	err := db.Preload("Config", "config_key = ?", configType).Where("user_id = ?", u.ID).Find(&us).Error
	return us, err
}

// GetEndpointByLabel returns the enpoint that the given label corresponds to. If no endpoint is found, an
// error is thrown.
func GetEndpointByLabel(label string) (Endpoint, error) {
	var endpoint Endpoint
	err := db.Where("label=?", label).First(&endpoint).Error
	return endpoint, err
}

// Delete Job removes the job from table
// error is thrown if anything happens.
func DeleteEndpoint(id uint64) error {
	err := db.Delete(&Endpoint{}, id).Error
	return err
}

// GetJobs returns the jobs
func GetEndpoints(u *User) ([]Endpoint, error) {
	us := []Endpoint{}
	err := db.Where("user_id = ?", u.ID).Find(&us).Error
	return us, err
}

// SaveJob saves the job to the database
func SaveEndpoint(u *Endpoint) error {
	err := db.Save(&u).Error
	db.Last(&u)
	return err
}

func CheckPortExist(availablePort int) bool {
	var count int64

	err := db.Model(&Endpoint{}).Where("listening_port=?", availablePort).Count(&count).Error
	if err != nil {
		return false
	}

	return count > 0
}

func GetAvailablePort() int {
	var availablePort int

	// Subquery to find available port
	subquery := db.Table("(SELECT ? AS listening_port "+
		"UNION "+
		"SELECT listening_port + 1 "+
		"FROM endpoints WHERE listening_port < ?) AS temp",
		MIN_AUTO_PORT, MAX_AUTO_PORT).
		Select("MIN(listening_port)").Where("listening_port NOT IN (?)",
		db.Table("endpoints").Select("listening_port").
			Where("listening_port BETWEEN ? AND ?", MIN_AUTO_PORT, MAX_AUTO_PORT))

	// Execute the subquery
	err := subquery.Row().Scan(&availablePort)
	if err != nil {
		return 0
	}
	return availablePort
}

func GetAvailableHealthPort() int {
	var availablePort int

	// Subquery to find available port
	subquery := db.Table("(SELECT ? AS health_port "+
		"UNION SELECT health_port + 1 "+
		"FROM endpoints WHERE health_port < ?) AS temp",
		MIN_HEALTH_PORT, MAX_HEALTH_PORT).
		Select("MIN(health_port)").Where("health_port NOT IN (?)",
		db.Table("endpoints").Select("health_port").
			Where("health_port BETWEEN ? AND ?", MIN_HEALTH_PORT, MAX_HEALTH_PORT))

	// Execute the subquery
	err := subquery.Row().Scan(&availablePort)
	if err != nil {
		return 0
	}
	return availablePort
}

func GetOpenApiFilePath(endpoint Endpoint, workDir string) (string, string) {

	targetDir := filepath.Join(workDir, fmt.Sprintf("input/%s", endpoint.Label))
	destinationFile := filepath.Join(targetDir, "openapi.yaml")

	return targetDir, destinationFile
}

func GetDockerStackDir(endpoint Endpoint, workDir string) string {

	targetDir := filepath.Join(workDir, fmt.Sprintf("docker/stack_%s", endpoint.Label))

	return targetDir
}

func GetDockerLogsDir(endpoint Endpoint, workDir string) string {

	targetDir := GetDockerStackDir(endpoint, workDir)
	logsDir := filepath.Join(targetDir, "/logs")

	return logsDir
}

// GetEndpointsByID returns the endpoint that the given id corresponds to. If no endpoint is found, an
// error is thrown.
// func GetEndpointByID(id uint64) (Endpoint, error) {
// 	var endpoint Endpoint
// 	err := db.Where("id=?", id).First(&endpoint).Error
// 	return endpoint, err
// }
