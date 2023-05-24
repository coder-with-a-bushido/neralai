package hls

import (
	"errors"
	"net"

	"github.com/pion/rtp"
)

type udpMediaConn struct {
	conn *net.UDPConn
	port int
}
type MediaConnections struct {
	audio *udpMediaConn
	video *udpMediaConn
}

func (m *MediaConnections) WriteAudio(audioPacket *rtp.Packet) {
	writeMedia(m.audio.conn, audioPacket)
}

func (m *MediaConnections) WriteVideo(videoPacket *rtp.Packet) {
	writeMedia(m.video.conn, videoPacket)
}

func writeMedia(conn *net.UDPConn, rtpPacket *rtp.Packet) {
	b := make([]byte, 1500)
	n, err := rtpPacket.MarshalTo(b)
	if err != nil {
		panic(err)
	}

	if _, writeErr := conn.Write(b[:n]); writeErr != nil {
		var opError *net.OpError
		if errors.As(writeErr, &opError) && opError.Err.Error() != "write: connection refused" {
			panic(writeErr)
		}
	}
}
