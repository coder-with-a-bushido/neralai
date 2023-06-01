package hls

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"

	"coder-with-a-bushido.in/neralai/internal/whip"
)

func startFFmpeg(ctx context.Context, resourceId string, resource *whip.Resource) {
	args := []string{"-protocol_whitelist", "file,udp,rtp", "-i", fmt.Sprintf("%s/%s/connection.sdp", OutputDir, resourceId)}
	if resource.Video.Available {
		args = append(args, "-map", "0:v")
	}
	if resource.Audio.Available {
		args = append(args, "-map", "0:a")
	}
	if resource.Video.Available {
		args = append(args, "-c:v", "libx264", "-crf", "21", "-preset", "veryfast", "-r", "24")
	}
	if resource.Audio.Available {
		args = append(args, "-c:a", "aac", "-b:a", "128k", "-ac", "2")
	}
	args = append(args, "-f", "hls", "-hls_time", "4", "-hls_list_size", "10", "-hls_flags", "delete_segments+omit_endlist")
	args = append(args, fmt.Sprintf("%s/%s/hls/stream.m3u8", OutputDir, resourceId))

	ffmpeg := exec.Command("ffmpeg", args...)

	logFile, err := os.Create(
		fmt.Sprintf("%s/%s/ffmpeg_log.txt", OutputDir, resourceId),
	)
	if err != nil {
		panic(err)
	}
	defer logFile.Close()
	ffmpeg.Stderr = logFile

	if err := ffmpeg.Start(); err != nil {
		panic(err)
	}
	log.Printf("Starting HLS output for resource: %s", resourceId)

	go func() {
		<-ctx.Done()
		ffmpeg.Process.Kill()
	}()
}
