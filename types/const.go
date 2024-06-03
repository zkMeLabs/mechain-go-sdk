package types

import (
	"os"
	"runtime"
	"time"
)

const (
	libName   = "greenfield-go-sdk"
	Version   = "v0.1.0"
	UserAgent = "Greenfield (" + runtime.GOOS + "; " + runtime.GOARCH + ") " + libName + "/" + Version

	HTTPHeaderAuthorization = "Authorization"

	HTTPHeaderContentLength   = "Content-Length"
	HTTPHeaderContentMD5      = "Content-MD5"
	HTTPHeaderContentType     = "Content-Type"
	HTTPHeaderTransactionHash = "X-Gnfd-Txn-Hash"
	HTTPHeaderUnsignedMsg     = "X-Gnfd-Unsigned-Msg"
	HTTPHeaderSignedMsg       = "X-Gnfd-Signed-Msg"
	HTTPHeaderPieceIndex      = "X-Gnfd-Piece-Index"
	HTTPHeaderRedundancyIndex = "X-Gnfd-Redundancy-Index"
	HTTPHeaderObjectID        = "X-Gnfd-Object-ID"
	HTTPHeaderIntegrityHash   = "X-Gnfd-Integrity-Hash"
	HTTPHeaderPieceHash       = "X-Gnfd-Piece-Hash"

	HTTPHeaderDate          = "X-Gnfd-Date"
	HTTPHeaderEtag          = "ETag"
	HTTPHeaderRange         = "Range"
	HTTPHeaderUserAgent     = "User-Agent"
	HTTPHeaderContentSHA256 = "X-Gnfd-Content-Sha256"

	HTTPHeaderUserAddress = "X-Gnfd-User-Address"

	ContentTypeXML = "application/xml"
	ContentDefault = "application/octet-stream"

	// EmptyStringSHA256 is the hex encoded sha256 value of an empty string
	EmptyStringSHA256       = `e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855`
	Iso8601DateFormatSecond = "2006-01-02T15:04:05Z"

	AdminURLPrefix    = "/greenfield/admin"
	AdminURLV1Version = "/v1"
	AdminURLV2Version = "/v2"
	AdminV1Version    = 1
	AdminV2Version    = 2

	CreateObjectAction  = "CreateObject"
	CreateBucketAction  = "CreateBucket"
	MigrateBucketAction = "MigrateBucket"

	ChallengeUrl           = "challenge"
	PrimaryRedundancyIndex = -1

	ContextTimeout   = time.Second * 30
	MaxHeadTryTime   = 4
	HeadBackOffDelay = time.Millisecond * 500
	NoSuchObjectErr  = "object has not been created"

	GetConnectionFail = "connection refused"

	MaxDownloadTryTime   = 3
	DownloadBackOffDelay = time.Millisecond * 500

	// MinPartSize - minimum part size 32MiB per object after which
	// putObject behaves internally as multipart.
	MinPartSize = 1024 * 1024 * 32

	TempFileSuffix = ".temp"            // Temp file suffix
	FilePermMode   = os.FileMode(0o664) // Default file permission

	WaitTxContextTimeOut = 1 * time.Second
	DefaultExpireSeconds = 1000
)
