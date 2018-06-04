package factories

import (
	"os"

	"github.com/giventocode/azure-blob-md5/internal"
)

//FileReader TODO
type FileReader struct {
	readDepth int
	fileName  string
	size      int64
}

func newFileReader(fileName string, size int64) *FileReader {
	return &FileReader{
		readDepth: defaultReadDepth,
		fileName:  fileName,
		size:      size,
	}
}

//Source TODO
func (b *FileReader) Source() string {
	return b.fileName
}
func (b *FileReader) Read() <-chan ReadResponse {
	response := make(chan ReadResponse, b.readDepth)

	go func() {
		defer close(response)
		bytesToRead := b.size
		count := 8 * internal.MB
		var offset int64
		fh, err := os.Open(b.fileName)

		if err != nil {
			response <- ReadResponse{err: err}
			return
		}

		for {
			if bytesToRead == 0 {
				return
			}
			if bytesToRead < count {
				count = bytesToRead
			}
			data := make([]byte, count)
			_, err := fh.ReadAt(data, offset)
			if err != nil {
				response <- ReadResponse{err: err}
				return
			}
			response <- ReadResponse{data: data}
			offset = offset + count
			bytesToRead = bytesToRead - count
		}
	}()

	return response
}
