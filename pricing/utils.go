package pricing

import (
	"bufio"
	"crypto/sha1"
	"encoding/hex"
	"errors"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

type (
	ReadCloser struct {
		io.Reader
		closefunc func() error
	}
)

func NewReadCloser(r io.Reader, closefunc func() error) *ReadCloser {
	return &ReadCloser{
		Reader:    r,
		closefunc: closefunc,
	}
}

func (rc *ReadCloser) Close() error {
	return rc.closefunc()
}

func get(urlString string) (io.ReadCloser, error) {

	cacheKey := sha1.Sum([]byte(urlString))
	cacheFile := hex.EncodeToString(cacheKey[:]) + filepath.Ext(urlString)
	if file, err := os.Open(filepath.Join(dataDir, cacheFile)); err == nil {
		return NewReadCloser(bufio.NewReader(file), file.Close), nil
	}

	resp, err := http.Get(urlString)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, errors.New(resp.Status)
	}

	file, err := os.OpenFile(filepath.Join(dataDir, cacheFile), os.O_CREATE|os.O_TRUNC|os.O_WRONLY|os.O_EXCL, 0600)
	if err != nil {
		return nil, err
	}
	return NewReadCloser(io.TeeReader(bufio.NewReader(resp.Body), file), func() error {
		resp.Body.Close()
		return file.Close()
	}), nil

}

func parseHumanReadableSize(s string) int64 {

	var m float64 = 1
	if i := strings.IndexAny(s, "bBkKmMgGtTpPeE"); i > 0 {
		switch s[i] {
		case 'k', 'K':
			m = 1 << 10
		case 'm', 'M':
			m = 1 << 20
		case 'g', 'G':
			m = 1 << 30
		case 't', 'T':
			m = 1 << 40
		case 'p', 'P':
			m = 1 << 50
		case 'e', 'E':
			m = 1 << 60
		}
		s = s[:i]
	}

	num, _ := strconv.ParseFloat(s, 64)
	return int64(num * m)

}
