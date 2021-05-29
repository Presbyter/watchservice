package unit

import (
	"bufio"
	"crypto/md5"
	"errors"
	"io"
	"os"
)

func Md5Hash(file string) ([]byte, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	h := md5.New()
	r := bufio.NewReader(f)
	buf := make([]byte, 1<<12)
	for {
		n, err := r.Read(buf)
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			return nil, err
		}
		h.Write(buf[:n])
	}
	return h.Sum(nil), nil
}
