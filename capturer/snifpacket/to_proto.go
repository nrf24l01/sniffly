package snifpacket

import (
	"encoding/json"

	pb "github.com/nrf24l01/sniffly/capture_receiver/proto"
)

func (sp *SnifPacket) ToProto() (*pb.Packet, error) {
	payload, err := json.Marshal(sp)
	if err != nil {
		return nil, err
	}
	
	return &pb.Packet{
		Timestamp:    sp.Timestamp,
		Payload:      payload,
		SourceId:     "capturer-go",
	}, nil
}