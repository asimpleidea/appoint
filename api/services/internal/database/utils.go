package database

import (
	"database/sql"
	"fmt"

	"github.com/asimpleidea/appoint/api/services/pkg/types"
	"gorm.io/gorm"
)

func checkServiceBeforePut(service *types.Service) (*Service, error) {
	serviceToReturn := &Service{Model: gorm.Model{
		ID:        service.ID,
		CreatedAt: service.CreatedAt,
		UpdatedAt: service.UpdatedAt,
		DeletedAt: func() gorm.DeletedAt {
			if service.DeletedAt != nil {
				return gorm.DeletedAt{
					Time:  *service.DeletedAt,
					Valid: true,
				}
			}

			return gorm.DeletedAt{
				Valid: false,
			}
		}(),
	}}

	// -- Check the name
	switch l := len(service.Name); {
	case l == 0:
		return nil, fmt.Errorf("no service name provided")
	case l > maxServiceNameLength:
		return nil, fmt.Errorf("service name too long")
	default:
		serviceToReturn.Name = service.Name
	}

	// -- Check the description
	if len(service.Description) > maxServiceDescriptionLength {
		return nil, fmt.Errorf("service description too long")
	}
	serviceToReturn.Description = service.Description

	// -- Check the price
	var price sql.NullFloat64
	if service.Price != nil {
		if *service.Price < 0 {
			return nil, fmt.Errorf("invalid price")
		}

		price = sql.NullFloat64{
			Valid:   true,
			Float64: *service.Price,
		}
	} else {
		price = sql.NullFloat64{
			Valid: false,
		}
	}
	serviceToReturn.Price = price
	serviceToReturn.PublicPrice = service.PublicPrice

	return serviceToReturn, nil
}
