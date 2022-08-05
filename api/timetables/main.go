package main

import (
	"encoding/json"
	"errors"
	"flag"
	"net/url"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"time"

	"github.com/asimpleidea/appoint/api/timetables/internal/database"
	"github.com/asimpleidea/appoint/api/timetables/pkg/types"
	"github.com/gofiber/fiber/v2"
	"github.com/rs/zerolog"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	coredb "github.com/asimpleidea/appoint/api/core/pkg/database"
)

const (
	fiberAppName string = "Timetables API server"
)

var (
	log zerolog.Logger
)

/*
	TODOs:
	- many of these IDs could be got in one only function
	- parse dows in one function or operations
*/

func main() {
	dbOpts := &coredb.Options{}
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

	dsn := coredb.GeneratePostgresDSN(*dbOpts)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal().Err(err).Msg("could not establish connection to the database, exiting...")
		return
	}
	ops = &database.Database{DB: db, Logger: log}
	log.Debug().Msg("connected to the database")

	// // -----------------------------------------
	// // Start the REST API server
	// // -----------------------------------------

	app := fiber.New(fiber.Config{
		AppName:               fiberAppName,
		ReadTimeout:           time.Minute,
		DisableStartupMessage: verbosity > 0,
	})

	timetables := app.Group("/timetables")

	timetables.Get("/:id", func(c *fiber.Ctx) error {
		var id uint
		{
			// TODO: this part may go to core?
			timetableID, err := url.PathUnescape(c.Params("id"))
			if err != nil || timetableID == "" {
				return c.Status(fiber.StatusBadRequest).
					Send([]byte("invalid id provided"))
			}

			tid, err := strconv.Atoi(timetableID)
			if err != nil {
				return c.Status(fiber.StatusBadRequest).
					Send([]byte("invalid id provided"))
			}

			id = uint(tid)
		}

		fullTimetable := false
		{
			getFull := strings.ToLower(c.Query("full-timetable", "false"))
			if getFull == "true" {
				fullTimetable = true
			}
		}

		tt, err := ops.GetTimetableByID(id, fullTimetable)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).
				Send([]byte(err.Error()))
		}

		return c.JSON(tt)
	})

	timetables.Post("/", func(c *fiber.Ctx) error {
		c.Accepts(fiber.MIMEApplicationJSON)

		if len(c.Body()) == 0 {
			return c.Status(fiber.StatusBadGateway).
				Send([]byte("no timetable provided"))
		}

		var newTimeTable *types.Timetable
		if err := json.Unmarshal(c.Body(), &newTimeTable); err != nil {
			return c.Status(fiber.StatusBadGateway).
				Send([]byte("invalid timetable provided"))
		}

		createdTt, err := ops.CreateTimetable(newTimeTable)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).
				Send([]byte(err.Error()))
		}

		return c.Status(fiber.StatusCreated).JSON(createdTt)
	})

	timetables.Delete("/:id", func(c *fiber.Ctx) error {
		var id uint
		{
			timetableID, err := url.PathUnescape(c.Params("id"))
			if err != nil || timetableID == "" {
				return c.Status(fiber.StatusBadRequest).
					Send([]byte("invalid id provided"))
			}

			tid, err := strconv.Atoi(timetableID)
			if err != nil {
				return c.Status(fiber.StatusBadRequest).
					Send([]byte("invalid id provided"))
			}

			id = uint(tid)
		}

		if err := ops.DeleteTimetable(id); err != nil {
			return c.Status(fiber.StatusInternalServerError).
				JSON(err)
		}

		return c.SendStatus(fiber.StatusOK)
	})

	timetables.Get(":id/:dow", func(c *fiber.Ctx) error {
		var id uint
		{
			timetableID, err := url.PathUnescape(c.Params("id"))
			if err != nil || timetableID == "" {
				return c.Status(fiber.StatusBadRequest).
					Send([]byte("invalid id provided"))
			}

			tid, err := strconv.Atoi(timetableID)
			if err != nil {
				return c.Status(fiber.StatusBadRequest).
					Send([]byte("invalid id provided"))
			}

			id = uint(tid)

			// TODO: function that just exists?
			if _, err := ops.GetTimetableByID(id, false); err != nil {
				if errors.Is(gorm.ErrRecordNotFound, err) {
					return c.SendStatus(fiber.StatusNotFound)
				}

				return c.Status(fiber.StatusInternalServerError).
					JSON(err)
			}
		}

		dow := strings.ToLower(c.Params("dow"))
		switch dow {
		// TODO: maybe use fiber regex for this?
		case "monday", "tuesday", "wednesday", "thursday", "friday",
			"saturday", "sunday":
			// OK
		default:
			return c.SendStatus(fiber.StatusNotFound)
		}

		res, err := ops.GetWeekDay(id, database.DOW(dow))
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(err)
		}

		return c.JSON(res)
	})

	timetables.Post(":id/:dow", func(c *fiber.Ctx) error {
		c.Accepts(fiber.MIMEApplicationJSON)

		var id uint
		{
			timetableID, err := url.PathUnescape(c.Params("id"))
			if err != nil || timetableID == "" {
				return c.Status(fiber.StatusBadRequest).
					Send([]byte("invalid id provided"))
			}

			tid, err := strconv.Atoi(timetableID)
			if err != nil {
				return c.Status(fiber.StatusBadRequest).
					Send([]byte("invalid id provided"))
			}

			id = uint(tid)
		}

		dow := strings.ToLower(c.Params("dow"))
		switch dow {
		// TODO: maybe use fiber regex for this?
		case "monday", "tuesday", "wednesday", "thursday", "friday",
			"saturday", "sunday":
			// OK
		default:
			return c.SendStatus(fiber.StatusNotFound)
		}

		if len(c.Body()) == 0 {
			return c.Status(fiber.StatusBadGateway).
				Send([]byte("no timetable provided"))
		}

		var newDOW []types.TimetableDay
		if err := json.Unmarshal(c.Body(), &newDOW); err != nil {
			return c.Status(fiber.StatusBadGateway).
				Send([]byte("invalid timetable provided"))
		}

		times := [][2]string{}
		for i := 0; i < len(newDOW); i++ {
			times = append(times, [2]string{newDOW[i].Opening, newDOW[i].Closing})
		}

		createdDow, err := ops.CreateWeekDay(id, database.DOW(dow), times)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).
				Send([]byte(err.Error()))
		}

		return c.Status(fiber.StatusCreated).JSON(createdDow)
	})

	timetables.Delete(":id/:dow", func(c *fiber.Ctx) error {
		var id uint
		{
			timetableID, err := url.PathUnescape(c.Params("id"))
			if err != nil || timetableID == "" {
				return c.Status(fiber.StatusBadRequest).
					Send([]byte("invalid id provided"))
			}

			tid, err := strconv.Atoi(timetableID)
			if err != nil {
				return c.Status(fiber.StatusBadRequest).
					Send([]byte("invalid id provided"))
			}

			id = uint(tid)
		}

		dow := strings.ToLower(c.Params("dow"))
		switch dow {
		// TODO: maybe use fiber regex for this?
		case "monday", "tuesday", "wednesday", "thursday", "friday",
			"saturday", "sunday":
			// OK
		default:
			return c.SendStatus(fiber.StatusNotFound)
		}

		if err := ops.DeleteWeekDay(id, database.DOW(dow)); err != nil {
			return c.Status(fiber.StatusInternalServerError).
				JSON(err)
		}

		return c.SendStatus(fiber.StatusOK)
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
