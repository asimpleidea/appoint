package types

import (
	"time"
)

type Service struct {
	ID          uint       `json:"id" yaml:"id"`
	CreatedAt   time.Time  `json:"created_at" yaml:"createdAt"`
	UpdatedAt   time.Time  `json:"updated_at" yaml:"updatedAt"`
	DeletedAt   *time.Time `json:"deleted_at" yaml:"deletedAt"`
	Name        string     `json:"name" yaml:"name"`
	Description string     `json:"description" yaml:"description"`
	Price       *float32   `json:"price" yaml:"price"`
	PublicPrice bool       `json:"public_price" yaml:"publicPrice"`
}
