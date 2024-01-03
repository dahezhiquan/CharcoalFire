package utils

// 空写入器

type EmptyWriter struct{}

func (ew *EmptyWriter) Write(p []byte) (n int, err error) {
	return len(p), nil
}
