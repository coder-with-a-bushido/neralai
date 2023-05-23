package whip

import (
	"context"
	"log"
	"sync"

	"github.com/google/uuid"
	"github.com/pion/rtp"
	"github.com/pion/webrtc/v3"
)

type Resource struct {
	id             string
	peerConnection *webrtc.PeerConnection
	ctx            context.Context

	Disconnect   chan<- struct{}
	AudioPackets <-chan *rtp.Packet
	VideoPackets <-chan *rtp.Packet
}

var (
	resourceMap     map[string]*Resource
	resourceMapLock sync.RWMutex
)

func addNewResource(resource *Resource) string {
	resourceMapLock.Lock()
	defer resourceMapLock.Unlock()

	resourceId := uuid.New().String()
	resource.id = resourceId

	resourceMap[resourceId] = resource
	go resource.closeOnSignal()
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
	if err := resource.peerConnection.Close(); err != nil {
		log.Println(err)
	}
	removeResource(resource.id)
}

// get Resource mapped to resourceId
func GetResource(resourceId string) *Resource {
	resourceMapLock.RLock()
	defer resourceMapLock.RUnlock()

	return resourceMap[resourceId]
}
