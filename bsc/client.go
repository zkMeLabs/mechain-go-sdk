package bsc

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"

	"github.com/ethereum/go-ethereum/ethclient"
	"google.golang.org/grpc"

	"github.com/zkMeLabs/mechain-go-sdk/bsctypes"
	"github.com/zkMeLabs/mechain-go-sdk/common"
)

// IClient - Declare all BSC SDK Client APIs, including APIs for multi messages & mechain executor
type IClient interface {
	IMultiMessageClient
	IMechainExecutorClient
	IBasicClient
	IAccountClient
}

// Client - The implementation for IClient, implement all Client APIs for Mechain SDK.
type Client struct {
	// The chain Client is used to interact with the blockchain
	chainClient *ethclient.Client
	// The HTTP Client is used to send HTTP requests to the mechain blockchain and sp
	httpClient *http.Client
	// The default account to use when sending transactions.
	defaultAccount *bsctypes.BscAccount
	// Whether the connection to the blockchain node is secure (HTTPS) or not (HTTP).
	secure bool
	// Host is the target sp server hostname，it is the host info in the request which sent to SP
	host       string
	rpcURL     string
	deployment *bsctypes.Deployment
}

// Option - Configurations for providing optional parameters for the Binance Smart Chain SDK Client.
type Option struct {
	// GrpcDialOption is the list of gRPC dial options used to configure the connection to the blockchain node.
	GrpcDialOption grpc.DialOption
	// DefaultAccount is the default account of Client.
	DefaultAccount *bsctypes.BscAccount
	// Secure is a flag that specifies whether the Client should use HTTPS or not.
	Secure bool
	// Transport is the HTTP transport used to send requests to the storage provider endpoint.
	Transport http.RoundTripper
	// Host is the target sp server hostname.
	Host string
}

func New(rpcURL string, env bsctypes.Environment, option Option) (IClient, error) {
	if rpcURL == "" {
		return nil, errors.New("fail to get grpcAddress and chainID to construct Client")
	}
	var (
		cc         *ethclient.Client
		deployment *bsctypes.Deployment
		jsonStr    string
		err        error
	)
	cc, err = ethclient.Dial(rpcURL)
	if err != nil {
		return nil, err
	}

	switch env {
	case bsctypes.BscDevnet:
		jsonStr = common.BscDevnet
	case bsctypes.BscQanet:
		jsonStr = common.BscQanet
	case bsctypes.BscTestnet:
		jsonStr = common.BscTestnet
	case bsctypes.BscMainnet:
		jsonStr = common.BscMainnet
	case bsctypes.OpBNBDevnet:
		jsonStr = common.OpBNBDevnet
	case bsctypes.OpBNBQanet:
		jsonStr = common.OpBNBQanet
	case bsctypes.OpBNBTestnet:
		jsonStr = common.OpBNBTestnet
	case bsctypes.OpBNBMainnet:
		jsonStr = common.OpBNBMainnet
	default:
		return nil, fmt.Errorf("invalid environment: %s", env)
	}

	err = json.Unmarshal([]byte(jsonStr), &deployment)
	if err != nil {
		log.Fatalf("failed to unmarshal JSON data: %v", err)
		return nil, err
	}

	c := Client{
		chainClient:    cc,
		httpClient:     &http.Client{Transport: option.Transport},
		defaultAccount: option.DefaultAccount, // it allows to be nil
		secure:         option.Secure,
		host:           option.Host,
		rpcURL:         rpcURL,
		deployment:     deployment,
	}

	return &c, nil
}
