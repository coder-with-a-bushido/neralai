ffmpeg -protocol_whitelist file,udp,rtp\
    -i output/$1/connection.sdp -map 0:v -map 0:a\
    -c:v libx264 -crf 21 -preset veryfast -r 24\
    -c:a aac -b:a 128k -ac 2\
    -f hls\
    -hls_time 4 -hls_list_size 10 -hls_flags delete_segments+omit_endlist\
    output/$1/stream.m3u8 2> ffmpeg_log.txt
