package utils

import (
	"crypto/hmac"
	"crypto/md5"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"sort"
	"strings"
)

// 计算文本内容SHA256
func generateHMACSHA256(message, key string) string {

	// time.Sleep(time.Second * 3)

	h := hmac.New(sha256.New, []byte(key))

	h.Write([]byte(message))

	sum := h.Sum(nil)

	return strings.ToUpper(hex.EncodeToString(sum))
}

// 计算文本内容的MD5
func generateMD5(message, key string) string {
	h := hmac.New(md5.New, []byte(key))
	h.Write([]byte(message))
	sum := h.Sum(nil)
	return strings.ToUpper(hex.EncodeToString(sum))
}

// 定义一个SignType的类型,可选的值为const里面的内容
type SignType string

const (
	HMACMD5    SignType = "md5"
	HMACSHA256 SignType = "sha256"
)

// 将data中的数据按照key的字典排序然后使用&连接
func GetSign(data map[string]any, key string, signType SignType) string {

	keys := make([]string, 0, len(data))

	for key := range data {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	appendStr := ""

	for _, k := range keys {
		// 如果key为sign或者值为空, 则不参数计算
		if k == "sign" || data[k] == nil {
			continue
		}
		appendStr += fmt.Sprintf("%v=%v&", k, data[k])
	}
	// 移除最后一个&
	if len(appendStr) > 0 {
		appendStr = appendStr[:len(appendStr)-1]
	}

	if signType == HMACMD5 {
		return generateMD5(appendStr, key)
	} else {
		return generateHMACSHA256(appendStr, key)
	}

}
