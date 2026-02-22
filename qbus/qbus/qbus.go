package qbus

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"net"
	"os"
	"strconv"
	"syscall"
)

type client struct {
	Socket        net.Conn
	DefName       string
	ExePath       string
	ClientVersion uint32
	SubscribeList []uint32 // pid that subcribe to this client
	SubcribeTo    []string
}

type QBus struct {
	Socket     net.Listener
	ClientList map[uint32]client   // use pid to access them
	DefNameMap map[string][]uint32 // use defname to search for pid
}

var QBUS_PATH = "../rootfs/rootfs/tmp/qbus"

var QBUS_ALREADY_STARTED = errors.New("QBUS_ALREADY_STARTED")
var QBUS_CREATE_SOCKET_FAILED = errors.New("QBUS_CREATE_SOCKET_FAILED")
var QBUS_OPEN_SOCKET_FAILED = errors.New("QBUS_OPEN_SOCKET_FAILED")
var QBUS_BAD_PACKET_FORMAT = errors.New("QBUS_BAD_PACKET_FORMAT")
var QBUS_CLIENT_CANT_SIGN = errors.New("QBUS_CLIENT_CANT_SIGN")

func getPeerPID(conn net.Conn) (uint32, error) {
	unixConn, ok := conn.(*net.UnixConn)
	if !ok {
		return 0, fmt.Errorf("not a unix connection")
	}

	rawConn, err := unixConn.File()
	if err != nil {
		return 0, err
	}
	defer rawConn.Close()

	cred, err := syscall.GetsockoptUcred(int(rawConn.Fd()), syscall.SOL_SOCKET, syscall.SO_PEERCRED)
	if err != nil {
		return 0, err
	}

	return uint32(cred.Pid), nil
}

func convertToBytes(inp any) []byte {
	var res = new(bytes.Buffer)

	binary.Write(res, binary.LittleEndian, inp)

	return res.Bytes()
}

func convertToStruct(inp []byte, out any) error {
	reader := bytes.NewReader(inp)

	return binary.Read(reader, binary.LittleEndian, out)
}

func getProcessPath(pid int) (string, error) {
	procPath := "/proc/" + strconv.Itoa(pid) + "/exe"

	exePath, err := os.Readlink(procPath)
	if err != nil {
		return "", err
	}

	return exePath, nil
}

func QBusInit() (QBus, error) {
	var bus QBus

	bus.ClientList = make(map[uint32]client, 50)   // pre alloc for faster access
	bus.DefNameMap = make(map[string][]uint32, 50) // pre alloc for faster access

	// check if qbus available
	if _, err := os.Stat(QBUS_PATH); err == nil {
		e := os.Remove(QBUS_PATH)

		if e != nil {
			return bus, QBUS_ALREADY_STARTED
		}
	}

	sock, e := net.Listen("unix", QBUS_PATH)
	if e != nil {
		fmt.Println(e)
		return bus, QBUS_OPEN_SOCKET_FAILED
	}

	bus.Socket = sock

	return bus, nil
}

func (bus *QBus) Open() {
	for {
		conn, e := bus.Socket.Accept()
		if e != nil {
			fmt.Println(e)
			continue
		}

		go bus.handlePacket(conn)
	}
}

const PACKET_LENGTH_SZ = 4
const PACKET_TYPE_SZ = 4

type PacketLength uint32
type PacketType uint32

// func (bus *QBus) handleConnection(conn net.Conn) {
// 	defer conn.Close()

// 	var packetLengthBuf = make([]byte, PACKET_LENGTH_SZ)
// 	var packetTypeBuf = make([]byte, PACKET_TYPE_SZ)
// 	var PacketLength uint32
// 	var packetType uint32

// 	// set first 4 byte as packet length (uint32)
// 	for {

// 		// read packet length
// 		n, _ := conn.Read(packetLengthBuf)

// 		if n != PACKET_LENGTH_SZ {
// 			continue
// 		}

// 		PacketLength = binary.LittleEndian.Uint32(packetLengthBuf)

// 		// read packet type
// 		n, _ = conn.Read(packetTypeBuf)
// 		if n != PACKET_TYPE_SZ {
// 			continue
// 		}

// 		packetType = binary.LittleEndian.Uint32(packetTypeBuf)

// 		// read buffer
// 		var buffer = make([]byte, PacketLength)
// 		n, _ = conn.Read(buffer)

// 		// ??
// 		if n != int(PacketLength) {
// 			continue
// 		}
// 		bus.handlePacket(conn, PacketType(packetType), buffer)
// 	}
// }

// packet def

const Magic uint32 = 1111

const (
	NOTHING = iota
	PACKET_HANDSHAKE_ID
	PACKET_RCODE_ID
	PACKET_DEF_NAME_SIGN_ID
	PACKET_SUB_ID
)

type HandShakePacket struct {
	Magic uint32
}

// for return code
type RCodePacket struct {
	ResCode uint32
}

type DefNameSignPacket []byte
type SubPacket []byte

const (
	RES_CODE_NOTHING = iota
	RES_CODE_OK
	RES_CODE_NOT_OK
	RES_CODE_DEF_NAME_ALREADY_USED
	RES_CODE_PERMISSION_DENIED
	RES_CODE_NO_CHANNEL
	RES_CODE_DEF_NAME_BLANK
)

func sendRCode(conn net.Conn, code uint32) {
	var rcode = RCodePacket{ResCode: code}

	Send(conn, PACKET_RCODE_ID, rcode)
}

func Send(sock net.Conn, packetType uint32, packet any) {
	buf := convertToBytes(packet)
	var length PacketLength = PacketLength(len(buf))

	var sendDat []byte
	sendDat = binary.LittleEndian.AppendUint32(sendDat, uint32(length))
	sendDat = binary.LittleEndian.AppendUint32(sendDat, packetType)
	sendDat = append(sendDat, buf...)

	sock.Write(sendDat)
}

func (bus *QBus) handlePacket(conn net.Conn) {

	defer conn.Close()

	var packetLengthBuf = make([]byte, PACKET_LENGTH_SZ)
	var packetTypeBuf = make([]byte, PACKET_TYPE_SZ)
	var PacketLength uint32
	var packetType uint32
	var curDefName string = ""
	var pid, _ = getPeerPID(conn)

	// set first 4 byte as packet length (uint32)
	for {

		// read packet length
		n, _ := conn.Read(packetLengthBuf)

		if n != PACKET_LENGTH_SZ {
			continue
		}

		PacketLength = binary.LittleEndian.Uint32(packetLengthBuf)

		// read packet type
		n, _ = conn.Read(packetTypeBuf)
		if n != PACKET_TYPE_SZ {
			continue
		}

		packetType = binary.LittleEndian.Uint32(packetTypeBuf)

		// read buffer
		var packet = make([]byte, PacketLength)
		n, _ = conn.Read(packet)

		// ??
		if n != int(PacketLength) {
			continue
		}

		switch packetType {
		case PACKET_HANDSHAKE_ID:
			// read
			var curPacket HandShakePacket
			e := convertToStruct(packet, &curPacket)

			if e != nil {
				sendRCode(conn, RES_CODE_NOT_OK)
				break
			}

			if curPacket.Magic != Magic {
				sendRCode(conn, RES_CODE_NOT_OK)
				break
			}

			// sign
			pid, e := getPeerPID(conn)
			bus.ClientList[uint32(pid)] = client{Socket: conn}

			sendRCode(conn, RES_CODE_OK)
		case PACKET_DEF_NAME_SIGN_ID:
			fmt.Println(string(packet), "sign!")
			curDefName = string(packet)
			_, ok := bus.DefNameMap[curDefName]

			if !ok {
				bus.DefNameMap[curDefName] = make([]uint32, 0)
			}
			bus.DefNameMap[curDefName] = append(bus.DefNameMap[curDefName], pid)
			sendRCode(conn, RES_CODE_OK)
		case PACKET_SUB_ID:
			if curDefName == "" {
				sendRCode(conn, RES_CODE_DEF_NAME_BLANK)
				return
			}

			im, ok := bus.DefNameMap[string(packet)]

			// prevent it sub to it self
			// if v == pid {
			// 	sendRCode(conn, RES_CODE_OK)
			// 	break
			// }

			if !ok {
				sendRCode(conn, RES_CODE_NO_CHANNEL)
				break
			}

			tmp := bus.ClientList[pid]
			tmp.SubcribeTo = append(tmp.SubcribeTo, string(packet))
			bus.ClientList[pid] = tmp

			for _, v := range im {
				c := bus.ClientList[v]
				c.SubscribeList = append(c.SubscribeList, pid)
				bus.ClientList[v] = c
				sendRCode(conn, RES_CODE_OK)
			}

			fmt.Println(bus.ClientList[bus.DefNameMap["com.qbusclient.test"][0]].SubscribeList)
		default:

		}

	}
}
