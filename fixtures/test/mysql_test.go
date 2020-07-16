package test

import (
	"database/sql"
	"fmt"
	"testing"

	"github.com/corbym/gocrest/is"
	"github.com/corbym/gocrest/then"
	"github.com/jmoiron/sqlx"

	"github.com/wjase/dbcontainers"
	"github.com/wjase/dbcontainers/docker"
	"github.com/wjase/dbcontainers/mysql"
)

func TestMain(m *testing.M) {
	// pull the image before you start testing so you don't blow your timeout
	docker.PullImage("mysql", "latest", docker.FromDockerHub)
	docker.PullImage("postgres", "latest", docker.FromDockerHub)
	m.Run()
}

func TestMysqlRunWith(t *testing.T) {
	cnt := dbcontainers.
		NewDBContainer().
		WithConfigurer(mysql.Apply).
		WithSchemaFolder("../testschema")

	fmt.Printf("Cfg is %v", cnt)
	dbcontainers.ExecuteWithRunningDB(t, cnt, func(t *testing.T, c *dbcontainers.DBContainer) {

		// Container has been created at this point and is ready
		// to accept DB connections.
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

		var agents = []Agent{}

		dbx := sqlx.NewDb(db, c.DriverType)
		tx := dbx.MustBegin()
		defer tx.Commit()

		err = tx.Select(&agents, "select * from agents")
		then.AssertThat(t, err, is.Nil())

		for _, agent := range agents {
			fmt.Printf("%v", agent)
		}

	})
}
