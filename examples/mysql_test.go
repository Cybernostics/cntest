package exmaples

import (
	"fmt"
	"testing"

	"github.com/corbym/gocrest/is"
	"github.com/corbym/gocrest/then"
	"github.com/jmoiron/sqlx"

	"github.com/cybernostics/cntest"
	"github.com/cybernostics/cntest/mysql"
)

func TestMysqlRunWith(t *testing.T) {

	// This sets up a mysql db server with all the bits randomised
	// you can access them via cnt.Props map. see mysql.Container() method for details.
	cnt := mysql.Container(cntest.PropertyMap{"sql": "../fixtures/testschema"})

	// This wrapper method ensures the container is cleaned up after the test is done
	cntest.ExecuteWithRunningDB(t, cnt, func(t *testing.T, c *cntest.Container) {

		// Open up our database connection.
		db, err := c.DBConnect(c.MaxStartTimeSeconds)

		// if there is an error opening the connection, handle it
		if err != nil {
			panic(err.Error())
		}

		// defer the close
		defer db.Close()

		// example ping to check connection
		err = db.Ping()
		then.AssertThat(t, err, is.Nil())

		// this is some sample db code using sqlx
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
