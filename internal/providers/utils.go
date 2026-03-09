// Package providers 提供云存储提供商的实现
package providers

import (
	"context"
	"io"
)

// streamReaderWithCancel 包装读取器，在关闭时取消上下文
type streamReaderWithCancel struct {
	io.ReadCloser
	cancel context.CancelFunc
}

// Close 关闭读取器并取消上下文
func (src *streamReaderWithCancel) Close() error {
	defer src.cancel()
	return src.ReadCloser.Close()
}

// newStreamReaderWithCancel 创建带取消功能的流读取器
func newStreamReaderWithCancel(reader io.ReadCloser, cancel context.CancelFunc) *streamReaderWithCancel {
	return &streamReaderWithCancel{
		ReadCloser: reader,
		cancel:     cancel,
	}
}