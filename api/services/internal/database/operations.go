package database

import (
	"errors"
	"fmt"

	"github.com/asimpleidea/appoint/api/services/pkg/types"
	"github.com/rs/zerolog"
	"gorm.io/gorm"
)

const (
	maxServiceNameLength        int = 100
	maxServiceDescriptionLength int = 300

	servicesTable string = "services"
)

type Database struct {
	DB     *gorm.DB
	Logger zerolog.Logger
}

func (d *Database) GetServiceByID(id uint) (*types.Service, error) {
	if id == 0 {
		return nil, fmt.Errorf("invalid id")
	}

	var service Service
	res := d.DB.Model(&Service{}).Scopes(byServiceID(id)).First(&service)
	if res.Error != nil {
		if errors.Is(res.Error, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("not found")
		}
	}

	return service.toAPI(), nil
}

func (d *Database) CreateService(service *types.Service) (*types.Service, error) {
	if service == nil {
		return nil, fmt.Errorf("no service provided")
	}

	serviceToCreate, err := checkServiceBeforePut(service)
	if err != nil {
		return nil, err
	}

	if service.ParentID != nil {
		if _, err := d.GetServiceByID(*service.ParentID); err != nil {
			return nil, fmt.Errorf("error while trying to get parent with ID %d: %w", *service.ParentID, err)
		}

		serviceToCreate.ParentID = service.ParentID
	}

	res := d.DB.Create(serviceToCreate)
	if res.Error != nil {
		return nil, res.Error
	}

	return serviceToCreate.toAPI(), nil
}

func (d *Database) UpdateService(service *types.Service) error {
	if service == nil {
		return fmt.Errorf("no service provided")
	}

	serviceToUpdate, err := checkServiceBeforePut(service)
	if err != nil {
		return err
	}

	serviceToUpdate.ID = service.ID

	if service.ParentID != nil {
		if _, err := d.GetServiceByID(*service.ParentID); err != nil {
			return fmt.Errorf("error while trying to get parent with ID %d: %w", *service.ParentID, err)
		}

		serviceToUpdate.ParentID = service.ParentID
	}

	res := d.DB.Save(serviceToUpdate)
	return res.Error
}

func (d *Database) DeleteService(id uint) error {
	if id == 0 {
		return fmt.Errorf("invalid id")
	}

	if _, err := d.GetServiceByID(id); err != nil {
		// TODO: check if not found
		return err
	}

	{
		var parentsCount int64
		res := d.DB.Model(&Service{}).Where("parent_id = ?", id).Count(&parentsCount)
		if res.Error != nil {
			return fmt.Errorf("error while checking if service has subservices: %w", res.Error)
		}

		if parentsCount > 0 {
			// TODO: provide way to delete all sub-services?
			return fmt.Errorf("service contains sub-services")
		}
	}

	res := d.DB.Scopes(byServiceID(id)).Delete(&Service{})
	return res.Error
}
