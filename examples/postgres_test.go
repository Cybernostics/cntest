package examples

import (
	"fmt"
	"testing"

	// import these to test a postgres container
	"github.com/corbym/gocrest/is"
	"github.com/cybernostics/cntest"
	"github.com/cybernostics/cntest/postgres"
	"github.com/jmoiron/sqlx"
)

func TestPostgresRunWith(t *testing.T) {
	cntest.PullImage("postgres", "11", cntest.FromDockerHub)
	cnt := postgres.Container(cntest.PropertyMap{"initdb_path": "../fixtures/testschema"})
	cntest.ExecuteWithRunningContainer(t, cnt, func(t *testing.T, c *cntest.Container) {

		// Open up our database connection.
		db, err := c.DBConnect(c.MaxStartTimeSeconds)
		assertThat(t, err, is.Nil())
		defer db.Close()

		err = db.Ping()
		assertThat(t, err, is.Nil())

		// Test some db code
		dbx := sqlx.NewDb(db, c.Props["driver"])

		store := AgentStore{dbx}
		agents, err := store.GetAgents()
		assertThat(t, err, is.Nil())

		for _, agent := range agents {
			fmt.Printf("%v\n", agent)
		}
	})
}
