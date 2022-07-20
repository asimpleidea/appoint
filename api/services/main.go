package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"net/url"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"time"

	"github.com/asimpleidea/appoint/api/services/internal/database"
	"github.com/asimpleidea/appoint/api/services/pkg/types"
	"github.com/gofiber/fiber/v2"
	"github.com/rs/zerolog"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

const (
	fiberAppName string = "Services API server"
)

var (
	log zerolog.Logger
)

// TODO: this must be put somewhere because used frequently
type DatabaseOptions struct {
	Host     string `json:"host" yaml:"host"`
	Port     int    `json:"port" yaml:"port"`
	User     string `json:"user" yaml:"user"`
	Password string `json:"password" yaml:"password"`
	Name     string `json:"name" yaml:"name"`
	SSLMode  bool   `json:"ssl_mode" yaml:"sslMode"`
	Timezone string `json:"timezone" yaml:"timezone"`
}

func generatePostgresDSN(opts DatabaseOptions) string {
	data := map[string]string{
		"host":     opts.Host,
		"port":     strconv.Itoa(opts.Port),
		"user":     opts.User,
		"password": opts.Password,
		"dbname":   opts.Name,
		"sslmode": func() string {
			if opts.SSLMode {
				return "enable"
			}

			return "disable"
		}(),
		"timeZone": opts.Timezone,
	}

	dsn := []string{}
	for k, v := range data {
		dsn = append(dsn, fmt.Sprintf("%s=%s", k, v))
	}

	return strings.Join(dsn, " ")
}

func main() {
	dbOpts := &DatabaseOptions{}
	verbosity := 1
	var ops *database.Database

	// -----------------------------------------
	// CLI Flags
	// -----------------------------------------

	flag.IntVar(&verbosity, "verbosity", 1, "the verbosity level.")

	// TODO: service names in CLI flags are temporary
	flag.StringVar(&dbOpts.Host, "database.host", "localhost",
		"the main database where to connect to.")
	flag.IntVar(&dbOpts.Port, "database.port", 5432,
		"the port to use to connect to the database.")
	flag.StringVar(&dbOpts.User, "database.user", "postgres",
		"the user to use to authenticate to the database.")
	flag.StringVar(&dbOpts.Password, "database.password", "",
		"the password to use to authenticate to the database.")
	flag.StringVar(&dbOpts.Name, "database.name", "appointments",
		"the name of the database to use.")
	flag.BoolVar(&dbOpts.SSLMode, "database.sslmode", true,
		"whether to use SSL mode.")
	flag.StringVar(&dbOpts.Timezone, "database.timezone", "Europe/Rome",
		"the timezone to use for dates.")
	flag.Parse()

	// -----------------------------------------
	// Set the logger
	// -----------------------------------------

	log = zerolog.New(os.Stderr).With().Logger()
	log.Info().Int("verbosity", verbosity).Msg("starting...")

	{
		logLevels := [4]zerolog.Level{zerolog.DebugLevel, zerolog.InfoLevel, zerolog.ErrorLevel}
		log = log.Level(logLevels[verbosity])
	}

	// -----------------------------------------
	// Connect to the database
	// -----------------------------------------

	dsn := generatePostgresDSN(*dbOpts)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal().Err(err).Msg("could not establish connection to the database, exiting...")
		return
	}
	ops = &database.Database{DB: db, Logger: log}
	log.Debug().Msg("connected to the database")

	// -----------------------------------------
	// Start the REST API server
	// -----------------------------------------

	app := fiber.New(fiber.Config{
		AppName:               fiberAppName,
		ReadTimeout:           time.Minute,
		DisableStartupMessage: verbosity > 0,
	})

	services := app.Group("/services")

	services.Get("/:id", func(c *fiber.Ctx) error {
		var id uint
		{
			serviceID, err := url.PathUnescape(c.Params("id"))
			if err != nil || serviceID == "" {
				return c.Status(fiber.StatusBadRequest).
					Send([]byte("invalid id provided"))
			}

			servID, err := strconv.Atoi(serviceID)
			if err != nil {
				return c.Status(fiber.StatusBadRequest).
					Send([]byte("invalid id provided"))
			}

			id = uint(servID)
		}

		service, err := ops.GetServiceByID(id)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).
				Send([]byte(err.Error()))
		}

		return c.JSON(service)
	})

	services.Post("/", func(c *fiber.Ctx) error {
		c.Accepts(fiber.MIMEApplicationJSON)

		if len(c.Body()) == 0 {
			return c.Status(fiber.StatusBadGateway).
				Send([]byte("no service provided"))
		}

		var newService *types.Service
		if err := json.Unmarshal(c.Body(), &newService); err != nil {
			return c.Status(fiber.StatusBadGateway).
				Send([]byte("invalid service provided"))
		}

		// This is to prevent having ID, CreatedAt etc. in the request as well.
		// TODO: find an alternative way?
		createdServ, err := ops.CreateService(&types.Service{
			Name:        newService.Name,
			ParentID:    newService.ParentID,
			Description: newService.Description,
			Price:       newService.Price,
			PublicPrice: newService.PublicPrice,
		})
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).
				Send([]byte(err.Error()))
		}

		return c.Status(fiber.StatusCreated).JSON(createdServ)
	})

	services.Put("/:id", func(c *fiber.Ctx) error {
		c.Accepts(fiber.MIMEApplicationJSON)

		var id uint
		{
			serviceID, err := url.PathUnescape(c.Params("id"))
			if err != nil || serviceID == "" {
				return c.Status(fiber.StatusBadRequest).
					Send([]byte("invalid id provided"))
			}

			servID, err := strconv.Atoi(serviceID)
			if err != nil {
				return c.Status(fiber.StatusBadRequest).
					Send([]byte("invalid id provided"))
			}

			id = uint(servID)
		}

		if len(c.Body()) == 0 {
			return c.Status(fiber.StatusBadGateway).
				Send([]byte("no service provided"))
		}

		var serviceToUpdate *types.Service
		if err := json.Unmarshal(c.Body(), &serviceToUpdate); err != nil {
			return c.Status(fiber.StatusBadGateway).
				Send([]byte("invalid service provided"))
		}

		existingService, err := ops.GetServiceByID(id)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).
				Send([]byte(err.Error()))
		}

		if err := ops.UpdateService(&types.Service{
			ID:          existingService.ID,
			ParentID:    serviceToUpdate.ParentID,
			Name:        serviceToUpdate.Name,
			Description: serviceToUpdate.Description,
			Price:       serviceToUpdate.Price,
			PublicPrice: serviceToUpdate.PublicPrice,
		}); err != nil {
			return c.Status(fiber.StatusInternalServerError).
				Send([]byte(err.Error()))
		}

		return c.SendStatus(fiber.StatusOK)
	})

	services.Delete("/:id", func(c *fiber.Ctx) error {
		var id uint
		{
			serviceID, err := url.PathUnescape(c.Params("id"))
			if err != nil || serviceID == "" {
				return c.Status(fiber.StatusBadRequest).
					Send([]byte("invalid id provided"))
			}

			servID, err := strconv.Atoi(serviceID)
			if err != nil {
				return c.Status(fiber.StatusBadRequest).
					Send([]byte("invalid id provided"))
			}

			id = uint(servID)
		}

		if err := ops.DeleteService(id); err != nil {
			return c.Status(fiber.StatusInternalServerError).
				Send([]byte(err.Error()))
		}

		return c.SendStatus(fiber.StatusGone)
	})

	go func() {
		if err := app.Listen(":8080"); err != nil {
			log.Err(err).Msg("error while listening")
		}
	}()

	// -----------------------------------------
	// Graceful shutdown
	// -----------------------------------------

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)
	<-stop

	log.Info().Msg("shutting down...")
	if err := app.Shutdown(); err != nil {
		log.Err(err).Msg("error while waiting for server to shutdown")
	}
	log.Info().Msg("goodbye!")
}
