package types

import (
	"time"
)

type Service struct {
	ID          uint       `json:"id" yaml:"id"`
	ParentID    *uint      `json:"parent_id,omitempty" yaml:"parentId,omitempty"`
	CreatedAt   time.Time  `json:"created_at" yaml:"createdAt"`
	UpdatedAt   time.Time  `json:"updated_at" yaml:"updatedAt"`
	DeletedAt   *time.Time `json:"deleted_at,omitempty" yaml:"deletedAt,omitempty"`
	Name        string     `json:"name" yaml:"name"`
	Description string     `json:"description" yaml:"description"`
	Price       *float64   `json:"price,omitempty" yaml:"price,omitempty"`
	PublicPrice bool       `json:"public_price" yaml:"publicPrice"`
}
