package database

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/asimpleidea/appoint/api/timetables/pkg/types"
	"github.com/rs/zerolog"
	"gorm.io/gorm"
)

/*
	TODOs:
	- function to check if timetable exists
*/

const (
	maxServiceNameLength        int = 100
	maxServiceDescriptionLength int = 300

	timetablesTable    string = "timetables"
	timetableDaysTable string = "timetable_days"
)

type Database struct {
	DB     *gorm.DB
	Logger zerolog.Logger
}

func (d *Database) GetTimetableByID(id uint, fullTimetable bool) (*types.Timetable, error) {
	if id == 0 {
		return nil, fmt.Errorf("invalid id")
	}

	var (
		timetable Timetable
		err       error
	)
	if err := d.DB.Model(&Timetable{}).
		Scopes(byTimetableID(id)).First(&timetable).Error; err != nil {
		return nil, err
	}

	timetableToReturn := timetable.ToAPI()

	if !fullTimetable {
		return timetableToReturn, nil
	}

	if timetableToReturn.Monday, err = d.GetWeekDay(id, Monday); err != nil {
		return nil, fmt.Errorf("cannot get monday data")
	}
	if timetableToReturn.Tuesday, err = d.GetWeekDay(id, Tuesday); err != nil {
		return nil, fmt.Errorf("cannot get tuesday data")
	}
	if timetableToReturn.Wednesday, err = d.GetWeekDay(id, Wednesday); err != nil {
		return nil, fmt.Errorf("cannot get wednesday data")
	}
	if timetableToReturn.Thursday, err = d.GetWeekDay(id, Thursday); err != nil {
		return nil, fmt.Errorf("cannot get thursday data")
	}
	if timetableToReturn.Friday, err = d.GetWeekDay(id, Friday); err != nil {
		return nil, fmt.Errorf("cannot get friday data")
	}
	if timetableToReturn.Saturday, err = d.GetWeekDay(id, Saturday); err != nil {
		return nil, fmt.Errorf("cannot get saturday data")
	}
	if timetableToReturn.Sunday, err = d.GetWeekDay(id, Sunday); err != nil {
		return nil, fmt.Errorf("cannot get sunday data")
	}

	return timetableToReturn, nil
}

func (d *Database) CreateTimetable(tt *types.Timetable) (*types.Timetable, error) {
	if tt == nil {
		return nil, fmt.Errorf("nil timetable provided")
	}

	now, _ := time.Parse("2006-01-02", time.Now().Format("2006-01-02"))

	// We make the comparisons in utc because we only care about the day here,
	// not the time.
	if tt.ValidFrom.UTC().Before(now.UTC()) {
		return nil, fmt.Errorf("cannot start a timetable before the current day")
	}

	if tt.ValidUntil != nil {
		if tt.ValidUntil.UTC().Before(tt.ValidFrom.UTC()) {
			return nil, fmt.Errorf("invalid end validity provided")
		}
	}

	timetableToCreate := &Timetable{
		Name:      tt.Name,
		ValidFrom: tt.ValidFrom,
		ValidUntil: func() sql.NullTime {
			if tt.ValidUntil != nil {
				return sql.NullTime{
					Time:  *tt.ValidUntil,
					Valid: true,
				}
			}

			return sql.NullTime{Valid: false}
		}(),
	}

	if err := d.DB.Create(timetableToCreate).Error; err != nil {
		return nil, err
	}

	return timetableToCreate.ToAPI(), nil
}

func (d *Database) GetWeekDay(timetableID uint, dow DOW) ([]types.TimetableDay, error) {
	switch dow {
	case Monday, Tuesday, Wednesday, Thursday, Friday, Saturday, Sunday:
		// ok
	default:
		return nil, fmt.Errorf("invalid day of week provided")
	}

	if timetableID == 0 {
		return nil, fmt.Errorf("invalid timetable provided")
	}

	{
		count := int64(0)
		if err := d.DB.Model(&Timetable{}).
			Scopes(byTimetableID(timetableID)).Count(&count).Error; err != nil {
			return nil, fmt.Errorf("cannot check if timetable exists: %w", err)
		}

		if count == 0 {
			return nil, gorm.ErrRecordNotFound
		}
	}

	dows := []TimetableDay{}
	if err := d.DB.Order("opening asc").Model(&TimetableDay{}).
		Scopes(byParentTimetableID(timetableID), byDayOfWeek(dow)).
		Find(&dows).Error; err != nil {
		return nil, err
	}

	converted := make([]types.TimetableDay, len(dows))
	for i := 0; i < len(dows); i++ {
		converted[i] = *dows[i].ToAPI()
	}

	return converted, nil
}

func (d *Database) CreateWeekDay(timetableID uint, dow DOW, openingClosing [][2]string) ([]types.TimetableDay, error) {
	if len(openingClosing) == 0 {
		return nil, fmt.Errorf("no opening closing times provided")
	}

	{
		count := int64(0)
		if err := d.DB.Model(&Timetable{}).
			Scopes(byTimetableID(timetableID)).Count(&count).Error; err != nil {
			return nil, fmt.Errorf("cannot check if timetable exists: %w", err)
		}

		if count == 0 {
			return nil, gorm.ErrRecordNotFound
		}
	}

	times := [][2]time.Time{}
	toCreate := []TimetableDay{}
	const timeFormat = "15:04"

	// First, check them...
	for _, t := range openingClosing {
		if len(t) == 0 {
			continue
		}

		if t[0] == "" {
			return nil, fmt.Errorf("invalid opening time provided")
		}

		opening, err := time.Parse(timeFormat, t[0])
		if err != nil {
			return nil, fmt.Errorf("invalid opening time provided: %w", err)
		}

		if t[1] == "" {
			return nil, fmt.Errorf("invalid closing time provided")
		}

		closing, err := time.Parse(timeFormat, t[1])
		if err != nil {
			return nil, fmt.Errorf("invalid closing time provided: %w", err)
		}

		if closing.Before(opening) {
			return nil, fmt.Errorf("invalid closing time provided")
		}

		for _, previous := range times {
			if !opening.After(previous[0]) && !opening.After(previous[1]) {
				return nil, fmt.Errorf("invalid opening time provided %s: %w", opening, err)
			}
		}

		times = append(times, [2]time.Time{opening, closing})
		toCreate = append(toCreate, TimetableDay{
			TimetableID: timetableID,
			Dow:         dow,
			Opening:     opening.Format(timeFormat),
			Closing:     closing.Format(timeFormat),
		})
	}

	d.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Scopes(byParentTimetableID(timetableID), byDayOfWeek(dow)).Delete(&TimetableDay{}).Error; err != nil {
			return fmt.Errorf("cannot delete existing timetable days: %w", err)
		}

		if err := tx.Create(toCreate).Error; err != nil {
			return fmt.Errorf("cannot create timetable days: %w", err)
		}

		return nil
	})

	weekDays := make([]types.TimetableDay, len(toCreate))
	for i := 0; i < len(toCreate); i++ {
		weekDays[i] = *toCreate[i].ToAPI()
	}

	return weekDays, nil
}

func (d *Database) DeleteWeekDay(timetableID uint, dow DOW) error {
	{
		count := int64(0)
		if err := d.DB.Model(&Timetable{}).
			Scopes(byTimetableID(timetableID)).Count(&count).Error; err != nil {
			return fmt.Errorf("cannot check if timetable exists: %w", err)
		}

		if count == 0 {
			return gorm.ErrRecordNotFound
		}
	}

	if err := d.DB.Scopes(byParentTimetableID(timetableID), byDayOfWeek(dow)).Delete(&TimetableDay{}).Error; err != nil {
		return fmt.Errorf("cannot delete timetable days: %w", err)
	}

	return nil
}

func (d *Database) DeleteTimetable(id uint) error {
	if id == 0 {
		return fmt.Errorf("invalid id provided")
	}

	if err := d.DB.Delete(&Timetable{}, id).Error; err != nil {
		return fmt.Errorf("could not delete timetable: %w", err)
	}

	if err := d.DB.Scopes(byParentTimetableID(id)).Delete(&Timetable{}).Error; err != nil {
		return fmt.Errorf("cannot delete days for timetable: %w", err)
	}

	return nil
}
