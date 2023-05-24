package whip

import (
	"context"
	"io"
	"log"

	"github.com/pion/rtp"
	"github.com/pion/webrtc/v3"
)

func NewConnection(ctx context.Context, offerSDP string, disconnect chan<- struct{}) (answerSDP string, resourceId string, err error) {
	config := webrtc.Configuration{
		ICEServers: []webrtc.ICEServer{
			{
				// URLs: []string{
				// 	"stun:turn.eyevinn.technology:3478",
				// 	"turn:turn.eyevinn.technology:3478",
				// },
			},
		},
	}
	peerConnection, err := api.NewPeerConnection(config)
	if err != nil {
		return "", "", nil
	}

	// make answer SDP use "recvonly" attribute
	if _, err = peerConnection.AddTransceiverFromKind(webrtc.RTPCodecTypeVideo, webrtc.RTPTransceiverInit{
		Direction: webrtc.RTPTransceiverDirectionRecvonly,
	}); err != nil {
		return "", "", nil
	}
	if _, err = peerConnection.AddTransceiverFromKind(webrtc.RTPCodecTypeAudio, webrtc.RTPTransceiverInit{
		Direction: webrtc.RTPTransceiverDirectionRecvonly,
	}); err != nil {
		return "", "", nil
	}

	// channels to pass the audio and video RTP packets
	audioPackets := make(chan *rtp.Packet)
	videoPackets := make(chan *rtp.Packet)

	peerConnection.OnTrack(
		func(
			remoteTrack *webrtc.TrackRemote,
			rtpReceiver *webrtc.RTPReceiver,
		) {
			// Read RTP packets of the track
			for {
				rtpPacket, _, readErr := remoteTrack.ReadRTP()
				if readErr != nil {
					if readErr == io.EOF {
						return
					}
					panic(readErr)
				}
				switch remoteTrack.Kind() {
				// If it's audio, send to audioPackets
				case webrtc.RTPCodecTypeAudio:
					audioPackets <- rtpPacket
				// If it's video, send to videoPackets
				case webrtc.RTPCodecTypeVideo:
					videoPackets <- rtpPacket
				}
			}
		})

	peerConnection.OnICEConnectionStateChange(func(i webrtc.ICEConnectionState) {
		// close connection on `ICEConnectionStateFailed`
		if i == webrtc.ICEConnectionStateFailed {
			log.Println("ICE connection state -> Failed")
			disconnect <- struct{}{}
		}
	})

	// set remote description from `offerSDP`
	if err := peerConnection.SetRemoteDescription(
		webrtc.SessionDescription{
			Type: webrtc.SDPTypeOffer,
			SDP:  string(offerSDP),
		},
	); err != nil {
		return "", "", nil
	}

	// Gather all ICE candidates beforehand
	// since we don't support Trickle ICE
	iceGatherComplete := webrtc.GatheringCompletePromise(peerConnection)

	// answer
	answer, err := peerConnection.CreateAnswer(nil)
	if err != nil {
		return "", "", err
	}
	if err = peerConnection.SetLocalDescription(answer); err != nil {
		return "", "", err
	}

	// when ICE gathering is complete
	<-iceGatherComplete
	// create a new `Resource` for this connection
	resourceId = addNewResource(
		&Resource{
			peerConnection: peerConnection,
			ctx:            ctx,
			Disconnect:     disconnect,
			AudioPackets:   audioPackets,
			VideoPackets:   videoPackets,
		},
	)
	return answer.SDP, resourceId, nil
}
