package whip

import (
	"context"
	"log"
	"sync"

	"github.com/aidarkhanov/nanoid"
	"github.com/pion/rtp"
	"github.com/pion/webrtc/v3"
)

type Resource struct {
	id             string
	peerConnection *webrtc.PeerConnection
	ctx            context.Context

	Disconnect chan<- struct{}
	Audio      resourceMedia
	Video      resourceMedia
}

type resourceMedia struct {
	Available  bool
	RTPPackets <-chan *rtp.Packet
}

var (
	resourceMap     map[string]*Resource
	resourceMapLock sync.RWMutex
)

func GetResource(resourceId string) *Resource {
	resourceMapLock.RLock()
	defer resourceMapLock.RUnlock()

	resource, exists := resourceMap[resourceId]
	if exists {
		return resource
	}
	return nil
}

func AddNewResource(resource *Resource) string {
	resourceMapLock.Lock()
	defer resourceMapLock.Unlock()

	resourceId := nanoid.New()
	resource.id = resourceId
	log.Printf("New resource created: %s\n", resourceId)

	resourceMap[resourceId] = resource
	go resource.closeOnCtxCancel()
	return resource.id
}

func RemoveResource(resourceId string) {
	resourceMapLock.Lock()
	defer resourceMapLock.Unlock()

	delete(resourceMap, resourceId)
}

// Close WHIP resource and end connection on context cancel.
// Has to be called right after creating a new `Resource`
func (resource *Resource) closeOnCtxCancel() {
	<-resource.ctx.Done()
	log.Printf("Closing resource: %s", resource.id)
	if err := resource.peerConnection.Close(); err != nil {
		log.Println(err)
	}
}
