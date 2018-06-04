package factories

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"syscall"
	"time"

	"github.com/giventocode/azure-blob-md5/internal"

	"github.com/Azure/azure-pipeline-go/pipeline"

	"github.com/Azure/azure-storage-blob-go/2017-07-29/azblob"
)

//azUtil TODO
type azUtil struct {
	serviceURL   *azblob.ServiceURL
	containerURL *azblob.ContainerURL
	creds        *azblob.SharedKeyCredential
}

//newAzUtil TODO
func newAzUtil(accountName string, accountKey string, container string, baseBlobURL string) (*azUtil, error) {
	creds := azblob.NewSharedKeyCredential(accountName, accountKey)

	pipeline := newPipeline(creds, azblob.PipelineOptions{
		Retry: azblob.RetryOptions{
			Policy:        azblob.RetryPolicyFixed,
			TryTimeout:    30 * time.Second,
			MaxTries:      500,
			RetryDelay:    100 * time.Millisecond,
			MaxRetryDelay: 2 * time.Second}})

	baseURL, err := parseBaseURL(accountName, baseBlobURL)
	if err != nil {
		return nil, err
	}

	surl := azblob.NewServiceURL(*baseURL, pipeline)
	curl := surl.NewContainerURL(container)

	return &azUtil{serviceURL: &surl,
		containerURL: &curl,
		creds:        creds}, nil
}
func (p *azUtil) downloadRange(blobName string, offset int64, count int64) ([]byte, error) {
	bburl := p.containerURL.NewBlockBlobURL(blobName)
	ctx := context.Background()
	res, err := bburl.Download(ctx, offset, count, azblob.BlobAccessConditions{}, false)

	if err != nil {
		return nil, err
	}

	opts := azblob.RetryReaderOptions{
		MaxRetryRequests: 30,
	}
	reader := res.Body(opts)

	data := make([]byte, count)
	tmp := make([]byte, count)

	//n, err := reader.Read(data)
	wr := bytes.NewBuffer(data)
	n, err := io.CopyBuffer(wr, reader, tmp)

	defer reader.Close()
	if err != nil {
		return nil, err
	}
	data = wr.Bytes()[count:]
	if n != count {
		return nil, fmt.Errorf(" received data len is different than expected. Expected:%d Received:%d ", count, n)
	}

	return data, nil
}
func (p *azUtil) blobExists(bburl azblob.BlockBlobURL) (bool, error) {
	ctx := context.Background()
	response, err := bburl.GetProperties(ctx, azblob.BlobAccessConditions{})

	if response != nil {
		defer response.Response().Body.Close()
		if response.StatusCode() == 200 {
			return true, nil
		}
	}

	storageErr, ok := err.(azblob.StorageError)

	if ok {
		errResp := storageErr.Response()
		if errResp != nil {
			defer errResp.Body.Close()
			if errResp.StatusCode == 404 {
				return false, nil
			}
		}
	}

	return false, err
}

//GetBlobURLWithReadOnlySASToken  TODO
func (p *azUtil) GetBlobURLWithReadOnlySASToken(blobName string, expTime time.Time) url.URL {
	bu := p.containerURL.NewBlobURL(blobName)
	bp := azblob.NewBlobURLParts(bu.URL())

	sas := azblob.BlobSASSignatureValues{BlobName: blobName,
		ContainerName: bp.ContainerName,
		ExpiryTime:    expTime,
		Permissions:   "r"}

	sq := sas.NewSASQueryParameters(p.creds)
	bp.SAS = sq
	return bp.URL()
}

//BlobCallback TODO
type BlobCallback func(*azblob.Blob, string) (bool, error)

//BlobItemInfo TODO
type BlobItemInfo struct {
	Blob azblob.Blob
	Err  error
}

//IterateBlobList TODO
func (p *azUtil) IterateBlobList(prefix string, chanDepth int) <-chan BlobItemInfo {

	blobs := make(chan BlobItemInfo, chanDepth)

	var marker azblob.Marker
	options := azblob.ListBlobsSegmentOptions{
		Details: azblob.BlobListingDetails{
			Metadata: true},
		Prefix: prefix}

	go func() {
		defer close(blobs)

		for {
			ctx := context.Background()
			response, err := p.containerURL.ListBlobsFlatSegment(ctx, marker, options)

			if err != nil {
				blobs <- BlobItemInfo{Err: err}
				return
			}
			for _, blob := range response.Blobs.Blob {
				blobs <- BlobItemInfo{Blob: blob}
			}

			if response.NextMarker.NotDone() {
				marker = response.NextMarker
				continue
			}

			break

		}
	}()

	return blobs
}

func parseBaseURL(accountName string, baseURL string) (*url.URL, error) {
	var err error
	var url *url.URL

	if baseURL == "" {
		url, err = url.Parse(fmt.Sprintf("https://%s.blob.core.windows.net", accountName))

		if err != nil {
			return nil, err
		}

		return url, nil
	}

	if url, err = url.Parse(fmt.Sprintf("https://%s.%s", accountName, baseURL)); err != nil {
		return nil, err
	}

	return url, nil

}

func newPipeline(c azblob.Credential, o azblob.PipelineOptions) pipeline.Pipeline {
	if c == nil {
		panic("c can't be nil")
	}

	// Closest to API goes first; closest to the wire goes last
	f := []pipeline.Factory{
		azblob.NewTelemetryPolicyFactory(o.Telemetry),
		azblob.NewUniqueRequestIDPolicyFactory(),
		azblob.NewRetryPolicyFactory(o.Retry),
		c,
	}

	f = append(f,
		pipeline.MethodFactoryMarker(), // indicates at what stage in the pipeline the method factory is invoked
		azblob.NewRequestLogPolicyFactory(o.RequestLog))

	return pipeline.NewPipeline(f, pipeline.Options{HTTPSender: newHTTPClientFactory(), Log: o.Log})
}

func newHTTPClientFactory() pipeline.Factory {
	return &clientPolicyFactory{}
}

type clientPolicyFactory struct {
}

// Create initializes a logging policy object.
func (f *clientPolicyFactory) New(next pipeline.Policy, po *pipeline.PolicyOptions) pipeline.Policy {
	return &clientPolicy{po: po}
}

type clientPolicy struct {
	po *pipeline.PolicyOptions
}

const winWSAETIMEDOUT syscall.Errno = 10060

// checks if the underlying error is a connectex error and if the underying cause is winsock timeout or temporary error, in which case we should retry.
func isWinsockTimeOutError(err error) net.Error {
	if uerr, ok := err.(*url.Error); ok {
		if derr, ok := uerr.Err.(*net.OpError); ok {
			if serr, ok := derr.Err.(*os.SyscallError); ok && serr.Syscall == "connectex" {
				if winerr, ok := serr.Err.(syscall.Errno); ok && (winerr == winWSAETIMEDOUT || winerr.Temporary()) {
					return &retriableError{error: err}
				}
			}
		}
	}
	return nil
}

func isDialConnectError(err error) net.Error {
	if uerr, ok := err.(*url.Error); ok {
		if derr, ok := uerr.Err.(*net.OpError); ok {
			if serr, ok := derr.Err.(*os.SyscallError); ok && serr.Syscall == "connect" {
				return &retriableError{error: err}
			}
		}
	}
	return nil
}

func isRetriableDialError(err error) net.Error {
	if derr := isWinsockTimeOutError(err); derr != nil {
		return derr
	}
	return isDialConnectError(err)
}

type retriableError struct {
	error
}

func (*retriableError) Timeout() bool {
	return false
}

func (*retriableError) Temporary() bool {
	return true
}

const tcpKeepOpenMinLength = 8 * int64(internal.MB)

func (p *clientPolicy) Do(ctx context.Context, request pipeline.Request) (pipeline.Response, error) {
	req := request.WithContext(ctx)

	if req.ContentLength < tcpKeepOpenMinLength {
		req.Close = true
	}

	r, err := pipelineHTTPClient.Do(req)
	pipresp := pipeline.NewHTTPResponse(r)
	if err != nil {
		if derr := isRetriableDialError(err); derr != nil {
			return pipresp, derr
		}
		err = pipeline.NewError(err, "HTTP request failed")
	}
	return pipresp, err
}

var pipelineHTTPClient = newpipelineHTTPClient()

func newpipelineHTTPClient() *http.Client {

	return &http.Client{
		Transport: &http.Transport{
			Proxy: http.ProxyFromEnvironment,
			Dial: (&net.Dialer{
				Timeout:   30 * time.Second,
				KeepAlive: 30 * time.Second,
				DualStack: true,
			}).Dial,
			MaxIdleConns:           100,
			MaxIdleConnsPerHost:    100,
			IdleConnTimeout:        60 * time.Second,
			TLSHandshakeTimeout:    10 * time.Second,
			ExpectContinueTimeout:  1 * time.Second,
			DisableKeepAlives:      false,
			DisableCompression:     false,
			MaxResponseHeaderBytes: 0}}

}
