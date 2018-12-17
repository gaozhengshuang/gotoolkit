/// @file rediswrapper.go
/// @brief
/// @author jackytse, xiejian1998@foxmail.com
/// @version 1.0
/// @date 2017-11-07

package utredis
import (
	"github.com/go-redis/redis"
	pb "github.com/golang/protobuf/proto"
)


//type ISerializableData interface {
//	Marshal()
//	Unmarshal()
//}

// --------------------------------------------------------------------------
/// @brief Set, Get
// --------------------------------------------------------------------------
// 设置 protobuff value
func SetProtoBin(rds *redis.Client, key string, data pb.Message) error {

	// proto 序列化
	buf, err := pb.Marshal(data)
	if err != nil { return err }

	// Set二进制
	err = rds.Set(key, buf, 0).Err()
	return err
}

// 设置 protobuff value by pipeline
func SetProtoBinPipeline(pipe redis.Pipeliner, key string, data pb.Message) error {

	// proto 序列化
	buf, err := pb.Marshal(data)
	if err != nil { return err }

	// Set二进制
	err = pipe.Set(key, buf, 0).Err()
	return err
}

func GetProtoBin(rds *redis.Client, key string, data pb.Message) error {
	// Get二进制
	str, err := rds.Get(key).Result()
	if err != nil { return err }

	// 反序列化
	rbuf :=[]byte(str)
	err = pb.Unmarshal(rbuf, data)
	return err
}

//func GetProtoBinPipeline(pipe redis.Pipeliner, key string, data pb.Message) error {
//	// Get二进制
//	str, err := pipe.Get(key).Result()
//	if err != nil { return err }
//
//	// 反序列化
//	rbuf :=[]byte(str)
//	err = pb.Unmarshal(rbuf, data)
//	return err
//}



// --------------------------------------------------------------------------
/// @brief HSet, HGet
// --------------------------------------------------------------------------
func HSetProtoBin(rds *redis.Client, key, field string, data pb.Message) error {
	// proto 序列化
	buf, err := pb.Marshal(data)
	if err != nil { return err }

	// Set二进制
	err = rds.HSet(key, field, buf).Err()
	return err
}

func HSetProtoBinPipeline(pipe redis.Pipeliner, key, field string, data pb.Message) error {
	// proto 序列化
	buf, err := pb.Marshal(data)
	if err != nil { return err }

	// Set二进制
	err = pipe.HSet(key, field, buf).Err()
	return err
}

