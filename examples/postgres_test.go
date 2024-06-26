package examples

import (
	"fmt"
	"github.com/corbym/gocrest/then"
	"testing"

	// import these to test a postgres container
	"github.com/corbym/gocrest/is"
	"github.com/cybernostics/cntest"
	"github.com/cybernostics/cntest/postgres"
	"github.com/jmoiron/sqlx"
)

func TestPostgresRunWith(t *testing.T) {
	cntest.PullImage("postgres", "13", cntest.FromDockerHub)
	cnt := postgres.Container(cntest.PropertyMap{"initdb_path": "../fixtures/testschema"})
	cnt.RemoveAfterTest=false
	cnt.StopAfterTest=false
	
	cntest.ExecuteWithRunningContainer(t, cnt, func(t *testing.T) {

		// Open up our database connection.
		db, err := cnt.DBConnect(cnt.MaxStartTimeSeconds)
		then.AssertThat(t, err, is.Nil())
		defer db.Close()

		err = db.Ping()
		then.AssertThat(t, err, is.Nil())

		// Test some db code
		dbx := sqlx.NewDb(db, cnt.Props["driver"])

		store := AgentStore{dbx}
		agents, err := store.GetAgents()
		then.AssertThat(t, err, is.Nil())

		for _, agent := range agents {
			fmt.Printf("%v yo\n", agent)
		}
	})
}
