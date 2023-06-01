package hls

import (
	"context"
	"fmt"
	"os"

	"coder-with-a-bushido.in/neralai/internal/whip"
	"github.com/pion/rtp"
)

func StreamFromWHIPResource(ctx context.Context, resourceId string) {
	resource := whip.GetResource(resourceId)

	createOutputResourceDir(resourceId)

	mediaConnections := NewStream()
	if err := mediaConnections.createOutputSDP(resourceId); err != nil {
		panic(err)
	}
	go startFFmpeg(ctx, resourceId, resource)

	go func() {
		<-ctx.Done()

	}()

	go func() {
		var audioPacket, videoPacket *rtp.Packet
		for {
			select {
			case audioPacket = <-resource.Audio.RTPPackets:
				mediaConnections.WriteAudio(audioPacket)
			case videoPacket = <-resource.Video.RTPPackets:
				mediaConnections.WriteVideo(videoPacket)
			case <-ctx.Done():
				mediaConnections.closeUDPConns()
				cleanupOutputResourceDir(resourceId)
				return
			}
		}
	}()
}

func createOutputResourceDir(resourceId string) {
	if err := os.MkdirAll(fmt.Sprintf("%s/%s/hls", OutputDir, resourceId), os.ModePerm); err != nil {
		panic(err)
	}
}

func cleanupOutputResourceDir(resourceId string) {
	if err := os.RemoveAll(
		fmt.Sprintf("%s/%s", OutputDir, resourceId),
	); err != nil {
		panic(err)
	}
}
