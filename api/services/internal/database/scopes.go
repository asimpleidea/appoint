package database

import "gorm.io/gorm"

func byServiceID(id uint) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		return db.
			Where("id = ? AND deleted_at IS NULL", id)
	}
}
