package main

import (
	"fmt"
	"log"
	"sync"

	"github.com/giventocode/azure-blob-md5/factories"
	"github.com/giventocode/azure-blob-md5/internal"
)

func init() {
	internal.Options.Init()
}

func main() {

	var blobSource bool
	var fileSource bool
	var err error
	var wg sync.WaitGroup

	blobSource, fileSource, err = internal.Options.Validate()

	if internal.Options.ShowVersion {
		log.Printf("Azure Blob MD5 Tool\n Version:%s", internal.Version)
		return
	}

	if err != nil {
		log.Fatal(err)
	}

	if blobSource {
		wg.Add(1)
		executeFactory("blob", factories.NewBlobHashFactory(internal.Options.BlobNameOrPrefix,
			internal.Options.ContainerName,
			internal.Options.AccountName,
			internal.Options.AccountKey,
			internal.Options.SetBlobMD5),
			&wg)
	}

	if fileSource {
		wg.Add(1)
		executeFactory("file", factories.NewFileHashFactory(internal.Options.FileSource),
			&wg)
	}

	wg.Wait()
}

func executeFactory(sourceType string, factory <-chan factories.MD5HashResult, wg *sync.WaitGroup) {
	go func() {
		defer wg.Done()
		for hashFromSource := range factory {
			if hashFromSource.Err != nil {
				log.Fatal(hashFromSource.Err)
			}
			fmt.Printf("%s\t%d\t%x\t%s\n", hashFromSource.Source, hashFromSource.Size, hashFromSource.MD5, sourceType)
		}
		return
	}()
}
