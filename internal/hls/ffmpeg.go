package hls

import (
	"context"
	"fmt"
	"log"
	"os/exec"

	"coder-with-a-bushido.in/neralai/internal/utils"
	"coder-with-a-bushido.in/neralai/internal/whip"
)

func startFFmpeg(ctx context.Context, resourceId string, resource *whip.Resource) {
	args := []string{"-protocol_whitelist", "file,udp,rtp", "-i", fmt.Sprintf("%s/%s/connection.sdp", utils.GetOutputDir(), resourceId)}
	if resource.Video.Available {
		args = append(args, "-map", "0:v")
	}
	if resource.Audio.Available {
		args = append(args, "-map", "0:a")
	}
	if resource.Video.Available {
		args = append(args, "-c:v", "libx264", "-crf", "23", "-preset", "veryfast", "-g", "60", "-sc_threshold", "0", "-b:v", "8000k", "-maxrate", "8000k", "-bufsize", "8000k")
	}
	if resource.Audio.Available {
		args = append(args, "-c:a", "aac", "-b:a", "128k", "-ac", "2")
	}
	args = append(args, "-f", "hls", "-hls_time", "4", "-hls_list_size", "10", "-hls_flags", "delete_segments+omit_endlist")
	args = append(args, fmt.Sprintf("%s/%s/hls/stream.m3u8", utils.GetOutputDir(), resourceId))

	ffmpeg := exec.Command("ffmpeg", args...)

	logFile, err := utils.NewLogFile(fmt.Sprintf("%s/%s/ffmpeg_log.txt", utils.GetOutputDir(), resourceId))
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
