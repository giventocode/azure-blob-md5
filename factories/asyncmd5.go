package factories

import (
	"crypto"
	"hash"
)

//AsyncMD5 struct
type AsyncMD5 struct {
	reader DataReader
	md5    hash.Hash
}

//newAsyncMD5 TODO
func newAsyncMD5(reader DataReader) *AsyncMD5 {
	return &AsyncMD5{
		reader: reader,
		md5:    crypto.MD5.New(),
	}
}

//Source TODO
func (a *AsyncMD5) Size() int64 {
	return a.reader.Size()
}

//Source TODO
func (a *AsyncMD5) Source() string {
	return a.reader.Source()
}

//Hash TODO
func (a *AsyncMD5) Hash() ([]byte, error) {

	for read := range a.reader.Read() {
		if read.err != nil {
			return nil, read.err
		}

		_, err := a.md5.Write(read.data)

		if err != nil {
			return nil, err
		}
	}
	return a.md5.Sum(nil)[:16], nil
}
