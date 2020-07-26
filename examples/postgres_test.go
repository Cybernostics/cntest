package examples

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

func TestPostgresRunWith(t *testing.T) {
	cnt := postgres.Container(cntest.PropertyMap{"sql": "../fixtures/testschema"})

	cntest.ExecuteWithRunningContainer(t, cnt, func(t *testing.T, c *cntest.Container) {

		// Open up our database connection.
		db, err := c.DBConnect(c.MaxStartTimeSeconds)
		then.AssertThat(t, err, is.Nil())
		defer db.Close()

		err = db.Ping()
		then.AssertThat(t, err, is.Nil())

		// Test some db code
		dbx := sqlx.NewDb(db, c.Props["driver"])

		store := AgentStore{dbx}
		agents, err := store.GetAgents()
		then.AssertThat(t, err, is.Nil())

		for _, agent := range agents {
			fmt.Printf("%v\n", agent)
		}

	})
}
