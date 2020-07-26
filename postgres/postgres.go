package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	// register the mysql driver
	"github.com/cybernostics/cntest"
	docker "github.com/cybernostics/cntest"
	"github.com/cybernostics/cntest/random"
	_ "github.com/lib/pq"
)

// Container creates a mysql opinionated container with defaults overridden by the
// supplied props for:
// db - name of the database. defaults to random
// dbuser - user name for the database. defaults to random
// dbpass - user password for the database. defaults to random
// sql - folder containing sh and sql files executed in lexical order when the container db starts
func Container(props docker.PropertyMap) *cntest.Container {
	return cntest.ContainerWith(Config(props))
}

// Config func to generate a configurer to use like this
// cntest.NewContainer(mysql.Config({db:"mydb",sqlFolder:"./testdb",user:"bob"}))
func Config(props docker.PropertyMap) func(*cntest.Container) error {
	driver := "postgres"
	props["driver"] = driver
	dbName := props.GetOrDefault("db", random.Name())
	props["db"] = dbName
	dbUser := props.GetOrDefault("dbuser", random.Name())
	props["dbuser"] = dbUser
	dbPass := props.GetOrDefault("dbpass", random.Name())
	props["dbpass"] = dbPass
	env := map[string]string{
		"POSTGRES_ALLOW_EMPTY_PASSWORD": "true",
		"POSTGRES_DB":                   dbName,
		"POSTGRES_USER":                 dbUser,
		"POSTGRES_PASSWORD":             dbPass,
	}

	return func(cnt *cntest.Container) error {
		cnt.Props.SetAll(props)
		cnt.AddAllEnv(env)
		cnt.WithImage("postgres")
		cnt.SetAppPort("5432")
		if sqlPath, ok := props["sql"]; ok {
			cnt.AddPathMap(docker.HostPath(sqlPath), cntest.ContainerPath("/docker-entrypoint-initdb.d"))
		}
		cnt.DBConnect = func(timeoutSeconds int) (*sql.DB, error) {
			connStr := fmt.Sprintf("host=%s port=%s user=%s "+
				"password=%s dbname=%s sslmode=disable",
				chk(cnt.IPAddress()), cnt.Port(), dbUser, dbPass, dbName)
			db, err := sql.Open(driver, connStr)

			// if there is an error opening the connection, handle it
			if err != nil {
				return nil, err
			}

			a, cancel := context.WithTimeout(context.Background(), time.Second*time.Duration(timeoutSeconds))
			defer cancel()

			err = db.PingContext(a)
			if err != nil {
				return nil, err
			}
			return db, nil
		}
		cnt.ContainerReady = func() (bool, error) {
			db, err := cnt.DBConnect(1)
			if err != nil {
				if strings.Contains(err.Error(), "connection refused") {
					return false, nil
				}
				fmt.Printf("Error %v\n", err)
				return false, err
			}
			defer db.Close()
			return true, nil
		}
		return nil
	}
}

func chk(val string, err error) string {
	if err != nil {
		panic(err)
	}
	return val
}
