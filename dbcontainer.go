package dbcontainers

import (
	"database/sql"
	"fmt"
	"strings"
	"time"

	"path/filepath"

	"github.com/goombaio/namegenerator"
	"github.com/wjase/dbcontainers/docker"
	"github.com/wjase/dbcontainers/wait"
)

var seed = time.Now().UTC().UnixNano()
var nameMaker = namegenerator.NewNameGenerator(seed)

// ConnectionStringFn overridden for each db type
type ConnectionStringFn func() string

// Configurer function for configuring the DBContainer
type Configurer func(config *DBContainer) *DBContainer

//DBContainer contains the params for creating a new test db container
type DBContainer struct {
	// will stop and remove the container after testing
	StopAfterTest bool
	// RemoveAfterTest remove the container after the test
	RemoveAfterTest bool
	// The docker image to use for the container
	DockerImage string
	// If not blank this is used to map to /docker-entrypoint-initdb.d/
	SchemaFolder string
	// The driver identifier
	DriverType string
	// The port to contact the database
	Port string
	// The database name to create
	DatabaseName string
	// The User permitted to access the database
	User string
	// The Password used to access the database
	Password string
	// The Maximum time to wait for the container
	MaxStartTimeSeconds int
	// The name of the running container
	name string
	// The environment
	Environment map[string]string
	// The IP address of the container
	IP string

	ConnectionStringFn ConnectionStringFn
}

// NewDBContainer creates a template config
func NewDBContainer() *DBContainer {
	cfg := &DBContainer{}
	cfg.MaxStartTimeSeconds = 30
	cfg.User = randomName()
	cfg.Password = randomName()
	cfg.DatabaseName = randomName()
	cfg.StopAfterTest = true
	cfg.RemoveAfterTest = true
	return cfg
}

func randomName() string {
	return strings.Replace(nameMaker.Generate(), "-", "", 1)
}

// Start the docker container
func (c *DBContainer) Start() (*docker.Container, error) {
	dc := docker.NewContainer(c.DockerImage)
	dc.AddPathMap(docker.VolumeMount{Host: docker.HostPath(c.SchemaFolder), Container: docker.ContainerPath("/docker-entrypoint-initdb.d")})
	dbport := docker.ContainerPort(c.Port)
	dc.AddPortMap(docker.PortMap{Container: dbport})
	dc.AddExposedPort(dbport)
	dc.AddAllEnv(c.Environment)
	dc.Name = c.ContainerName()
	dc.Start()
	ip, err := dc.IPAddress()
	if err != nil {
		return nil, err
	}
	c.IP = ip
	_, err = wait.UntilTrue(c.MaxStartTimeSeconds, c.CanConnect)
	if err != nil {
		fmt.Printf("Failed to reach container on startup. Forcing stop")
		dc.Stop(5)
		return nil, err
	}
	return dc, nil
}

// CanConnect tries to ping a db connection to confirm it is ready to accept connections
func (c *DBContainer) CanConnect() (bool, error) {
	db, err := sql.Open(c.DriverType, c.ConnectionStringFn())

	// if there is an error opening the connection, handle it
	if err != nil {
		return false, err
	}

	// defer the close till after the main function has finished
	// executing
	defer db.Close()

	err = db.Ping()
	if err != nil {
		if strings.Contains(err.Error(), "connection refused") {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

// ContainerName returns the generated name for the container
func (c *DBContainer) ContainerName() string {
	if len(c.name) == 0 {
		c.name = c.DriverType + "-" + randomName()
	}
	return c.name
}

// WithSchemaFolder - maps to /docker-entrypoint-initdb.d/ in the container
func (c *DBContainer) WithSchemaFolder(schemaFolder string) *DBContainer {
	if filepath.IsAbs(schemaFolder) {
		c.SchemaFolder = schemaFolder
	} else {
		folder, err := filepath.Abs(schemaFolder)
		if err != nil {
			panic(err)
		}
		c.SchemaFolder = folder

	}
	return c
}

// WithDriverType - string to specify the db driver type
func (c *DBContainer) WithDriverType(driverType string) *DBContainer {
	c.DriverType = driverType
	return c
}

// WithStopAfterTest - if true the container is stopped after the test
// (temporarily set to false when looping round tests)
func (c *DBContainer) WithStopAfterTest(stopAfterTest bool) *DBContainer {
	c.StopAfterTest = stopAfterTest
	return c
}

// WithRemoveAfterTest - if true the container is removed after the test
// (implies stop after test if true)
func (c *DBContainer) WithRemoveAfterTest(removeAfterTest bool) *DBContainer {
	c.RemoveAfterTest = removeAfterTest
	return c
}

// WithPort - specify the DB port to use
func (c *DBContainer) WithPort(port string) *DBContainer {
	c.Port = port
	return c
}

// WithDockerImage - specify the docker image to use
func (c *DBContainer) WithDockerImage(image string) *DBContainer {
	c.DockerImage = image
	return c
}

// WithDatabaseName - specify the DB name to create in the container
func (c *DBContainer) WithDatabaseName(databaseName string) *DBContainer {
	c.DatabaseName = databaseName
	return c
}

// WithUser - username to use to contact the DB
func (c *DBContainer) WithUser(user string) *DBContainer {
	c.User = user
	return c
}

// WithPassword - password to access the database
func (c *DBContainer) WithPassword(password string) *DBContainer {
	c.Password = password
	return c
}

// WithMaxStartTimeSeconds - Wait at most this long for the database container to start
func (c *DBContainer) WithMaxStartTimeSeconds(maxStartTimeSeconds int) *DBContainer {
	c.MaxStartTimeSeconds = maxStartTimeSeconds
	return c
}

// WithContainerName - The name of the container once created.
// If not specified a random one is generated.
func (c *DBContainer) WithContainerName(containerName string) *DBContainer {
	c.name = containerName
	return c
}

// WithEnvironment - map of environment settings
func (c *DBContainer) WithEnvironment(environment map[string]string) *DBContainer {
	c.Environment = environment
	return c
}

// WithConfigurer - Allows you to extend the configuration of containers or use one of the prepaked ones
func (c *DBContainer) WithConfigurer(fn Configurer) *DBContainer {
	return fn(c)
}
