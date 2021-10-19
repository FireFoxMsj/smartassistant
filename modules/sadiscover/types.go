package sadiscover

const (
	packetHead = 0x2131

	reserveTokenKey = 0xfffe

	_packetHeaderOffset = 0 // 包头偏移
	_packetHeaderSize   = 2 // 包头大小

	_packetLenOffset = _packetHeaderOffset + _packetHeaderSize // 包长度偏移
	_packetLenSize   = 2                                       // 包长度大小

	_packetReserveOffset = _packetLenOffset + _packetLenSize // 预留偏移
	_packetReserveSize   = 2                                 // 预留大小，2个字节

	_packetDeviceIDOffset = _packetReserveOffset + _packetReserveSize // 设备id偏移
	_packetDeviceIDSize   = 6                                         // 设备id大小，4哥字节

	_packetSerialNumOffset = _packetDeviceIDOffset + _packetDeviceIDSize // 序列号偏移
	_packetSerialNumSize   = 4                                           // 序列号大小

	_packetMD5SumOffset = _packetSerialNumOffset + _packetSerialNumSize // md5校验偏移
	_packetMD5SumSize   = 16                                            // md5校验大小

	_packetValidDataOffset = _packetMD5SumOffset + _packetMD5SumSize // 有效数据偏移
	_packetValidDataSize   = 32                                      // 有效数据大小
)

type Info struct {
	Model string `json:"model"`
	SwVer string `json:"sw_ver"`
	HwVer string `json:"hw_ver"`
	Port  int    `json:"port"`
	SaID  string `json:"sa_id"`
}

type Result struct {
	ID int  `json:"id"`
	Re Info `json:"result"`
}

type Protocol struct {
}

type Packet struct {
	Head      []byte
	Len       int
	Reserve   []byte
	DeviceID  int
	SerialNum int64
	MD5Sum    []byte
	Data      []byte
}
