# Azure Blob MD5

[![Go Report Card](https://goreportcard.com/badge/github.com/giventocode/azure-blob-md5)](https://goreportcard.com/report/github.com/giventocode/azure-blob-md5)

Asynchronously computes the MD5 hash of blobs in Azure Blob Storage to maximize performance.  Azure blob MD5 computes the hash of several blobs concurrently. Local files are also supported so that you can use this functionality as a validation step for your data transfers to Azure blob storage.

![](azure-md5-blob.gif?raw=true)

## Getting Started

Pre-requisites

[Install Go](https://golang.org/dl/)

Get and build from the source.

```bash
go get github.com/giventocode/azure-blob-md5
go build -o bmd5 github.com/giventocode/azure-blob-md5
```

Set the credentials to the storage account via the environment variables.

```bash
export ACCOUNT_NAME=<YOUR_ACCOUNT_NAME>
export ACCOUNT_KEY=<YOUR_ACCOUNT_KEY>
```

## Examples

The following calculates the MD5 hash for the blob *file* in container *docs*

```bash
./bmd5 -b file -c docs
```

You can use the -m option to set the Content-MD5 property of the blob after the MD5 hash is calculated.

```bash
./bmd5 -m -b file -c docs
```

Calculating the MD5 hash for local files is supported.

```bash
./bmd5 -f file
```

Blobs and local files can be set and MD5 hashes will be calculated for both sources. The target scenario is validation of a data uploads.

```bash
./bmd5 -f file -b file -c docs
```