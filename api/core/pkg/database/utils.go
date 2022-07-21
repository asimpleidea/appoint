package database

import (
	"fmt"
	"strconv"
	"strings"
)

func GeneratePostgresDSN(opts Options) string {
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
