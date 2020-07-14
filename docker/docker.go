package docker

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"regexp"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/mount"
	"github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
	"github.com/wjase/dbcontainers/wait"
)

var api *client.Client

type HostPort string
type ContainerPort string
type HostPath string
type ContainerPath string

const NOPORT = HostPort("")

type PortMap struct {
	Host      HostPort
	Container ContainerPort
}

func (p PortMap) AddTo(portmap nat.PortMap) {
	portmap[p.Container.Nat()] = []nat.PortBinding{p.Host.Binding()}
}

func (p ContainerPort) Nat() nat.Port {
	theNat, err := nat.NewPort("tcp", string(p))
	if err != nil {
		panic(err)
	}
	return theNat
}

func (p HostPort) Binding() nat.PortBinding {
	port := p
	if port == NOPORT {
		listener, err := net.Listen("tcp", ":0")
		if err != nil {
			panic(err)
		}

		port = HostPort(fmt.Sprintf("%d", listener.Addr().(*net.TCPAddr).Port))
	}
	return nat.PortBinding{
		HostIP:   "0.0.0.0",
		HostPort: string(p),
	}
}

type VolumeMount struct {
	Host      HostPath
	Container ContainerPath
}

func (v *VolumeMount) Mount() mount.Mount {
	return mount.Mount{
		Source: string(v.Host),
		Target: string(v.Container),
		Type:   mount.TypeBind,
	}
}

//ImageRefFn This function type provides a hook to generate a custome docker repo path if needed
// You only need this if you're not using the standard docker registry
type ImageRefFn func(string) string

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

//Container a simplified API over the docker client API
type Container struct {
	// This is the name of the running container
	Name string
	// These two are used to create the new container
	Config     *container.Config
	HostConfig *container.HostConfig
	// Once the container is started this has all the instance data
	Instance container.ContainerCreateCreatedBody
	// Overload this if you want to get images from a different repo
	GetImageRef ImageRefFn
}

func NewContainer(image string) *Container {
	return &Container{
		Config:     &container.Config{Image: image, Tty: true},
		HostConfig: &container.HostConfig{},
		GetImageRef: func(image string) string {
			return fmt.Sprintf("docker.io/library/%s:latest", image)
		},
	}
}

func (c *Container) SetName(name string) {
	c.Name = name
}

func (c *Container) AddPathMap(pathMap VolumeMount) {
	if c.HostConfig.Mounts == nil {
		c.HostConfig.Mounts = make([]mount.Mount, 0)
	}
	c.HostConfig.Mounts = append(c.HostConfig.Mounts, pathMap.Mount())
}

func (c *Container) AddPortMap(portMap PortMap) {
	if c.HostConfig.PortBindings == nil {
		c.HostConfig.PortBindings = make(nat.PortMap)
	}
	portMap.AddTo(c.HostConfig.PortBindings)
}

func (c *Container) AddExposedPort(port ContainerPort) {
	if c.Config.ExposedPorts == nil {
		c.Config.ExposedPorts = make(nat.PortSet)
	}
	c.Config.ExposedPorts[port.Nat()] = struct{}{}
}

func (c *Container) AddAllEnv(aMap map[string]string) {
	for key, val := range aMap {
		c.AddEnv(key, val)
	}
}

func (c *Container) AddEnv(key, value string) {
	c.Config.Env = append(c.Config.Env, fmt.Sprintf("%s=%s", key, value))
}

func (c *Container) Start() (string, error) {

	//docker.io/library/%s:latest
	reader, err := API().ImagePull(context.Background(), c.GetImageRef(c.Config.Image), types.ImagePullOptions{})
	if err != nil {
		panic(err)
	}
	io.Copy(os.Stdout, reader)

	c.Instance, err = API().ContainerCreate(
		context.Background(),
		c.Config,
		c.HostConfig,
		nil, c.Name)

	if err != nil {
		panic(err)
	}

	API().ContainerStart(context.Background(), c.Instance.ID, types.ContainerStartOptions{})
	fmt.Printf("Container %s is starting\n", c.Instance.ID)

	if len(c.Instance.Warnings) > 0 {
		for _, warning := range c.Instance.Warnings {
			fmt.Println(warning)
		}
	}

	return c.Instance.ID, nil
}

func (c *Container) logsMatch(pattern string) func() (bool, error) {
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
				fmt.Printf("%s:%s", c.Name, line)
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

func (c *Container) AwaitLogPattern(timeoutSeconds int, patternRegex string) (started bool, err error) {
	return wait.UntilTrue(timeoutSeconds, c.logsMatch(patternRegex))
}

func (c *Container) AwaitStartup(timeoutSeconds int) (started bool, err error) {
	return wait.UntilTrue(timeoutSeconds, c.IsStarted)
}

func (c *Container) IPAddress() (string, error) {
	inspect, err := API().ContainerInspect(context.Background(), c.Instance.ID)
	if err != nil {
		return "", err
	}
	return inspect.NetworkSettings.IPAddress, nil
}

// Check if tcp port is open
func (c *Container) Check(timeoutSeconds int) (ret bool, err error) {
	var port string
	for key := range c.Config.ExposedPorts {
		port = key.Port()
	}

	ip, err := c.IPAddress()
	if err != nil {
		return false, err
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
			ret = true
			err = nil
			return
		}
		if conn != nil {
			err = conn.Close()
			return
		}
	}
}

func (c *Container) Stop(timeoutSeconds int) (ok bool, err error) {
	err = API().ContainerStop(context.Background(), c.Instance.ID, nil)
	if err != nil {
		return
	}
	if timeoutSeconds > 0 {
		return c.AwaitCompletion(timeoutSeconds)
	}
	return true, nil
}

func (c *Container) AwaitCompletion(timeoutSeconds int) (ok bool, err error) {
	return wait.UntilTrue(timeoutSeconds, c.IsExited)
}

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

func (c *Container) Remove() error {
	return API().ContainerRemove(context.Background(), c.Instance.ID, types.ContainerRemoveOptions{Force: true})
}

//IsStarted returns true if the container is in the started state
// Will error if the container has already exited
func (c *Container) IsStarted() (started bool, err error) {
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

//IsStarted returns true if the container is in the started state
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
