package hls

import (
	"errors"
	"fmt"
	"net"

	"coder-with-a-bushido.in/neralai/internal/utils"
	"github.com/pion/rtp"
)

type udpMediaConn struct {
	conn *net.UDPConn
	port int
}
type rtpForward struct {
	audio udpMediaConn
	video udpMediaConn
}

func startRTPForward() rtpForward {
	var mediaConnections rtpForward
	mediaConnections.audio = createUDPConn()
	mediaConnections.video = createUDPConn()
	return mediaConnections
}

func createUDPConn() udpMediaConn {
	var udpMediaConn udpMediaConn
	freePort, err := utils.GetFreePortEven()
	if err != nil {
		panic(err)
	}

	// Create remote addr with random port
	var raddr *net.UDPAddr
	if raddr, err = net.ResolveUDPAddr("udp", fmt.Sprintf("127.0.0.1:%d", freePort)); err != nil {
		panic(err)
	}

	// Set the port for UDP conn
	udpMediaConn.port = raddr.Port

	// Dial udp
	if udpMediaConn.conn, err = net.DialUDP("udp", nil, raddr); err != nil {
		panic(err)
	}

	return udpMediaConn
}

func (mediaConnections *rtpForward) closeUDPConns() {
	for _, udpMediaConn := range []udpMediaConn{mediaConnections.audio, mediaConnections.video} {
		if closeErr := udpMediaConn.conn.Close(); closeErr != nil {
			panic(closeErr)
		}
	}
}

func (mediaConnections *rtpForward) writeAudio(audioPacket *rtp.Packet) {
	writeMedia(mediaConnections.audio.conn, audioPacket)
}

func (mediaConnections *rtpForward) writeVideo(videoPacket *rtp.Packet) {
	writeMedia(mediaConnections.video.conn, videoPacket)
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
