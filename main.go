package main

import (
	"fmt"
	"log"

	"github.com/giventocode/azure-blob-md5/factories"
	"github.com/giventocode/azure-blob-md5/internal"
)

func init() {
	internal.Options.Init()
}

func main() {

	if err := internal.Options.Validate(); err != nil {
		log.Fatal(err)
	}

	factory := factories.NewBlobHashFactory(internal.Options.BlobNameOrPrefix,
		internal.Options.ContainerName,
		internal.Options.AccountName, 
		internal.Options.AccountKey)

	for blobHash := range factory {

		if blobHash.Err != nil {
			log.Fatal(blobHash.Err)
		}
		fmt.Printf("%s\t%x\n", blobHash.Source, blobHash.MD5)
	}
}
