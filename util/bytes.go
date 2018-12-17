/// @file bytes.go
/// @brief
/// @author jackytse, xiejian1998@foxmail.com
/// @version 1.0
/// @date 2018-03-18

package util
import (
	"encoding/binary"
	"encoding/json"
	"math"
	"bytes"
)


func Int16ToBytes(i int16) []byte {
	var buf = make([]byte, 4)
	binary.LittleEndian.PutUint16(buf, uint16(i))
	return buf
}

func BytesToInt16(buf []byte) int16 {
	return int16(binary.LittleEndian.Uint16(buf))
}

func Int32ToBytes(i int32) []byte {
	var buf = make([]byte, 4)
	binary.LittleEndian.PutUint32(buf, uint32(i))
	return buf
}

func BytesToInt32(buf []byte) int32 {
	return int32(binary.LittleEndian.Uint32(buf))
}

func Int64ToBytes(i int64) []byte {
	var buf = make([]byte, 8)
	binary.LittleEndian.PutUint64(buf, uint64(i))
	return buf
}

func BytesToInt64(buf []byte) int64 {
	return int64(binary.LittleEndian.Uint64(buf))
}

func BoolToBytes(b bool) []byte {
	var buf = make([]byte, 1)
	if b == true { 
		buf[0] = 1
	}else {
		buf[0] = 0
	}
	return buf
}

func BytesToBool(buf []byte) bool {
	var data bool = (buf[0] != 0)
	return data
}

func Float32ToBytes(float float32) []byte {
	bits := math.Float32bits(float)
	bytes := make([]byte, 4)
	binary.LittleEndian.PutUint32(bytes, bits)
	return bytes
}

func BytesToFloat32(bytes []byte) float32 {
	bits := binary.LittleEndian.Uint32(bytes)
	return math.Float32frombits(bits)
}

func Float64ToBytes(float float64) []byte {
	bits := math.Float64bits(float)
	bytes := make([]byte, 8)
	binary.LittleEndian.PutUint64(bytes, bits)
	return bytes
}

func BytesToFloat64(bytes []byte) float64 {
	bits := binary.LittleEndian.Uint64(bytes)
	return math.Float64frombits(bits)
}

func MapToJsonBytes(jmap map[string]interface{}) ([]byte, error) {
	bytes, err := json.Marshal(jmap)
	return bytes, err
}

func JsonBytesToMap(bytes []byte) (map[string]interface{}, error) {
	v := make(map[string]interface{})
	err := json.Unmarshal(bytes, &v)
	return v, err
}

func BytesToString(p []byte) string {
    for i := 0; i < len(p); i++ {
        if p[i] == 0 {
            return string(p[0:i])
        }
    }
    return string(p)
}

func StringToBytes(p string) []byte {
    return []byte(p)
}

// --------------------------------------------------------------------------
/// @brief 序列化结构体原生二进制，结构体必须固定长度
/// @brief 注意：可以使用定长数组[N]byte代替string和[]byte变长字段
// --------------------------------------------------------------------------
func StructToBytes(obj interface{}) ([]byte, error) {
	buf := new(bytes.Buffer)
	err := binary.Write(buf, binary.LittleEndian, obj)
	//fmt.Println("序列化成功")
	//fmt.Println("enc len=", buf.Len())
	//fmt.Println("enc bytes", buf.Bytes())
	return buf.Bytes(), err
}

// --------------------------------------------------------------------------
/// @brief 二进制到结构体反序列化，结构体必须固定长度
/// @brief 注意：可以使用定长数组[N]byte代替string和[]byte变长字段
// --------------------------------------------------------------------------
func BytesToStruct(buf []byte, obj interface{}) error {
	bbuf := bytes.NewBuffer(buf)
	err := binary.Read(bbuf, binary.LittleEndian, obj)
	return err
}


