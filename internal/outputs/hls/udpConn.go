package hls

import (
	"errors"
	"fmt"
	"log"
	"net"
	"os"

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

func NewMediaConnections() MediaConnections {
	var mediaConnections MediaConnections
	mediaConnections.audio = createUDPConn()
	mediaConnections.video = createUDPConn()
	return mediaConnections
}

func createUDPConn() *udpMediaConn {
	var udpMediaConn udpMediaConn
	freePort, err := getFreePortEven()
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
	log.Printf("Making UDP conn from %s to %s", udpMediaConn.conn.LocalAddr().String(), udpMediaConn.conn.RemoteAddr().String())

	return &udpMediaConn
}

func (mediaConnections *MediaConnections) closeUDPConns() {
	for _, udpMediaConn := range []*udpMediaConn{mediaConnections.audio, mediaConnections.video} {
		if closeErr := udpMediaConn.conn.Close(); closeErr != nil {
			panic(closeErr)
		}
	}
}

func (mediaConnections *MediaConnections) WriteAudio(audioPacket *rtp.Packet) {
	writeMedia(mediaConnections.audio.conn, audioPacket)
}

func (mediaConnections *MediaConnections) WriteVideo(videoPacket *rtp.Packet) {
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

func (mediaConnections *MediaConnections) createOutputSDP(outputDir, resourceId string) error {
	file, err := os.Create(
		fmt.Sprintf("%s/%s/connection.sdp", outputDir, resourceId),
	)
	if err != nil {
		return err
	}
	defer file.Close()

	sdpContent := fmt.Sprintf("v=0\no=- 0 0 IN IP4 127.0.0.1\ns=Neralai\nc=IN IP4 127.0.0.1\nt=0 0\nm=audio %d RTP/AVP 111\na=rtpmap:111 OPUS/48000/2\nm=video %d RTP/AVP 96\na=rtpmap:96 VP8/90000",
		mediaConnections.audio.port,
		mediaConnections.video.port,
	)
	if _, err = file.WriteString(sdpContent); err != nil {
		return err
	}

	if err = file.Sync(); err != nil {
		return err
	}
	return nil
}
