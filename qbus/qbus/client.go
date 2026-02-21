package qbus

import (
	"encoding/binary"
	"fmt"
	"net"
)

type QBusClient struct {
	sock net.Conn
}

func NewQBusClient() (QBusClient, error) {
	var client QBusClient

	return client, nil
}

func (client *QBusClient) Send(packetType uint32, buf []byte) {
	var length PacketLength = PacketLength(len(buf))

	var sendDat []byte
	sendDat = binary.LittleEndian.AppendUint32(sendDat, uint32(length))
	sendDat = binary.LittleEndian.AppendUint32(sendDat, packetType)
	sendDat = append(sendDat, buf...)

	client.sock.Write(sendDat)
}

type packet struct {
	Length     PacketLength
	Type       PacketType
	PacketData []byte
}

func (client *QBusClient) wait4Packet() (packet, error) {
	var packetLengthBuf = make([]byte, PACKET_LENGTH_SZ)
	var packetTypeBuf = make([]byte, PACKET_TYPE_SZ)
	var packetLength uint32
	var packetType uint32

	// set first 4 byte as packet length (uint32)

	// read packet length
	n, _ := client.sock.Read(packetLengthBuf)

	if n != PACKET_LENGTH_SZ {
		return packet{}, QBUS_BAD_PACKET_FORMAT
	}

	packetLength = binary.LittleEndian.Uint32(packetLengthBuf)

	// read packet type
	n, _ = client.sock.Read(packetTypeBuf)
	if n != PACKET_TYPE_SZ {
		return packet{}, QBUS_BAD_PACKET_FORMAT
	}

	packetType = binary.LittleEndian.Uint32(packetTypeBuf)

	// read buffer
	var buffer = make([]byte, packetLength)
	n, _ = client.sock.Read(buffer)

	// ??
	if n != int(packetLength) {
		return packet{}, QBUS_BAD_PACKET_FORMAT
	}

	return packet{PacketLength(packetLength), PacketType(packetType), buffer}, nil
}

func (client *QBusClient) Open() error {

	sock, e := net.Dial("unix", QBUS_PATH)
	if e != nil {
		return QBUS_OPEN_SOCKET_FAILED
	}

	client.sock = sock

	defer sock.Close()

	client.Send(PACKET_HANSHAKE_ID, convertToBytes(HandShakePacket{Magic: Magic}))

	for {
		packet, e := client.wait4Packet()

		if e != nil {
			fmt.Println("oh sus")
			break
		}

		switch packet.Type {
		case PACKET_RCODE_ID:
			var rcode RCodePacket
			e := convertToStruct(packet.PacketData, &rcode)

			fmt.Println(e)

			fmt.Println("get rcode=", rcode.ResCode)
		}
	}

	return nil
}
