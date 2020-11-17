package mysql

import (
	"database/sql"
	"fmt"
	"strings"
	"time"

	// register the mysql driver
	"github.com/cybernostics/cntest"
	"github.com/cybernostics/cntest/random"

	// if you import the mysql test config you want to test mysql
	_ "github.com/go-sql-driver/mysql"
	"golang.org/x/net/context"
)

// Container creates a mysql opinionated container with defaults overridden by the
// supplied props for:
// 		db - name of the database. defaults to random
// 		dbuser - user name for the database. defaults to random
// 		dbpass - user password for the database. defaults to random
// 		initdb_path - folder containing sh and sql files executed in lexical order when the container db starts
func Container(props cntest.PropertyMap) *cntest.Container {
	return cntest.ContainerWith(Config(props))
}

// Config func to generate a configurer to use like this
// cntest.NewContainer(mysql.Config({"db":"mydb","initdb_path":"./testdb",user:"bob"}))
func Config(props cntest.PropertyMap) func(*cntest.Container) error {
	driver := "mysql"
	props["driver"] = driver
	dbName := props.GetOrDefault("db", random.Name())
	props["db"] = dbName
	dbUser := props.GetOrDefault("dbuser", random.Name())
	props["dbuser"] = dbUser
	dbPass := props.GetOrDefault("dbpass", random.Name())
	props["dbpass"] = dbPass

	env := map[string]string{
		"MYSQL_ALLOW_EMPTY_PASSWORD": "true",
		"MYSQL_DATABASE":             dbName,
		"MYSQL_USER":                 dbUser,
		"MYSQL_PASSWORD":             dbPass,
	}

	return func(cnt *cntest.Container) error {
		cnt.Props.SetAll(props)
		cnt.AddAllEnv(env)
		cnt.WithImage("mysql")
		cnt.SetAppPort("3306")
		if sqlPath, ok := props["initdb_path"]; ok {
			cnt.AddPathMap(cntest.HostPath(sqlPath), cntest.ContainerPath("/docker-entrypoint-initdb.d"))
		}
		cnt.DBConnect = func(timeoutSeconds int) (*sql.DB, error) {
			connStr := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s", dbUser, dbPass, "127.0.0.1", cnt.HostPort(), dbName)
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
			db, err := cnt.DBConnect(5)
			if err != nil {
				if strings.Contains(err.Error(), "connection refused") {
					return false, nil
				}
				if strings.Contains(err.Error(), "EOF") {
					return false, nil
				}
				if strings.Contains(err.Error(), "driver: bad connection") {
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
