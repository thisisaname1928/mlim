package qbus

import (
	"encoding/binary"
	"fmt"
	"net"
)

type QBusClient struct {
	DefName string
	sock    net.Conn
}

func NewQBusClient(DefName string) (QBusClient, error) {
	var client QBusClient

	client.DefName = DefName

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

func (client *QBusClient) wait4ResCode() uint32 {
	packet, e := client.wait4Packet()
	if e != nil {
		return RES_CODE_NOT_OK
	}

	var rcode RCodePacket
	e = convertToStruct(packet.PacketData, &rcode)
	if e != nil {
		return RES_CODE_NOT_OK
	}

	return rcode.ResCode
}

func (client *QBusClient) SubChannel(name string) {
	client.Send(PACKET_SUB_ID, []byte(name))
}

func (client *QBusClient) Open() error {

	sock, e := net.Dial("unix", QBUS_PATH)
	if e != nil {
		return QBUS_OPEN_SOCKET_FAILED
	}

	client.sock = sock

	defer sock.Close()

	// do a handshake
	client.Send(PACKET_HANDSHAKE_ID, convertToBytes(HandShakePacket{Magic: Magic}))
	resCode := client.wait4ResCode()

	if resCode != RES_CODE_OK {
		fmt.Println("sus1")
		return QBUS_CLIENT_CANT_SIGN
	}
	// sign def name
	var defNamePacket DefNameSignPacket
	defNamePacket = append(defNamePacket, []byte(client.DefName)...)
	client.Send(PACKET_DEF_NAME_SIGN_ID, convertToBytes(defNamePacket))
	resCode = client.wait4ResCode()

	if resCode != RES_CODE_OK {
		fmt.Println("sus")
		return QBUS_CLIENT_CANT_SIGN
	}

	// test
	client.SubChannel("com.qbusclient.test")

	if resCode != RES_CODE_OK {
		fmt.Println("sus")
		return QBUS_CLIENT_CANT_SIGN
	}

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
