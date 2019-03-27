package ioutil

import (
	"io"

	"github.com/zpab123/sco/scoerr" // 错误
)

// 将 data 数据 全部写入 conn
func WriteAll(conn io.Writer, data []byte) error {
	left := len(data)
	for left > 0 {
		n, err := conn.Write(data)
		if n == left && err == nil { // handle most common case first
			return nil
		}

		if n > 0 {
			data = data[n:]
			left -= n
		}

		if err != nil && !scoerr.IsTimeoutError(err) {
			return err
		}
	}

	return nil
}

// 持续读取数据，直到data容量为0
func ReadAll(conn io.Reader, data []byte) error {
	left := len(data)
	for left > 0 {
		n, err := conn.Read(data)
		if n == left && err == nil { // handle most common case first
			return nil
		}

		if n > 0 {
			data = data[n:]
			left -= n
		}

		if err != nil && !scoerr.IsTimeoutError(err) {
			return err
		}
	}

	return nil
}
