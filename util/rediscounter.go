/// @file rediscounter.go
/// @brief
/// @author jackytse, xiejian1998@foxmail.com
/// @version 1.0
/// @date 2018-05-11

package util
import "github.com/go-redis/redis"
import "time"
import "fmt"

type RedisCounter struct {
	set map[string]int32
	rds *redis.Client
}

func NewRedisCounter() *RedisCounter {
	tc := &RedisCounter{}
	tc.set = make(map[string]int32)
	tc.rds = nil
	return tc
}

func (r *RedisCounter) Init(rds *redis.Client) {
	r.set = make(map[string]int32)
	r.rds = rds
}

func (r *RedisCounter) Set(key string, num int32) {
	r.set[key] = num
}

func (r *RedisCounter) IncrBy(key string, num int32) {
	r.set[key] += num
}

func (r *RedisCounter) IncrByDate(keypart string, id int32, num int32) {
	datetime := time.Now().Format("2006-01-02")
	key := fmt.Sprintf("%s_%d_%s", datetime, id, keypart)
	r.set[key] += num
}

func (r *RedisCounter) DecrBy(key string, num int32) {
	r.set[key] -= num
}

// Deprecated!
func (r *RedisCounter) Save() {
	r.DBSave()
}

// Deprecated!
func (r *RedisCounter) BatchSave(num int32) {
	r.DBBatchSave(num)
}

// 全部存储
func (r *RedisCounter) DBSave() {
	for k , v := range r.set {
		r.rds.IncrBy(k, int64(v))
	}
	r.set = make(map[string]int32)
}

// 分批存储
func (r *RedisCounter) DBBatchSave(num int32) {
	var count int32 = 0
	for k , v := range r.set {
		r.rds.IncrBy(k, int64(v))
		delete(r.set, k)
		if count++; count >= num {
			break
		}
	}
	//r.set = make(map[string]int32)
}

