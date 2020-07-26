# cntest
[![Go Report Card](https://goreportcard.com/badge/github.com/cybernostics/cntest?style=flat-square)](https://goreportcard.com/report/github.com/cybernostics/cntest)
[![Go Doc](https://img.shields.io/badge/godoc-reference-blue.svg?style=flat-square)](http://godoc.org/github.com/cybernostics/cntest)
[![Release](https://img.shields.io/github/release/cybernostics/cntest.svg?style=flat-square)](https://github.com/cybernostics/cntest/releases/latest)

Docker container-based testing and db container testing made easy.

## Background
CnTest Tries to take most of the boilerplate away in creating containers.
Typically you would have a combination of scripts and code to start a container and then wait for it to
start, or run to completion or reach a point where it logs something, before running a test.

CnTest allows you to do all of that within the context of a go test.
This includes cleaning up after itself so you dont have a bunch of zombie containers left around.
It also uses random identifiers for things like dbnames and usernames so you don't have to manually
configure them, and also this means that you can spin up multiple db hosts for example to test against.

# Usage

```golang
// from examples/mysql_test.go

let generateProject = project => {
  func TestMysqlRunWith(t *testing.T) {

	// This sets up a mysql db server with all the bits randomised
	// you can access them via cnt.Props map. see mysql.Container() method for details.
	cnt := mysql.Container(cntest.PropertyMap{"sql": "../path/to/folder/containing/your/init/sql"})

	// This wrapper method ensures the container is cleaned up after the test is done
	cntest.ExecuteWithRunningContainer(t, cnt, func(t *testing.T, c *cntest.Container) {

		// Open up our database connection.
		db, err := c.DBConnect(c.MaxStartTimeSeconds)
		// if err...
		defer db.Close()

		// example ping to check connection
		err = db.Ping()
		// if err...

		// Use sql.db in your tests 

	})
}
```

See more examples/ for examples of tests for 
 * a hello world container
 * a mysql db
 * a postgres db

