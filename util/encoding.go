/// @file encoding.go
/// @brief
/// @author jackytse, xiejian1998@foxmail.com
/// @version 1.0
/// @date 2018-07-04

package util
import (
	"fmt"
	"io"
	"encoding/hex"
	"crypto/md5"
	"crypto/sha256"
	"crypto/hmac"
)

// 方式1
func MD5(src string) string {
	signbytes := []byte(src)
	md5array  := md5.Sum(signbytes) // md5加密
	md5bytes  := []byte(md5array[:])
	return fmt.Sprintf("%x", md5bytes)
}

// 方式2
func MD5_1(src string) string {
	h := md5.New()
	h.Write([]byte(src))
	md5bytes := h.Sum(nil)
	return hex.EncodeToString(md5bytes)
}

// 方式3
func MD5_2(src string) string {
	h := md5.New()
	io.WriteString(h, src)
	return hex.EncodeToString(h.Sum(nil))
}

func SHA256(src string) string {
	signbytes := []byte(src)
	sha256array := sha256.Sum256(signbytes) // sha-256
	sha256bytes	:= []byte(sha256array[:])
	return fmt.Sprintf("%x", sha256bytes)
}

func HMAC_SHA256(key, src string) string {
	hash := hmac.New(sha256.New, []byte(key))
	io.WriteString(hash, src)
	return hex.EncodeToString(hash.Sum(nil))
}
