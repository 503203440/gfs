package models

import "io"

// 定义一个结构实现io.Writer接口
type MultiWriter struct {
	writers []io.Writer
}

// 让MultiWriter实现io.Write()接口
// MultiWriter的Write方法就是把自身的writers的都Write一遍
func (mw *MultiWriter) Write(p []byte) (n int, err error) {
	for _, w := range mw.writers {
		n, err = w.Write(p)
		if err != nil {
			return n, err
		}
	}
	return len(p), nil
}

func NewMultiWrite(write ...io.Writer) *MultiWriter {
	return &MultiWriter{
		writers: write,
	}
}
