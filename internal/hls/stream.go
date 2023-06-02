package hls

import (
	"context"
	"errors"
	"fmt"

	"coder-with-a-bushido.in/neralai/internal/utils"
	"coder-with-a-bushido.in/neralai/internal/whip"
	"github.com/pion/rtp"
)

func NewStreamFromWHIPResource(ctx context.Context, resourceId string) {
	resource := whip.GetResource(resourceId)
	if resource == nil {
		panic(errors.New("Invalid Resource Id"))
	}

	if err := utils.CreateDir(fmt.Sprintf("%s/%s/hls", utils.GetOutputDir(), resourceId)); err != nil {
		panic(err)
	}

	rtpForward := startRTPForward()
	sdpContent := fmt.Sprintf("v=0\no=- 0 0 IN IP4 127.0.0.1\ns=Neralai\nc=IN IP4 127.0.0.1\nt=0 0\nm=audio %d RTP/AVP 111\na=rtpmap:111 OPUS/48000/2\nm=video %d RTP/AVP 96\na=rtpmap:96 VP8/90000",
		rtpForward.audio.port,
		rtpForward.video.port,
	)
	if err := utils.CreateAndWriteToFile(fmt.Sprintf("%s/%s/connection.sdp", utils.GetOutputDir(), resourceId), sdpContent); err != nil {
		panic(err)
	}
	go startFFmpeg(ctx, resourceId, resource)

	go func() {
		var audioPacket, videoPacket *rtp.Packet
		for {
			select {
			case audioPacket = <-resource.Audio.RTPPackets:
				rtpForward.writeAudio(audioPacket)
			case videoPacket = <-resource.Video.RTPPackets:
				rtpForward.writeVideo(videoPacket)
			case <-ctx.Done():
				rtpForward.closeUDPConns()
				utils.DeleteDir(fmt.Sprintf("%s/%s", utils.GetOutputDir(), resourceId))
				return
			}
		}
	}()
}
