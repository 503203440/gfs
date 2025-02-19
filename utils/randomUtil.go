package utils

import "math/rand/v2"

// 生成随机字符串
func GenerateRandomString(length int) string {
	// 定义字符集
	charSet := "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	// 初始化随机数种子
	// rand.Seed(time.Now().UnixNano())
	// 创建一个字符串缓冲区
	var buffer [256]byte
	// 生成随机字符串
	for i := 0; i < length; i++ {
		buffer[i] = charSet[rand.IntN(len(charSet))]
	}
	return string(buffer[:length])
}
