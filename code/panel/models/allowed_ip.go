package models

// AllowedIP represents the allowed_ip model.
type AllowedIP struct {
	ID         int64  `gorm:"default:uuid_short()"`
	EndpointID uint64 `json:"endpoint_id" sql:"not null"`
	IPData     string `json:"ip_data" sql:"not null"`
	IsRange    int    `json:"is_range" sql:"not null"`
}

// GetAllowedIPByEndpointID returns the allowed ip by it's endpoint id. If no endpoint is found, an
// error is thrown.
func GetAllowedIPByEndpointID(endpoint_id uint64) ([]AllowedIP, error) {
	var denied_tokens []AllowedIP
	err := db.Where("endpoint_id=?", endpoint_id).Find(&denied_tokens).Error
	return denied_tokens, err
}

// SaveAllowedIP saves the allowed_ip to the database
func SaveAllowedIP(u *AllowedIP) error {
	err := db.Save(&u).Error
	return err
}

// CheckAllowedIPExist checks if allowed_ip already exists
func CheckAllowedIPExist(ip string, endpoint_id uint64) bool {
	var count int64
	db.Table("allowed_ips").Where("ip_data=?", ip).Where("endpoint_id=?", endpoint_id).Count(&count)
	return count > 0
}

// Delete DeleteAllowedIP removes the allowed_ip from table
// error is thrown if anything happens.
func DeleteAllowedIP(endpoint_id uint64, id int64) error {
	err := db.Where("endpoint_id=?", endpoint_id).Delete(&AllowedIP{}, id).Error
	return err
}
