package models

type DeniedToken struct {
	ID         uint64 `gorm:"default:uuid_short()"`
	EndpointID uint64 `json:"endpoint_id" sql:"not null"`
	Token      string `json:"token" sql:"not null"`
}

func GetDeniedTokensByEndpointID(endpoint_id uint64) ([]DeniedToken, error) {
	var denied_tokens []DeniedToken
	err := db.Where("endpoint_id=?", endpoint_id).Find(&denied_tokens).Error
	return denied_tokens, err
}

func SaveDeniedToken(u *DeniedToken) error {
	err := db.Save(&u).Error
	return err
}

func CheckDeniedTokenExist(token string, endpoint_id uint64) bool {
	var count int64
	db.Table("denied_tokens").Where("token=?", token).Where("endpoint_id=?", endpoint_id).Count(&count)
	return count > 0
}

func DeleteDenyToken(endpoint_id uint64, id int64) error {
	err := db.Where("endpoint_id=?", endpoint_id).Delete(&DeniedToken{}, id).Error
	return err
}
