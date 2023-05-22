package whip

import (
	"context"
	"log"
	"sync"

	"github.com/google/uuid"
	"github.com/pion/webrtc/v3"
)

type Resource struct {
	id             string
	peerConnection *webrtc.PeerConnection
	ctx            context.Context
	Disconnect     chan<- struct{}
}

var (
	resourceMap     map[string]*Resource
	resourceMapLock sync.Mutex
)

func addNewResource(resource *Resource) string {
	resourceMapLock.Lock()
	defer resourceMapLock.Unlock()

	resourceId := uuid.New().String()
	resource.id = resourceId

	resourceMap[resourceId] = resource
	resource.closeOnSignal()
	return resource.id
}

func removeResource(resourceId string) {
	resourceMapLock.Lock()
	defer resourceMapLock.Unlock()

	delete(resourceMap, resourceId)
}

// to be called on every newly created Resource
func (resource *Resource) closeOnSignal() {
	<-resource.ctx.Done()
	log.Println("Closing resource!")
	removeResource(resource.id)
}

// get Resource mapped to resourceId
func GetResource(resourceId string) *Resource {
	resourceMapLock.Lock()
	defer resourceMapLock.Unlock()

	return resourceMap[resourceId]
}
