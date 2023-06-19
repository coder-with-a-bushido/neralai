# Debug

## HLS output

If there's a problem with the HLS output, then this is how you should debug.

1. First comment out all the ffmpeg related lines from `internal/hls/stream.go`.
2. Copy the stream/resource id from console output.
3. If you want to play the stream as a video, run:
   `sh -x ./ffplay.sh <STREAM_ID>`

   Or, if you want to create HLS files with ffmpeg and check logs, run:
   `sh -x ./ffmpeg_hls.sh <STREAM_ID>`

4. You can check the log file for the ffmpeg process at `output/<STREAM_ID>/hls/ffmpeg_log.txt`
