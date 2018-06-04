package factories

import "github.com/giventocode/azure-blob-md5/internal"

//BlobReader TODO
type BlobReader struct {
	readDepth int
	az        azUtil
	blobName  string
	size      int64
}

func newBlobReader(blobName string, size int64, az azUtil) *BlobReader {
	return &BlobReader{
		readDepth: defaultReadDepth,
		blobName:  blobName,
		size:      size,
		az:        az,
	}
}

//Source TODO
func (b *BlobReader) Source() string {
	return b.blobName
}

func (b *BlobReader) Read() <-chan ReadResponse {
	response := make(chan ReadResponse, b.readDepth)

	go func() {
		defer close(response)
		bytesToRead := b.size
		count := 8 * internal.MB
		var offset int64
		for {
			if bytesToRead == 0 {
				return
			}
			if bytesToRead < count {
				count = bytesToRead
			}
			data, err := b.az.downloadRange(b.blobName, offset, count)
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
