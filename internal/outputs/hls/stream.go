package hls

import (
	"context"

	"coder-with-a-bushido.in/neralai/internal/whip"
)

func NewStream(ctx context.Context, resourceId string, resource *whip.Resource, outputDir string) MediaConnections {
	mediaConnections := NewMediaConnections()

	if err := mediaConnections.createOutputSDP(outputDir, resourceId); err != nil {
		panic(err)
	}
	go startFFmpeg(ctx, outputDir, resourceId, resource)

	go func() {
		<-ctx.Done()
		mediaConnections.closeUDPConns()
	}()

	return mediaConnections
}
