package database

import "gorm.io/gorm"

func byTimetableID(id uint) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		return db.
			Where("id = ?", id)
	}
}

func byParentTimetableID(id uint) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		return db.
			Where("timetable_id = ?", id)
	}
}

func byDayOfWeek(dow DOW) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		return db.
			Where("dow = ?", dow)
	}
}
