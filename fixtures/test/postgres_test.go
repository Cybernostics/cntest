package test

import (
	"database/sql"
	"fmt"
	"testing"

	"github.com/corbym/gocrest/is"
	"github.com/corbym/gocrest/then"
	"github.com/jmoiron/sqlx"

	"github.com/wjase/dbcontainers"
	"github.com/wjase/dbcontainers/postgres"
)

type Agent struct {
	AgentCode   string `db:"agent_code"`
	WorkingArea string `db:"working_area"`
	AgentName   string `db:"agent_name"`
	Commission  string `db:"commission"`
	PhoneNo     string `db:"phone_no"`
	Country     string `db:"country"`
}

func TestPostgresRunWith(t *testing.T) {
	cnt := dbcontainers.
		NewDBContainer().
		WithConfigurer(postgres.Apply).
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

		var agents = []Agent{}

		dbx := sqlx.NewDb(db, c.DriverType)
		tx := dbx.MustBegin()
		err = tx.Select(&agents, "select * from agents")
		then.AssertThat(t, err, is.Nil())

		if err == nil {
			err = tx.Commit()
		}

		then.AssertThat(t, err, is.Nil())
		for _, agent := range agents {
			fmt.Printf("%v", agent)
		}

	})
}
