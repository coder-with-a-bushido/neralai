package outputs

import (
	"context"
	"fmt"
	"os"

	"coder-with-a-bushido.in/neralai/internal/outputs/hls"
	"coder-with-a-bushido.in/neralai/internal/whip"
	"github.com/pion/rtp"
)

type Options struct {
	Recording bool
	HLSStream bool
}

// type Output struct {
// 	id      string // resource ID
// 	options OutputOptions
// 	cxt     context.Context
// }

func StartFromWHIPResource(ctx context.Context, resourceId string, outputOptions Options) {
	resource := whip.GetResource(resourceId)

	createOutputResourceDir(resourceId)

	var mediaConnections hls.MediaConnections
	if outputOptions.HLSStream == true {
		mediaConnections = hls.NewStream(ctx, resourceId, outputDir)
	}

	go func() {
		var audioPacket, videoPacket *rtp.Packet
		for {
			select {
			case audioPacket = <-resource.AudioPackets:
				if outputOptions.HLSStream == true {
					mediaConnections.WriteAudio(audioPacket)
				}
			case videoPacket = <-resource.VideoPackets:
				if outputOptions.HLSStream == true {
					mediaConnections.WriteVideo(videoPacket)
				}
			case <-ctx.Done():
				cleanupOutputResourceDir(resourceId)
				break
			}
		}
	}()
}

func createOutputResourceDir(resourceId string) {
	if err := os.Mkdir(fmt.Sprintf("%s/%s", outputDir, resourceId), os.ModePerm); err != nil {
		panic(err)
	}
}

func cleanupOutputResourceDir(resourceId string) {
	if err := os.RemoveAll(
		fmt.Sprintf("%s/%s", outputDir, resourceId),
	); err != nil {
		panic(err)
	}
}
