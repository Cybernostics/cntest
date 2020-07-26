package exmaples

import (
	"fmt"
	"testing"

	"github.com/corbym/gocrest/is"
	"github.com/corbym/gocrest/then"
	"github.com/jmoiron/sqlx"

	// import these to test a postgres container
	"github.com/cybernostics/cntest"
	"github.com/cybernostics/cntest/postgres"
)

// A sample DTO for DB records
type Agent struct {
	AgentCode   string `db:"agent_code"`
	WorkingArea string `db:"working_area"`
	AgentName   string `db:"agent_name"`
	Commission  string `db:"commission"`
	PhoneNo     string `db:"phone_no"`
	Country     string `db:"country"`
}

func TestPostgresRunWith(t *testing.T) {
	cnt := postgres.Container(cntest.PropertyMap{"sql": "../fixtures/testschema"})

	fmt.Printf("Cfg is %v", cnt)
	cntest.ExecuteWithRunningDB(t, cnt, func(t *testing.T, c *cntest.Container) {

		// Container has been created at this point and is ready
		// to accept DB connections.
		fmt.Printf("container : %s\n", c.ContainerName())

		// Open up our database connection.
		db, err := c.DBConnect(c.MaxStartTimeSeconds)

		// if there is an error opening the connection, handle it
		if err != nil {
			panic(err.Error())
		}

		// defer the close
		// executing
		defer db.Close()

		err = db.Ping()
		then.AssertThat(t, err, is.Nil())

		var agents = []Agent{}

		dbx := sqlx.NewDb(db, c.Props["driver"])
		tx := dbx.MustBegin()
		defer tx.Commit()

		err = tx.Select(&agents, "select * from agents")
		then.AssertThat(t, err, is.Nil())

		for _, agent := range agents {
			fmt.Printf("%v\n", agent)
		}

	})
}
