package hls

import (
	"context"
	"errors"
	"fmt"
	"log"

	"github.com/coder-with-a-bushido/neralai/internal/utils"
	"github.com/coder-with-a-bushido/neralai/internal/whip"
	"github.com/pion/rtp"
)

// Start HLS with input from WHIP resource.
func NewStreamFromWHIPResource(ctx context.Context, resourceId string) {
	resource := whip.GetResource(resourceId)
	if resource == nil {
		panic(errors.New("Invalid Resource Id"))
	}

	// Output file paths
	resourceDirPath := fmt.Sprintf("%s/%s", utils.GetOutputDir(), resourceId)
	sdpFilePath := resourceDirPath + "/connection.sdp"
	ffmpegLogFilePath := resourceDirPath + "/ffmpeg_log.txt"
	hlsDir := resourceDirPath + "/hls"
	hlsPlaylistFilePath := hlsDir + "/stream.m3u8"

	// Create output directory for that resource.
	if err := utils.CreateDir(hlsDir); err != nil {
		panic(err)
	}

	// Prepare RTP packet forward via UDP and create SDP file for it
	rtpForward := startRTPForward()
	sdpContent := fmt.Sprintf("v=0\no=- 0 0 IN IP4 127.0.0.1\ns=Neralai\nc=IN IP4 127.0.0.1\nt=0 0\nm=audio %d RTP/AVP 111\na=rtpmap:111 OPUS/48000/2\nm=video %d RTP/AVP 96\na=rtpmap:96 VP8/90000",
		rtpForward.audio.port,
		rtpForward.video.port,
	)
	if err := utils.WriteToFile(sdpFilePath, sdpContent); err != nil {
		panic(err)
	}

	ffmpeg := &ffmpeg{
		sdpFilePath:         sdpFilePath,
		ffmpegLogFilePath:   ffmpegLogFilePath,
		hlsPlaylistFilePath: hlsPlaylistFilePath,
	}
	// Start ffmpeg processing for the WHIP resource
	err := ffmpeg.startProcess(ctx, resource)
	if err != nil {
		panic(err)
	}
	log.Printf("Starting HLS output for resource: %s", resourceId)

	// Write audio, video RTP packets from WHIP to `rtpForward``
	go func() {
		var audioPacket, videoPacket *rtp.Packet
		for {
			select {
			case audioPacket = <-resource.Audio.RTPPackets:
				rtpForward.writeAudio(audioPacket)
			case videoPacket = <-resource.Video.RTPPackets:
				rtpForward.writeVideo(videoPacket)
			case <-ctx.Done():
				rtpForward.endRTPForward()
				ffmpeg.endProcess()
				return
			}
		}
	}()
}
