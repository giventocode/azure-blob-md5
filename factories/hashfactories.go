package factories

import (
	"log"
	"os"
	"path/filepath"
)

const defaultReadDepth int = 10
const defaultHashResultsDepth int = 6

//ReadResponse response from a read operation
type ReadResponse struct {
	data []byte
	err  error
}

//DataReader reads sequentially and asynchronously from a source.
type DataReader interface {
	Source() string
	Read() <-chan ReadResponse
	Size() int64
}

//MD5HashResult TODO
type MD5HashResult struct {
	Source string
	MD5    []byte
	Err    error
	Size int64
}

//NewBlobHashFactory todo
func NewBlobHashFactory(pattern string, container string, accountName string, accountKey string) <-chan MD5HashResult {
	az, err := newAzUtil(accountName, accountKey, container, "")
	if err != nil {
		log.Fatal(err)
	}

	factory := func() <-chan AsyncMD5 {
		blobMD5 := make(chan AsyncMD5, 1000)
		go func() {
			defer close(blobMD5)
			for blobItem := range az.IterateBlobList(pattern, 1000) {
				if blobItem.Err != nil {
					log.Fatal(blobItem.Err)
				}
				blobReader := newBlobReader(blobItem.Blob.Name, *blobItem.Blob.Properties.ContentLength, *az)
				blobMD5 <- *newAsyncMD5(blobReader)
			}
		}()
		return blobMD5
	}

	return md5Hash(factory())

}

//NewFileHashFactory TODO
func NewFileHashFactory(pattern string) <-chan MD5HashResult {
	files, err := filepath.Glob(pattern)
	if err != nil {
		log.Fatal(err)
	}

	factory := func() <-chan AsyncMD5 {
		fileMD5 := make(chan AsyncMD5, 1000)
		go func() {
			defer close(fileMD5)

			for _, file := range files {
				fileStat, err := os.Stat(file)

				if err != nil {
					log.Fatal(err)
				}

				if fileStat.IsDir() {
					err := filepath.Walk(file, func(path string, f os.FileInfo, err error) error {
						if err != nil {
							return err
						}

						if !f.IsDir() {
							fullFileName := filepath.Join(path, f.Name())

							fileReader := newFileReader(fullFileName, f.Size())
							fileMD5 <- *newAsyncMD5(fileReader)
						}

						return nil
					})

					if err != nil {
						log.Fatal(err)
					}
					continue
				}

				fileReader := newFileReader(file, fileStat.Size())
				fileMD5 <- *newAsyncMD5(fileReader)
			}
		}()
		return fileMD5
	}

	return md5Hash(factory())

}
func md5Hash(hashfactory <-chan AsyncMD5) <-chan MD5HashResult {
	results := make(chan MD5HashResult, defaultHashResultsDepth)

	go func() {
		defer close(results)
		for hashItem := range hashfactory {
			md5, err := hashItem.Hash()

			if err != nil {
				results <- MD5HashResult{Err: err}
				return
			}

			results <- MD5HashResult{MD5: md5, Source: hashItem.Source()}
		}
	}()

	return results
}
