package internal

import (
	"fmt"
	"os"

	flag "github.com/spf13/pflag"
)

type options struct {
	BlobNameOrPrefix string
	ContainerName    string
	AccountName      string
	AccountKey       string
	SetBlobMD5       bool
	FileSource       string
	ShowVersion      bool
}

//Options TODO
var Options = &options{}

const storageAccountKeyEnvVar = "ACCOUNT_KEY"
const storageAccountNameEnvVar = "ACCOUNT_NAME"


const (
	blobNameMsg      = "Blob name (e.g. myblob.txt) or prefix."
	containerNameMsg = "Container name (e.g. mycontainer)."
	accountNameMsg   = "Storage account name (e.g. mystorage).\n\tCan also be specified via the " + storageAccountNameEnvVar + " environment variable."
	accountKeyMsg    = "Storage account key string.\n\tCan also be specified via the " + storageAccountKeyEnvVar + " environment variable."
	setBlobMD5Msg    = "Set Content-MD5 property of the blob with the calculated value"
	fileSourceMsg    = "File name or pattern. If set, the MD5 hash will be calculated for the files that match the criteria"
	showVersionMsg   = "Display current version"
)

func (o *options) Init() {

	flag.Usage = func() {
		printUsageDefaults("b", "blob-name-or-prefix", "", blobNameMsg)
		printUsageDefaults("c", "container-name", "", containerNameMsg)
		printUsageDefaults("a", "account-name", "", accountNameMsg)
		printUsageDefaults("k", "account-key", "", accountKeyMsg)
		printUsageDefaults("m", "set-blob-md5", "", setBlobMD5Msg)
		printUsageDefaults("f", "file-source-pattern", "", fileSourceMsg)
		printUsageDefaults("v", "version", "", showVersionMsg)
	}
	flag.BoolVarP(&o.SetBlobMD5, "set-blob-md5", "m", false, setBlobMD5Msg)
	flag.BoolVarP(&o.ShowVersion, "version", "v", false, showVersionMsg)
	flag.StringVarP(&o.BlobNameOrPrefix, "blob-name-or-prefix", "b", "", blobNameMsg)
	flag.StringVarP(&o.ContainerName, "container-name", "c", "", containerNameMsg)
	flag.StringVarP(&o.AccountName, "account-name", "a", os.Getenv(storageAccountNameEnvVar), accountNameMsg)
	flag.StringVarP(&o.AccountKey, "account-key", "k", os.Getenv(storageAccountKeyEnvVar), accountKeyMsg)
	flag.StringVarP(&o.FileSource, "file-source-pattern", "f", "", fileSourceMsg)

}

func (o *options) Validate() (blobSource bool, fileSource bool, err error) {
	flag.Parse()

	errBlobSource := o.validateBlobSource()
	errFileSource := o.validateFileSource()

	if errBlobSource != nil && errFileSource != nil {
		return false, false, fmt.Errorf(" Invalid options. A file source or a blob source be set.\nFile Source:\n%v\nBlobSource:\n%v", errFileSource, errBlobSource)
	}

	if errBlobSource != nil && errFileSource == nil {
		return false, true, nil
	}

	if errBlobSource == nil {
		return true, errFileSource == nil, nil
	}

	//it should not get here...
	return
}

func (o *options) validateFileSource() error {
	if o.FileSource == "" {
		return fmt.Errorf("File source pattern not specified via option -f")
	}

	return nil
}
func (o *options) validateBlobSource() error {
	var err error

	if o.AccountKey == "" {
		err = fmt.Errorf("Storage account key is not set")
	}

	if o.AccountName == "" {
		err = fmt.Errorf("Storage account name is not set\n%v", err)
	}

	if o.ContainerName == "" {
		err = fmt.Errorf("Container name is missing\n%v", err)
	}

	return err
}

func printUsageDefaults(shortflag string, longflag string, defaultVal string, description string) {
	defaultMsg := ""
	if defaultVal != "" {
		defaultMsg = fmt.Sprintf("\n\tDefault value: %v", defaultVal)
	}
	fmt.Fprintln(os.Stderr, fmt.Sprintf("-%v, --%v :\n\t%v%v", shortflag, longflag, description, defaultMsg))
}
