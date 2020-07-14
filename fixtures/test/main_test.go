package test

import (
	"database/sql"
	"fmt"
	"testing"

	_ "github.com/go-sql-driver/mysql"
	"github.com/wjase/dbcontainers"
	"github.com/wjase/dbcontainers/mysql"
)

func TestMysqlRunWith(t *testing.T) {
	cnt := dbcontainers.
		NewDBContainer().
		WithConfigurer(mysql.Apply).
		WithSchemaFolder("../testschema")

	fmt.Printf("Cfg is %v", cnt)
	dbcontainers.ExecuteWithRunningDB(t, cnt, func(t *testing.T, c *dbcontainers.DBContainer) {
		fmt.Printf("container : %s\n", c.ContainerName())

		// Open up our database connection.
		fmt.Printf("Connecting to %s", c.ConnectionStringFn())
		db, err := sql.Open(c.DriverType, c.ConnectionStringFn())

		// if there is an error opening the connection, handle it
		if err != nil {
			panic(err.Error())
		}

		// defer the close till after the main function has finished
		// executing
		defer db.Close()

		err = db.Ping()
		if err != nil {
			panic(err.Error())
		}
	})
}
