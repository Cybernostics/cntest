package cntest

import "fmt"

// GroupedContainer container object
// tracks containers on which this depends
type GroupedContainer struct {
	Container *Container
	dependsOn map[string]*GroupedContainer
	started   chan bool
}

// NewGroupedContainer constructor with a container
func NewGroupedContainer(withContainer *Container) *GroupedContainer {
	return &GroupedContainer{
		Container: withContainer,
		dependsOn: map[string]*GroupedContainer{},
		started:   make(chan bool),
	}
}

// ContainerGroup manages a group of connected containers
// like docker-compose but in code
type ContainerGroup map[string]*GroupedContainer

// Add adds a container to the group keyed by its name
func (cg ContainerGroup) Add(cnt *GroupedContainer) {
	cg[cnt.Container.name] = cnt
}

// Start starts all the containers
func (cg ContainerGroup) Start() {
	for _, container := range cg {
		go container.Start()
	}
}

// Await blocks until all containers have started
func (cg ContainerGroup) Await() {
	for _, eachContainer := range cg {
		eachContainer.Await()
	}
}

// Start starts a container once all the dependents start
func (gc *GroupedContainer) Start() {
	for _, depend := range gc.dependsOn {
		depend.Await()
	}
	_, err := gc.Container.Start()
	if err != nil {
		_, _ = fmt.Printf("Error starting container %v: %v", gc.Container, err)
	}
	started, err := gc.Container.AwaitIsReady()
	if err != nil {
		_, _ = fmt.Printf("Error starting container %v: %v", gc.Container, err)
	}
	if started {
		gc.SignalStarted()
	}
}

// SignalStarted signals
func (gc *GroupedContainer) SignalStarted() {
	// multiple tasks reading this channel are unblocked
	close(gc.started)
}

// Await uses a for range on a channel to wait for all dependent containers to start
func (gc *GroupedContainer) Await() {
	for range gc.started {
	}
}

// DependsOn is called to ensure this container wont start until these ones have
func (gc *GroupedContainer) DependsOn(containers ...*GroupedContainer) {
	for _, eachContainer := range containers {
		gc.dependsOn[eachContainer.Container.name] = eachContainer
	}
}
