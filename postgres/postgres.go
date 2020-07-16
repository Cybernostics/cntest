package postgres

import (
	"fmt"
	// register the mysql driver
	_ "github.com/lib/pq"
	"github.com/wjase/dbcontainers"
)

//Apply A configurer which configures a PostgresHost and connection string
func Apply(c *dbcontainers.DBContainer) *dbcontainers.DBContainer {

	env := map[string]string{
		"POSTGRES_ALLOW_EMPTY_PASSWORD": "true",
		"POSTGRES_DB":                   c.DatabaseName,
		"POSTGRES_USER":                 c.User,
		"POSTGRES_PASSWORD":             c.Password,
	}
	c.Environment = env
	c.DockerImage = "postgres"

	c.ConnectionStringFn = func() string {
		return fmt.Sprintf("host=%s port=%s user=%s "+
			"password=%s dbname=%s sslmode=disable",
			c.IP, c.Port, c.User, c.Password, c.DatabaseName)
	}
	return c.WithDriverType("postgres").WithPort("5432")
}
