package main

import (
	"context"
	"log"

	"github.com/zkMeLabs/mechain-go-sdk/client"
	"github.com/zkMeLabs/mechain-go-sdk/types"
)

func main() {
	account, err := types.NewAccountFromPrivateKey("test", privateKey)
	if err != nil {
		log.Fatalf("New account from private key error, %v", err)
	}
	cli, err := client.New(chainId, rpcAddr,
		client.Option{
			DefaultAccount: account,
			OffChainAuthOption: &client.OffChainAuthOption{
				Seed:                 "test_seed",
				Domain:               "https://test.domain.com",
				ShouldRegisterPubKey: true,
			},
		})
	if err != nil {
		log.Fatalf("unable to new mechain client, %v", err)
	}
	ctx := context.Background()
	// list object
	objects, err := cli.ListObjects(ctx, bucketName, types.ListObjectsOptions{
		true, "", "", "/", "", 10, "", "",
	})
	log.Println("list objects result:")
	for _, obj := range objects.Objects {
		i := obj.ObjectInfo
		log.Printf("object: %s, status: %s\n", i.ObjectName, i.ObjectStatus)
	}
}
