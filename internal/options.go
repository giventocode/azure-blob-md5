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
)

func (o *options) Init() {

	flag.Usage = func() {
		printUsageDefaults("b", "blob-name-or-prefix", "", blobNameMsg)
		printUsageDefaults("c", "container-name", "", containerNameMsg)
		printUsageDefaults("a", "account-name", "", accountNameMsg)
		printUsageDefaults("k", "account-key", "", accountKeyMsg)
		printUsageDefaults("m", "set-blob-md5", "", setBlobMD5Msg)
	}

	stringVarAlias(&o.BlobNameOrPrefix, "b", "blob-name-or-prefix", "", blobNameMsg)
	stringVarAlias(&o.ContainerName, "c", "container-name", "", containerNameMsg)
	stringVarAlias(&o.AccountName, "a", "account-name", "", accountNameMsg)
	stringVarAlias(&o.AccountKey, "k", "account-key", "", accountKeyMsg)
	boolVarAlias(&o.SetBlobMD5, "m", "set-blob-md5", false, setBlobMD5Msg)
}

func (o *options) Validate() error {
	flag.Parse()

	if o.AccountName == "" {
		o.AccountName = os.Getenv(storageAccountNameEnvVar)
	}

	if o.AccountKey == "" {
		o.AccountKey = os.Getenv(storageAccountKeyEnvVar)
	}

	if o.AccountKey == "" {
		return fmt.Errorf("Storage account key is not set")
	}

	if o.AccountName == "" {
		return fmt.Errorf("Storage account name is not set")
	}

	if o.ContainerName == "" {
		return fmt.Errorf("Container name is missing")
	}

	return nil
}

func printUsageDefaults(shortflag string, longflag string, defaultVal string, description string) {
	defaultMsg := ""
	if defaultVal != "" {
		defaultMsg = fmt.Sprintf("\n\tDefault value: %v", defaultVal)
	}
	fmt.Fprintln(os.Stderr, fmt.Sprintf("-%v, --%v :\n\t%v%v", shortflag, longflag, description, defaultMsg))
}

func stringVarAlias(varPtr *string, shortflag string, longflag string, defaultVal string, description string) {
	flag.StringVar(varPtr, shortflag, defaultVal, description)
	flag.StringVar(varPtr, longflag, defaultVal, "")
}

func boolVarAlias(varPtr *bool, shortflag string, longflag string, defaultVal bool, description string) {
	flag.BoolVar(varPtr, shortflag, defaultVal, description)
	flag.BoolVar(varPtr, longflag, defaultVal, "")
}
