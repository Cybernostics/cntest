package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"strings"
	"time"

	// register the mysql driver
	"github.com/cybernostics/cntest"
	"github.com/cybernostics/cntest/random"

	// if you import the postgres test config you want to test postgres
	_ "github.com/lib/pq"
)

// Container creates a mysql opinionated container with defaults overridden by the
// supplied props for:
// db - name of the database. defaults to random
// dbuser - user name for the database. defaults to random
// dbpass - user password for the database. defaults to random
// sql - folder containing sh and sql files executed in lexical order when the container db starts
func Container(props cntest.PropertyMap) *cntest.Container {
	return cntest.ContainerWith(Config(props))
}

// Config func to generate a configurer to use like this
// cntest.NewContainer(mysql.Config({db:"mydb",sqlFolder:"./testdb",user:"bob"}))
func Config(props cntest.PropertyMap) func(*cntest.Container) error {
	driver := "postgres"
	image  := props.GetOrDefault("image","postgres:13")
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
		cnt.WithImage(image)
		cnt.SetAppPort("5432")
		if sqlPath, ok := props["initdb_path"]; ok {
			if _, err := os.Stat(sqlPath); os.IsNotExist(err) {
				// path/to/whatever does not exist
				return fmt.Errorf("initdb_path does not exist. This should point to the db init scripts")
			}
			cnt.AddPathMap(cntest.HostPath(sqlPath),
				cntest.ContainerPath("/docker-entrypoint-initdb.d"))
		}
		cnt.DBConnect = func(timeoutSeconds int) (*sql.DB, error) {
			connStr := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable",
				dbUser,
				dbPass,
				"127.0.0.1",
				cnt.HostPort(),
				dbName)
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

		dbOKFn1 := cnt.LogsMatch("received fast shutdown request")
		dbOKFn2 := cnt.LogsMatch("database system is ready to accept connections")

		cnt.ContainerReady = func() (bool, error) {

			if running, err := cnt.IsRunning(); !running || err != nil {
				if exited, err := cnt.IsExited(); exited || err != nil {
					return false, err
				} else {
					return false, err
				}
			}
			fmt.Printf("Attempting to connect to DB...")
			db, err := cnt.DBConnect(1)
			if err != nil {
				if strings.Contains(err.Error(), "connection refused") || strings.Contains(err.Error(), "connection reset by peer") {
					fmt.Printf("Connection refused\n")
					return false, nil
				}
				if strings.Contains(err.Error(), "EOF") {
					fmt.Printf("EOF\n")
					return false, nil
				}
				fmt.Printf("Error %v\n%s\n", err, err.Error())
				return false, err
			}
			defer db.Close()

			fmt.Printf("Cheking logs...\n")
			matches, err := dbOKFn1()
			if err != nil {
				fmt.Printf("not matched required log\n")
				return false, err
			}
			if matches {
				matches, err = dbOKFn2()
				if err != nil {
					fmt.Printf("not matched required log\n")
					return false, err
				}

			}
			fmt.Printf("Container reachable\n")
			return matches, nil
		}
		return nil
	}
}
