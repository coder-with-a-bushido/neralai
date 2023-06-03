ffmpeg -protocol_whitelist file,udp,rtp\
    -i output/$1/connection.sdp -map 0:v -map 0:a\
    -c:v libx264 -crf 23 -preset veryfast -g 60 -sc_threshold 0 -b:v 8000k -maxrate 8000k -bufsize 8000k\
    -c:a aac -b:a 128k -ac 2\
    -f hls\
    -hls_time 4 -hls_list_size 10 -hls_flags delete_segments+omit_endlist\
    output/$1/hls/stream.m3u8 2> output/$1/hls/ffmpeg_log.txt
