package hls

import (
	"net"

	"github.com/coder-with-a-bushido/neralai/internal/utils"
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

// Prepare UDP connections for forwarding RTP packets to ffmpeg.
func startRTPForward() rtpForward {
	var rtpForward rtpForward
	var err error
	if rtpForward.audio.port, err = utils.GetFreePortEven(); err != nil {
		panic(err)
	} else if rtpForward.audio.conn, err = utils.NewLocalUDPConn(rtpForward.audio.port); err != nil {
		panic(err)
	}

	if rtpForward.video.port, err = utils.GetFreePortEven(); err != nil {
		panic(err)
	} else if rtpForward.video.conn, err = utils.NewLocalUDPConn(rtpForward.video.port); err != nil {
		panic(err)
	}
	return rtpForward
}

// Close the UDP connections used for forwarding RTP packets.
func (rtpForward *rtpForward) endRTPForward() {
	if err := utils.CloseLocalUDPConn(rtpForward.audio.conn); err != nil {
		panic(err)
	} else if err := utils.CloseLocalUDPConn(rtpForward.video.conn); err != nil {
		panic(err)
	}
}

// Forward audio RTP packet.
func (rtpForward *rtpForward) writeAudio(audioPacket *rtp.Packet) {
	if err := utils.WriteRTPPacketToUDPConn(rtpForward.audio.conn, audioPacket); err != nil {
		panic(err)
	}
}

// Forward video RTP packet.
func (rtpForward *rtpForward) writeVideo(videoPacket *rtp.Packet) {
	if err := utils.WriteRTPPacketToUDPConn(rtpForward.video.conn, videoPacket); err != nil {
		panic(err)
	}
}
