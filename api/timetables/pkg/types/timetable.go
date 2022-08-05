package types

import "time"

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

type Timetable struct {
	ID         uint           `json:"id" yaml:"id"`
	CreatedAt  time.Time      `json:"created_at" yaml:"createdAt"`
	UpdatedAt  time.Time      `json:"updated_at" yaml:"updatedAt"`
	DeletedAt  *time.Time     `json:"deleted_at,omitempty" yaml:"deletedAt,omitempty"`
	Name       string         `json:"name" yaml:"name"`
	ValidFrom  time.Time      `json:"valid_from" yaml:"validFrom"`
	ValidUntil *time.Time     `json:"valid_until" yaml:"validUntil"`
	Monday     []TimetableDay `json:"monday,omitempty" yaml:"monday,omitempty"`
	Tuesday    []TimetableDay `json:"tuesday,omitempty" yaml:"tuesday,omitempty"`
	Wednesday  []TimetableDay `json:"wednesday,omitempty" yaml:"wednesday,omitempty"`
	Thursday   []TimetableDay `json:"thursday,omitempty" yaml:"thursday,omitempty"`
	Friday     []TimetableDay `json:"friday,omitempty" yaml:"friday,omitempty"`
	Saturday   []TimetableDay `json:"saturday,omitempty" yaml:"saturday,omitempty"`
	Sunday     []TimetableDay `json:"sunday,omitempty" yaml:"sunday,omitempty"`
}

type TimetableDay struct {
	ID          uint       `json:"id" yaml:"id"`
	CreatedAt   time.Time  `json:"created_at" yaml:"createdAt"`
	UpdatedAt   time.Time  `json:"updated_at" yaml:"updatedAt"`
	DeletedAt   *time.Time `json:"deleted_at,omitempty" yaml:"deletedAt,omitempty"`
	TimeTableID uint       `json:"timetable_id" yaml:"timetableId"`
	DayOfWeek   DOW        `json:"day_of_week" yaml:"dayOfWeek"`
	Opening     string     `json:"opening" yaml:"opening"`
	Closing     string     `json:"closing" yaml:"closing"`
}
