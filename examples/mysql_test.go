package examples

import (
	"fmt"
	"github.com/corbym/gocrest/then"
	"testing"

	"github.com/corbym/gocrest/is"
	"github.com/jmoiron/sqlx"

	"github.com/cybernostics/cntest"
	"github.com/cybernostics/cntest/mysql"
)

func TestMysqlRunWith(t *testing.T) {

	cntest.PullImage("mysql", "8", cntest.FromDockerHub)

	// This sets up a mysql db server with all the bits randomised
	// you can access them via cnt.Props map. see mysql.Container() method for details.
	cnt := mysql.Container(cntest.PropertyMap{"initdb_path": "../fixtures/testschema"})

	// This wrapper method ensures the container is cleaned up after the test is done
	cntest.ExecuteWithRunningContainer(t, cnt, func(t *testing.T) {
		// Open up our database connection.
		db, err := cnt.DBConnect(cnt.MaxStartTimeSeconds)
		then.AssertThat(t, err, is.Nil())
		defer db.Close()

		// example ping to check connection
		err = db.Ping()
		then.AssertThat(t, err, is.Nil())

		// Test some db code
		dbx := sqlx.NewDb(db, cnt.Props["driver"])

		store := AgentStore{dbx}
		agents, err := store.GetAgents()
		then.AssertThat(t, err, is.Nil())

		for _, agent := range agents {
			fmt.Printf("%v\n", agent)
		}

	})
}
