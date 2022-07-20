package database

import (
	"database/sql"
	"time"

	"github.com/asimpleidea/appoint/api/services/pkg/types"
	"gorm.io/gorm"
)

/*
	TODOs:
	- field IsCategoryOnly bool
	- field GalleryID *uint
*/

type Service struct {
	gorm.Model
	ParentID    *uint
	Name        string `gorm:"size:100"`
	Description string `gorm:"size:300"`
	Price       sql.NullFloat64
	PublicPrice bool
}

func (s *Service) TableName() string {
	return servicesTable
}

func (s *Service) toAPI() *types.Service {
	return &types.Service{
		ID:        s.ID,
		ParentID:  s.ParentID,
		CreatedAt: s.CreatedAt,
		UpdatedAt: s.UpdatedAt,
		DeletedAt: func() *time.Time {
			if s.DeletedAt.Valid {
				return &s.DeletedAt.Time
			}

			return nil
		}(),
		Name:        s.Name,
		Description: s.Description,
		Price: func() *float64 {
			if s.Price.Valid {
				return &s.Price.Float64
			}

			return nil
		}(),
		PublicPrice: s.PublicPrice,
	}
}
