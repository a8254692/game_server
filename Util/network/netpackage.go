package network

import (
	"BilliardServer/Util/log"
	"encoding/binary"

	"google.golang.org/protobuf/proto"
)

const (
	LinkType_Drive   uint16 = iota //主动连接
	LinkType_Passive               //被动连接
)

// 定义buff辅助类
type MyBuff struct {
	data []byte //数据
}

// 获取数据
func (this *MyBuff) GetBytes() []byte {
	return this.data
}

// 写入数据
func (this *MyBuff) WriteInt(n int) {
	buff := make([]byte, 4)
	//写入数据长度
	binary.LittleEndian.PutUint32(buff, (uint32)(n))
	this.data = append(this.data, buff...)
}

// 写入数据
func (this *MyBuff) WriteUint16(n uint16) {
	buff := make([]byte, 2)
	//写入数据长度
	binary.LittleEndian.PutUint16(buff, n)
	this.data = append(this.data, buff...)
}

// 写入数据
func (this *MyBuff) WriteUint32(n uint32) {
	buff := make([]byte, 4)
	//写入数据长度
	binary.LittleEndian.PutUint32(buff, n)
	this.data = append(this.data, buff...)
}

// 写入字符串
func (this *MyBuff) WriteString(s string) {
	buff := []byte(s)
	this.WriteInt(len(buff))
	this.data = append(this.data, buff...)
}

// 写入byte数组
func (this *MyBuff) WriteBytes(buff []byte) {
	this.WriteInt(len(buff))
	this.data = append(this.data, buff...)
}

func (this *MyBuff) GetString() string {
	l := binary.LittleEndian.Uint32(this.data[0:4])
	return string(this.data[4 : l+4])
}
func ReadUint16String(data []byte, startIdx uint16) string {
	len := binary.LittleEndian.Uint16(data[startIdx:])
	return string(data[startIdx+2 : startIdx+2+len])
}
func ReadUint32String(data []byte, startIdx uint32) string {
	len := binary.LittleEndian.Uint32(data[startIdx:])
	return string(data[startIdx+4 : startIdx+4+len])
}

// 定义网络数据包
type MsgBody struct {
	LenBody  uint32 //数据包大小32位
	ComeType uint16 //从哪个服来
	GotoType uint16 //到哪儿去
	Data     []byte //数据
}

func (this *MsgBody) Init(comeType uint16, gotoType uint16) {
	this.ComeType = comeType
	this.GotoType = gotoType
}
func (this *MsgBody) SetData(data []byte) {
	this.Data = data
	this.LenBody = 2 + 2 + 4 + uint32(len(this.Data))
}

// 反序列化
func (this *MsgBody) Unmarshal(msg proto.Message) error {
	err := proto.Unmarshal(this.Data, msg.(proto.Message))
	if err != nil {
		log.Waring("-->MsgBody Unmarshal Error:", err)
	}
	return err
}

// 序列化
func (this *MsgBody) Marshal(param interface{}) error {
	data, err := proto.Marshal(param.(proto.Message))
	if err == nil {
		this.Data = data
	} else {
		log.Waring("-->MsgBody Marshal Error:", err)
	}
	return err
}

// 获取数据包的总数据
func (this *MsgBody) ConvertBytes() []byte {
	var sumLen uint32
	var lenComeType uint32
	var LenGotoType uint32
	sumLen = 4
	lenComeType = 2
	LenGotoType = 2
	this.LenBody = lenComeType + LenGotoType + uint32(len(this.Data))
	//构建数据缓冲
	buff := make([]byte, sumLen+this.LenBody)
	//写入数据长度
	binary.LittleEndian.PutUint32(buff, this.LenBody)
	binary.LittleEndian.PutUint16(buff[4:], this.ComeType)
	binary.LittleEndian.PutUint16(buff[6:], this.GotoType)
	copy(buff[sumLen+lenComeType+LenGotoType:], this.Data)
	return buff
}
