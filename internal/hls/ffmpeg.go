package hls

import (
	"context"
	"io"
	"os"
	"os/exec"
	"syscall"
	"time"

	"github.com/coder-with-a-bushido/neralai/internal/utils"
	"github.com/coder-with-a-bushido/neralai/internal/whip"
)

type ffmpeg struct {
	cmd     *exec.Cmd
	logFile *os.File
	stdin   io.WriteCloser

	sdpFilePath         string
	ffmpegLogFilePath   string
	hlsPlaylistFilePath string
}

// Run `ffmpeg“ command that gets input from a SDP file and creates HLS output.
func (ffmpeg *ffmpeg) startProcess(ctx context.Context, resource *whip.Resource) error {
	// Construct the ffmpeg command arguments
	args := []string{"-protocol_whitelist", "file,udp,rtp", "-i", ffmpeg.sdpFilePath}
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
	args = append(args, ffmpeg.hlsPlaylistFilePath)

	ffmpegCmd := exec.Command("ffmpeg", args...)

	// Log file for the ffmpeg command
	logFile, err := utils.OpenFile(ffmpeg.ffmpegLogFilePath)
	if err != nil {
		return err
	}
	ffmpegCmd.Stderr = logFile

	// Stdin for input
	stdin, err := ffmpegCmd.StdinPipe()
	if err != nil {
		return err
	}

	// Create a process group for the command
	// Useful for killing the process and its subprocesses with a single pid
	ffmpegCmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}

	// Start the command
	if err := ffmpegCmd.Start(); err != nil {
		return err
	}
	ffmpeg.cmd = ffmpegCmd
	ffmpeg.logFile = logFile
	ffmpeg.stdin = stdin
	return nil
}

func (ffmpeg *ffmpeg) endProcess() {
	// Perform Ctrl+C action
	ffmpeg.stdin.Write([]byte("q"))
	// Wait for the process to terminate gracefully
	done := make(chan error)
	go func() { done <- ffmpeg.cmd.Wait() }()
	select {
	case <-done:
	case <-time.After(10 * time.Second):
		// Kill ffmpeg process and all its subprocesses on a timeout of 10s
		// if they have not yet shut down
		ffmpeg.stdin.Close()
		syscall.Kill(-ffmpeg.cmd.Process.Pid, syscall.SIGKILL)
	}
	ffmpeg.logFile.Close()
}
