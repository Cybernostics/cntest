package cntest

import (
	"bufio"
	"context"
	"database/sql"
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/cybernostics/cntest/random"
	"github.com/cybernostics/cntest/wait"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/mount"
	"github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
)

var api *client.Client

// TCPConnectFn override this to provide a tcp connect function
type TCPConnectFn = func(timeoutSeconds int) (net.Conn, error)

// DBConnectFn override this to provide a DB connect function
type DBConnectFn = func(timeoutSeconds int) (*sql.DB, error)

// ContainerReadyFn should return true when the container is ready to be used
type ContainerReadyFn func() (bool, error)

// HostPort typesafe string subtype
type HostPort string

// ContainerPort typesafe string subtype
type ContainerPort string

// HostPath typesafe string subtype
type HostPath string

// ContainerPath typesafe string subtype
type ContainerPath string

// NOPORT constant for no port specified
const NOPORT = HostPort("")

// PortMap defines a host to container port mapping
type PortMap struct {
	Host      HostPort
	Container ContainerPort
}

// AddTo adds this mapping to a docker Portmap map
func (p PortMap) AddTo(portmap nat.PortMap) {
	portmap[p.Container.Nat()] = []nat.PortBinding{p.Host.Binding()}
}

// Nat does the string formatting for a nat port
func (p ContainerPort) Nat() nat.Port {
	theNat, err := nat.NewPort("tcp", string(p))
	if err != nil {
		panic(err)
	}
	return theNat
}

// Binding Returns a host binding and optionally creates a random one if none provided
func (p HostPort) Binding() nat.PortBinding {
	port := p
	if port == NOPORT {
		listener, err := net.Listen("tcp", ":0")
		defer listener.Close()
		if err != nil {
			panic(err)
		}

		port = HostPort(fmt.Sprintf("%d", listener.Addr().(*net.TCPAddr).Port))
	}
	return nat.PortBinding{
		HostIP:   "0.0.0.0",
		HostPort: string(port),
	}
}

// VolumeMount creates a Volume mount from host to container
type VolumeMount struct {
	Host      HostPath
	Container ContainerPath
}

// Mount is a converter to a docker api mount struct
func (v *VolumeMount) Mount() mount.Mount {
	absMount, _ := filepath.Abs(string(v.Host))
	return mount.Mount{
		Source: absMount,
		Target: string(v.Container),
		Type:   mount.TypeBind,
	}
}

//ImageRefFn This function type provides a hook to generate a custom docker repo path if needed
// You only need this if you're not using the standard docker registry
type ImageRefFn func(image string, version string) string

//API returns current API client or creates on first call
func API() *client.Client {
	if api != nil {
		return api
	}
	api, err := client.NewEnvClient()
	if err != nil {
		fmt.Println("Unable to create docker client")
		panic(err)
	}
	return api
}

// PropertyMap Custom container properties used to configure containers
// They are stored here for later reference eg by tests
// eg They could include db username/password for a db type container
type PropertyMap map[string]string

//Container a simplified API over the docker client API
type Container struct {

	// Props are metadata for a container type. eg dbusername for a db container
	// They are not passed directly to the container environment
	Props PropertyMap

	// This is the main app port for the container
	containerPort ContainerPort

	// These two are used to create the new container using the Docker API
	Config     *container.Config
	HostConfig *container.HostConfig

	// Instance Once the container is started this has all the instance data
	Instance container.ContainerCreateCreatedBody

	// GetImageRepoSource Overload this if you want to get images from a different repo
	GetImageRepoSource ImageRefFn `json:"-,"`

	// will stop and remove the container after testing
	StopAfterTest bool
	// RemoveAfterTest remove the container after the test
	RemoveAfterTest bool
	// The docker image to use for the container
	dockerImage string
	// The Maximum time to wait for the container
	MaxStartTimeSeconds int
	// The name of the running container
	name string

	// NamePrefix is combined with a random suffix to create the containerName
	// if the name is blank when the container is started
	NamePrefix string
	// The environment
	environment map[string]string
	// The IP address of the container
	iP string

	// ContainerReady Override this if you want to check more than an ok response to a TCP connect
	ContainerReady ContainerReadyFn `json:"-,"`

	// TCPConnect fn to connect to a TCP connection
	TCPConnect TCPConnectFn `json:"-,"`

	// DBConnect fn to connect to a DB connection
	DBConnect DBConnectFn `json:"-,"`
}

// SetIfMissing sets the value if it isn't already
func (p PropertyMap) SetIfMissing(key string, value string) {
	if val, ok := p[key]; !ok {
		p[key] = val
	}
}

// GetOrDefault returns the specified key value or a default
func (p PropertyMap) GetOrDefault(key string, defaultValue string) string {
	if val, ok := p[key]; ok {
		return val
	}
	return defaultValue
}

// SetAll sets all the props on the provided map
func (p PropertyMap) SetAll(amap PropertyMap) {
	for key, val := range amap {
		p[key] = val
	}
}

// FromDockerHub is the default formatter for a docker image resource
// provide your own for private repos
func FromDockerHub(image string, version string) string {
	if version == "" {
		version = "latest"
	}
	return fmt.Sprintf("docker.io/library/%s:%s", image, version)
}

// ContainerConfigFn custom function that configures a container
type ContainerConfigFn func(c *Container) error

// NewContainer constructor fn
func NewContainer() *Container {

	cnt := Container{
		Props:               PropertyMap{},
		Config:              &container.Config{Tty: true},
		HostConfig:          &container.HostConfig{},
		GetImageRepoSource:  FromDockerHub,
		MaxStartTimeSeconds: 30,
		StopAfterTest:       true,
		RemoveAfterTest:     true,
	}
	cnt.ContainerReady = cnt.IsRunning
	return &cnt

}

// ContainerWith contstructor which takes a custom configurer
func ContainerWith(fn ContainerConfigFn) *Container {
	cnt := NewContainer()
	fn(cnt)
	return cnt
}

// PullImage like docker pull cmd
func PullImage(image string, version string, getRepoFn ImageRefFn) {
	reader, err := API().ImagePull(context.Background(), getRepoFn(image, version), types.ImagePullOptions{})
	if err != nil {
		panic(err)
	}
	io.Copy(os.Stdout, reader)
}

// SetName container name
func (c *Container) SetName(name string) {
	c.name = name
}

// SetAppPort sets the main port for the container if it has one
func (c *Container) SetAppPort(port string) *Container {
	return c.SetPort(port, "")
}

// SetPort sets the main port to be used by the container
// Other port mappings can be added but this one is considered the big kahuna
// for checking readiness for example
func (c *Container) SetPort(port string, mappedHostPort string) *Container {
	c.containerPort = ContainerPort(port)
	if len(mappedHostPort) == 0 {
		c.MapToRandomHostPort(c.containerPort)
	} else {
		c.AddPortMap(HostPort(mappedHostPort), c.containerPort)
	}
	return c
}

// MapToRandomHostPort like --ports cmd switch for mapping ports
func (c *Container) MapToRandomHostPort(containerPort ContainerPort) {
	c.addPortBinding(PortMap{Container: containerPort, Host: NOPORT})
	c.AddExposedPort(containerPort)
}

// AddPathMap like -v cmd switch for mapping paths
func (c *Container) AddPathMap(Host HostPath, Container ContainerPath) {
	pathMap := VolumeMount{Host, Container}
	if c.HostConfig.Mounts == nil {
		c.HostConfig.Mounts = make([]mount.Mount, 0)
	}
	c.HostConfig.Mounts = append(c.HostConfig.Mounts, pathMap.Mount())
}

// AddPortMap like -p cmd line switch for adding port mappings
func (c *Container) AddPortMap(host HostPort, container ContainerPort) {
	c.addPortBinding(PortMap{host, container})
}

func (c *Container) addPortBinding(portMap PortMap) {
	if c.HostConfig.PortBindings == nil {
		c.HostConfig.PortBindings = make(nat.PortMap)
	}
	if portMap.Host == NOPORT {

	}
	portMap.AddTo(c.HostConfig.PortBindings)
}

// AddExposedPort expose a container port
func (c *Container) AddExposedPort(port ContainerPort) {
	if c.Config.ExposedPorts == nil {
		c.Config.ExposedPorts = make(nat.PortSet)
	}
	c.Config.ExposedPorts[port.Nat()] = struct{}{}
}

// AddAllEnv adds the mapped values to the config
func (c *Container) AddAllEnv(aMap map[string]string) {
	for key, val := range aMap {
		c.AddEnv(key, val)
	}
}

// AddEnv adds the value to the config
func (c *Container) AddEnv(key, value string) {
	c.Config.Env = append(c.Config.Env, fmt.Sprintf("%s=%s", key, value))
}

// Start starts the container
func (c *Container) Start() (string, error) {

	c.name = c.ContainerName()

	instance, err := API().ContainerCreate(
		context.Background(),
		c.Config,
		c.HostConfig,
		nil, c.name)

	if err != nil {
		return "", err
	}

	c.Instance = instance
	err = API().ContainerStart(context.Background(), c.Instance.ID, types.ContainerStartOptions{})
	if err != nil {
		return "", err
	}

	c.iP, err = c.InspectIPAddress()
	if err != nil {
		return "", err
	}

	fmt.Printf("Container %s is starting\n", c.Instance.ID)

	if len(c.Instance.Warnings) > 0 {
		for _, warning := range c.Instance.Warnings {
			fmt.Println(warning)
		}
	}

	return c.Instance.ID, nil
}

// ContainerName returns the generated name for the container
func (c *Container) ContainerName() string {

	if len(c.name) == 0 {
		c.name = c.NamePrefix + "-" + random.Name()
	}
	return c.name
}

// LogsMatch returns a fn matcher for bool wait fns
func (c *Container) LogsMatch(pattern string) func() (bool, error) {
	var logPattern = regexp.MustCompile(pattern)
	return func() (bool, error) {
		logsOptions := types.ContainerLogsOptions{
			Details:    true,
			ShowStderr: true,
			ShowStdout: true,
			Until:      "all",
		}

		if logsReader, err := API().ContainerLogs(context.Background(), c.Instance.ID, logsOptions); err != nil {
			bufferedLogs := bufio.NewReader(logsReader)
			defer logsReader.Close()
			for {
				line, error := bufferedLogs.ReadString('\n')
				fmt.Printf("%s:%s", c.name, line)
				if error == io.EOF {
					break
				} else if error != nil {
					return false, nil
				}
				if logPattern.MatchString(line) {
					return true, nil
				}
			}
		}

		return false, nil
	}
}

// Logs returns the container logs as a string or error
func (c *Container) Logs() (string, error) {

	logsOptions := types.ContainerLogsOptions{
		Details:    true,
		ShowStderr: true,
		ShowStdout: true,
		Tail:       "all",
	}

	logsReader, err := API().ContainerLogs(context.Background(), c.Instance.ID, logsOptions)
	if err != nil {
		return "", err
	}
	defer logsReader.Close()
	buf := new(strings.Builder)
	_, err = io.Copy(buf, logsReader)
	// check errors
	return buf.String(), err
}

// AwaitLogPattern waits for the container to start based on expected log message patterns
func (c *Container) AwaitLogPattern(timeoutSeconds int, patternRegex string) (started bool, err error) {
	return wait.UntilTrue(timeoutSeconds, c.LogsMatch(patternRegex))
}

// AwaitIsRunning waits for the container is in the running state
func (c *Container) AwaitIsRunning() (started bool, err error) {
	return wait.UntilTrue(c.MaxStartTimeSeconds, c.IsRunning)
}

// AwaitIsReady waits for the container is in the running state
func (c *Container) AwaitIsReady() (started bool, err error) {
	return wait.UntilTrue(c.MaxStartTimeSeconds, c.ContainerReady)
}

// IPAddress retrieve the IP address of the running container
func (c *Container) IPAddress() (string, error) {
	if len(c.iP) != 0 {
		return c.iP, nil
	}
	ip, err := c.InspectIPAddress()
	if err != nil {
		return "", err
	}
	c.iP = ip
	return c.iP, nil
}

// InspectIPAddress uses docker inspect to find out the ip address
func (c *Container) InspectIPAddress() (string, error) {
	inspect, err := API().ContainerInspect(context.Background(), c.Instance.ID)
	if err != nil {
		return "", err
	}
	return inspect.NetworkSettings.IPAddress, nil
}

// Port returns the containers port
func (c *Container) Port() ContainerPort {
	return c.containerPort
}

// ConnectTCP Connects tot he container port using a TCP connection
func (c *Container) ConnectTCP(timeoutSeconds int) (net.Conn, error) {
	var port string
	for key := range c.Config.ExposedPorts {
		port = key.Port()
	}

	ip, err := c.IPAddress()
	if err != nil {
		return nil, err
	}

	host := fmt.Sprintf("%s:%s", ip, port)
	var timeout time.Duration
	if timeoutSeconds == 0 {
		timeout = time.Duration(10) * time.Minute
	} else {
		timeout = time.Duration(timeoutSeconds) * time.Minute
	}
	for {
		var conn net.Conn
		conn, err = net.DialTimeout("tcp", host, timeout)
		if err != nil {
			return nil, err
		}
		return conn, nil
	}

}

// Check if tcp port is open
func (c *Container) Check(timeoutSeconds int) (ret bool, err error) {

	conn, err := c.ConnectTCP(timeoutSeconds)
	if err != nil {
		return false, nil
	}
	defer func() {
		conn.Close()
	}()
	return true, nil
}

// Stop stops the container
func (c *Container) Stop(timeoutSeconds int) (ok bool, err error) {
	err = API().ContainerStop(context.Background(), c.Instance.ID, nil)
	if err != nil {
		return
	}
	if timeoutSeconds > 0 {
		return c.AwaitExit(timeoutSeconds)
	}
	return true, nil
}

// AwaitExit waits for the container to stop
func (c *Container) AwaitExit(timeoutSeconds int) (ok bool, err error) {
	return wait.UntilTrue(timeoutSeconds, c.IsExited)
}

// RunCmd execs the specified command and args on the container
func (c *Container) RunCmd(cmd []string) (io.Reader, error) {
	cmdConfig := types.ExecConfig{AttachStdout: true, AttachStderr: true,
		Cmd: cmd,
	}
	ctx := context.Background()
	execID, _ := API().ContainerExecCreate(ctx, c.Instance.ID, cmdConfig)
	fmt.Println(execID)

	res, err := API().ContainerExecAttach(ctx, execID.ID, types.ExecStartCheck{})
	if err != nil {
		return nil, err
	}

	err = API().ContainerExecStart(ctx, execID.ID, types.ExecStartCheck{})
	if err != nil {
		return nil, err
	}

	return res.Reader, nil
}

// Remove deletes the container permanently
func (c *Container) Remove() error {
	return API().ContainerRemove(context.Background(), c.Instance.ID, types.ContainerRemoveOptions{Force: true})
}

// IsRemoveAfterTest true if the container should be removed
func (c *Container) IsRemoveAfterTest() bool {
	return c.RemoveAfterTest
}

// IsRunning returns true if the container is in the started state
// Will error if the container has already exited
func (c *Container) IsRunning() (started bool, err error) {
	inspect, err := API().ContainerInspect(context.Background(), c.Instance.ID)
	status := inspect.State.Status
	if err != nil {
		return false, err
	}
	if status == "running" {
		return true, nil
	}
	if status == "exited" {
		return false, errors.New("Container Already exited")
	}
	fmt.Printf("Status is %s\n", status)
	return false, nil
}

//IsExited returns true if the container has exited
func (c *Container) IsExited() (started bool, err error) {
	inspect, err := API().ContainerInspect(context.Background(), c.Instance.ID)
	if err != nil {
		return false, err
	}
	if inspect.State.Status == "exited" {
		return true, nil
	}
	return false, nil
}

// WithImage sets the image name and the default container name prefix
func (c *Container) WithImage(image string) *Container {
	c.Config.Image = image
	c.NamePrefix = image
	return c
}

// IsStopAfterTest if true then stop the container after the test
func (c *Container) IsStopAfterTest() bool {
	return c.StopAfterTest
}
