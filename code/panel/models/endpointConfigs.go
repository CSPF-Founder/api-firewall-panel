package models

import (
	"slices"
	"strconv"

	"gorm.io/gorm"
)

const (
	CLIENT_IP_HEADER  = "CLIENT_IP_HEADER"
	DENY_TOKEN_HEADER = "DENY_TOKEN_HEADER"
)

var ConfigKeys = []string{CLIENT_IP_HEADER, DENY_TOKEN_HEADER}

// Define a constant array of strings
var IPHeaders = []string{"x-forwarded-for"}
var TokenHeaders = []string{"x-api-key", "Authorization"}

// EndpointConfigurations represents the endpoint_configuration model.
type EndpointConfig struct {
	ID          uint64 `gorm:"default:uuid_short()"`
	EndpointID  uint64 `json:"endpoint_id" sql:"not null"`
	ConfigKey   string `json:"config_key" sql:"not null"`
	ConfigValue string `json:"config_value" sql:"null"`
	CustomValue string `gorm:"-"`
}

func (e *EndpointConfig) AfterFind(tx *gorm.DB) error {
	if e.ConfigKey == CLIENT_IP_HEADER {
		e.CustomValue = strconv.FormatBool(!slices.Contains(IPHeaders, e.ConfigValue))
	} else if e.ConfigKey == DENY_TOKEN_HEADER {
		e.CustomValue = strconv.FormatBool(!slices.Contains(TokenHeaders, e.ConfigValue))
	}
	return nil
}

func SaveEndpointConfig(u *EndpointConfig) error {
	return db.Save(&u).Error
}

func GetConfigsByEndpointID(endpointID uint64) ([]EndpointConfig, error) {
	var ecList []EndpointConfig
	err := db.Where("endpoint_id=?", endpointID).Find(&ecList).Error
	return ecList, err
}

func GetConfigByEndpointIDAndKey(endpointID uint64, headerKey string) (EndpointConfig, error) {
	var ec EndpointConfig
	err := db.Where("endpoint_id=?", endpointID).Where("config_key=?", headerKey).First(&ec).Error
	return ec, err
}

// Delete DeleteEndpointConfig removes the config from table
// error is thrown if anything happens.
func DeleteEndpointConfig(endpoint_id uint64, headerKey string) error {
	err := db.Where("config_key=?", headerKey).Delete(&EndpointConfig{}, endpoint_id).Error
	return err
}
