package database

import (
	"database/sql"
	"time"

	"github.com/asimpleidea/appoint/api/timetables/pkg/types"
	"gorm.io/gorm"
)

/*
	TODOs:
	- put gorm tags
*/

type Timetable struct {
	gorm.Model
	Name       string
	ValidFrom  time.Time
	ValidUntil sql.NullTime
}

func (t *Timetable) ToAPI() *types.Timetable {
	return &types.Timetable{
		ID:        t.ID,
		CreatedAt: t.CreatedAt,
		UpdatedAt: t.UpdatedAt,
		DeletedAt: func() *time.Time {
			var deleted time.Time
			if !t.DeletedAt.Valid {
				return nil
			}

			deleted = t.DeletedAt.Time
			return &deleted
		}(),
		Name:      t.Name,
		ValidFrom: t.ValidFrom,
		ValidUntil: func() *time.Time {
			var until time.Time
			if !t.ValidUntil.Valid {
				return nil
			}

			until = t.ValidUntil.Time
			return &until
		}(),
	}
}

func (t *Timetable) TableName() string {
	return "timetables"
}

type DOW string

const (
	Monday    DOW = "monday"
	Tuesday   DOW = "tuesday"
	Wednesday DOW = "wednesday"
	Thursday  DOW = "thursday"
	Friday    DOW = "friday"
	Saturday  DOW = "saturday"
	Sunday    DOW = "sunday"
)

type TimetableDay struct {
	gorm.Model
	TimetableID uint
	Dow         DOW
	Opening     string
	Closing     string
}

func (t *TimetableDay) ToAPI() *types.TimetableDay {
	return &types.TimetableDay{
		ID:        t.ID,
		CreatedAt: t.CreatedAt,
		UpdatedAt: t.UpdatedAt,
		DeletedAt: func() *time.Time {
			var deleted time.Time
			if !t.DeletedAt.Valid {
				return nil
			}

			deleted = t.DeletedAt.Time
			return &deleted
		}(),
		TimeTableID: t.TimetableID,
		DayOfWeek:   types.DOW(t.Dow),
		Opening:     t.Opening,
		Closing:     t.Closing,
	}
}

func (t *TimetableDay) TableName() string {
	return "timetable_days"
}
