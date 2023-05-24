package hls

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
)

func NewStream(ctx context.Context, resourceId string, outputDir string) MediaConnections {
	var mediaConnections MediaConnections
	mediaConnections.audio = createUDPConn()
	mediaConnections.video = createUDPConn()

	if err := createOutputSDP(mediaConnections, resourceId, outputDir); err != nil {
		panic(err)
	}

	go func() {
		<-ctx.Done()
		closeUDPConn(mediaConnections.audio.conn)
		closeUDPConn(mediaConnections.video.conn)
	}()

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

func closeUDPConn(conn net.PacketConn) {
	if closeErr := conn.Close(); closeErr != nil {
		panic(closeErr)
	}
}

func createOutputSDP(mediaConnections MediaConnections, resourceId string, outputDir string) error {
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
