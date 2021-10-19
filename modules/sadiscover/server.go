package sadiscover

import (
	"bytes"
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/md5"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/zhiting-tech/smartassistant/modules/config"
	"github.com/zhiting-tech/smartassistant/pkg/logger"
	"github.com/zhiting-tech/smartassistant/pkg/rand"
	"net"
)

var saID uint32
var key []byte
var helloPacket = []byte{0x21, 0x31, 0x00, 0x20, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff}

func initSaID() {
	// 转换saID字符串为数值类型
	saIDBytes := []byte(config.GetConf().SmartAssistant.ID)
	data := uint32(0)
	for _, b := range saIDBytes {
		data = (data << 8) | uint32(b)
	}
	saID = data
}

type Server struct {
}

func NewSaDiscoverServer() *Server {
	return &Server{}
}

func (s *Server) Run(ctx context.Context) {
	initSaID()

	// 随机生成一个token
	m := md5.New()
	m.Write([]byte(rand.String(32)))
	token, _ := hex.DecodeString(hex.EncodeToString(m.Sum(nil)))

	go s.readFromUDP(token)

	<-ctx.Done()
	logger.Warning("sa discover server stopped")
}

func (s *Server) readFromUDP(token []byte) {
	addr, err := net.ResolveUDPAddr("udp", ":54321")
	if err != nil {
		logger.Error("[sa discover] Can't resolve address: ", err)
		return
	}
	conn, err := net.ListenUDP("udp", addr)
	if err != nil {
		logger.Error("[sa discover] Error listening:", err)
		return
	}
	defer conn.Close()

	data := make([]byte, 1024)
	logger.Info("starting sa discover server")
	for {
		n, remoteAddr, err := conn.ReadFromUDP(data)
		if err != nil {
			logger.Error("[sa discover] failed to read UDP msg because of ", err.Error())
			continue
		}
		logger.Printf("[sa discover] read %v from %v", n, remoteAddr)

		// 响应hello应答包
		if bytes.Compare(helloPacket, data[:n]) == 0 {
			logger.Printf("[sa discover] get hello packet from %v", remoteAddr)
			_, err := conn.WriteTo(helloResponse(), remoteAddr)
			if err != nil {
				logger.Printf("[sa discover] write response error %v", err)
			} else {
				logger.Printf("[sa discover] hello response to %v", remoteAddr)
			}
			continue
		}

		// 解密data
		packet, err := decode(data)
		if err != nil {
			logger.Warn("[sa discover] decode error", err)
			continue
		}

		// 加密token，返回token
		if binary.BigEndian.Uint16(packet.Reserve) == reserveTokenKey {
			key = packet.MD5Sum
			logger.Infof("[sa discover] get key message %v from %v", string(packet.MD5Sum), remoteAddr)
			msg := encode(token, key, saID)
			_, err := conn.WriteTo(msg, remoteAddr)
			if err != nil {
				logger.Warn("[sa discover] write to error", err)
			}
			continue
		}

		// 返回sa信息
		logger.Printf("[sa discover] result: %v", packet.Data)
		result, err := decrypt(packet.Data, token)
		if err != nil {
			logger.Warn("[sa discover] decrypt error", err)
		}
		logger.Infof("[sa discover] get result bytes %v from %v", result, remoteAddr)
		logger.Infof("[sa discover] get result %s from %v", string(result), remoteAddr)

		msg := make(map[string]interface{})
		if err = json.Unmarshal(result, &msg); err != nil {
			continue
		}
		method, ok := msg["method"]
		if !ok {
			continue
		}
		if method == "get_prop.info" {
			re := Result{
				ID: int(msg["id"].(float64)),
				Re: Info{
					Model: "smart_assistant",
					SwVer: "0.0.1",
					HwVer: "1.0.1",
					Port:  config.GetConf().SmartAssistant.Port,
					SaID:  config.GetConf().SmartAssistant.ID,
				},
			}
			toMsg, _ := json.Marshal(re)
			buf := encode(toMsg, token, saID)
			logger.Warn("[sa discover] write to detail buf ", buf)
			_, err := conn.WriteTo(buf, remoteAddr)
			if err != nil {
				logger.Warn("[sa discover] write to error", err)
				continue
			}
		}
	}
}

func helloResponse() []byte {
	data := make([]byte, 32)

	binary.BigEndian.PutUint16(data[0:2], 0x2131)
	binary.BigEndian.PutUint16(data[2:4], 0x20)
	binary.BigEndian.PutUint32(data[6:], saID)

	return data
}

func decode(buf []byte) (packet Packet, err error) {

	head := buf[:_packetLenOffset]
	if binary.BigEndian.Uint16(head) != 0x2131 {
		err = errors.New("invalid packet")
		return
	}

	defer func() {
		if err := recover(); err != nil {
			logger.Println("[sa discover] decode err ", err)
		}
	}()

	packetLenBytes := buf[_packetLenOffset:_packetReserveOffset]
	packet.Len = int(binary.BigEndian.Uint16(packetLenBytes)) // 包长2个字节

	if packet.Len > len(buf) {
		err = errors.New("[sa discover] invalid packet len ")
		logger.Println("[sa discover] invalid packet len ", packet.Len)
		return
	}

	// fmt.Printf("<<-- %x\n", buf[:packet.Len])
	packet.Reserve = buf[_packetReserveOffset:_packetDeviceIDOffset] // 预留2个字节

	deviceIDBytes := buf[_packetDeviceIDOffset:_packetSerialNumOffset]
	packet.DeviceID = int(binary.BigEndian.Uint32(deviceIDBytes)) // 设备ID 6个字节

	serialNumBytes := buf[_packetSerialNumOffset:_packetMD5SumOffset] // 序列号4个字节
	packet.SerialNum = int64(binary.BigEndian.Uint32(serialNumBytes))

	packet.MD5Sum = buf[_packetMD5SumOffset:_packetValidDataOffset] // MD5校验,16个字节

	// 可能有数组超出索引恐慌错误, 由上面的recover恢复
	packet.Data = buf[_packetValidDataOffset:packet.Len] // 有效数据, (包长-32)个字节

	logger.Printf("[sa discover] get packet:%++v\n", packet)

	return
}

func decrypt(enc []byte, token []byte) (result []byte, err error) {
	key, iv := getKeyAndIV(token)

	// Create a new Cipher Block from the key
	block, err := aes.NewCipher(key)
	if err != nil {
		return
	}
	bm := cipher.NewCBCDecrypter(block, iv)
	result = make([]byte, len(enc))
	bm.CryptBlocks(result, enc)
	result, err = pkcs7Unpad(result, bm.BlockSize())
	if err != nil {
		return
	}
	return
}

func pkcs7Unpad(data []byte, blockSize int) ([]byte, error) {
	length := len(data)
	if length == 0 {
		return nil, errors.New("pkcs7: Data is empty")
	}
	if length%blockSize != 0 {
		return nil, errors.New("pkcs7: Data is not block-aligned")
	}
	padLen := int(data[length-1])
	ref := bytes.Repeat([]byte{byte(padLen)}, padLen)
	if padLen > blockSize || padLen == 0 || !bytes.HasSuffix(data, ref) {
		return nil, errors.New("pkcs7: Invalid padding")
	}
	return data[:length-padLen], nil
}

func encode(msg, key []byte, deviceID uint32) (buf []byte) {

	cryptoMsg := encrypt(msg, key)
	length := len(cryptoMsg) + 32
	buf = make([]byte, length)

	binary.BigEndian.PutUint16(buf[0:2], packetHead)                                        // 包头2个字节
	binary.BigEndian.PutUint16(buf[_packetLenOffset:_packetReserveOffset], uint16(length))  // 长度2个字节
	binary.BigEndian.PutUint16(buf[_packetReserveOffset:_packetDeviceIDOffset], 0)       // 预留位2个字节
	binary.BigEndian.PutUint32(buf[_packetDeviceIDOffset:_packetSerialNumOffset], deviceID) // 设备ID6个字节
	binary.BigEndian.PutUint32(buf[_packetSerialNumOffset:_packetMD5SumOffset], 1)       // 序列号4个字节

	copy(buf[_packetValidDataOffset:], cryptoMsg) // 32后面，把加密后的信息放入去
	sum := md5Hash(buf)
	copy(buf[_packetMD5SumOffset:_packetValidDataOffset], sum[:]) // 把整个buf md5加密，放入16-32的有效位置

	return
}

func encrypt(msg, token []byte) (dst []byte) {
	key, iv := getKeyAndIV(token)

	logger.Infof("[sa discover] get key md5 string：%x", key)
	logger.Info("[sa discover] get key：", key)
	logger.Infof("[sa discover] get iv md5 string：%x", iv)
	logger.Info("[sa discover] get iv：", iv)

	// Create a new Cipher Block from the key
	block, err := aes.NewCipher(key)
	if err != nil {
		return
	}

	bm := cipher.NewCBCEncrypter(block, iv)

	msg, _ = pkcs7Pad(msg, block.BlockSize())
	dst = make([]byte, len(msg))
	bm.CryptBlocks(dst, msg)
	return
}

// pkcs7pad add pkcs7 padding
func pkcs7Pad(data []byte, blockSize int) ([]byte, error) {
	if blockSize < 0 || blockSize > 256 {
		return nil, fmt.Errorf("pkcs7: Invalid block size %d", blockSize)
	} else {
		padLen := blockSize - len(data)%blockSize
		padding := bytes.Repeat([]byte{byte(padLen)}, padLen)
		return append(data, padding...), nil
	}
}

func getKeyAndIV(t []byte) ([]byte, []byte) {
	key := md5Hash(t)
	// iv := md5Hash(md5Hash(key), t)
	iv := md5Hash(key, t)
	return key, iv
}

func md5Hash(dataBytes ...[]byte) (result []byte) {
	hash := md5.New()
	for _, data := range dataBytes {
		_, err := hash.Write(data)
		if err != nil {
			return
		}
	}
	return hash.Sum(nil)
}
