package main

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"time"

	"github.com/bnb-chain/greenfield-go-sdk/client"
	"github.com/bnb-chain/greenfield-go-sdk/types"
)

// it is the example of basic storage SDKs usage
func main() {
	account, err := types.NewAccountFromPrivateKey("test", privateKey)
	if err != nil {
		log.Fatalf("New account from private key error, %v", err)
	}
	// cli, err := client.New(chainId, rpcAddr, client.Option{DefaultAccount: account})
	cli, err := client.New(chainId, rpcAddr, client.Option{
		DefaultAccount: account,
		OffChainAuthOption: &client.OffChainAuthOption{
			Seed:                 "test_seed",
			Domain:               "https://test.domain.com",
			ShouldRegisterPubKey: true,
		},
	})
	if err != nil {
		log.Fatalf("unable to new greenfield client, %v", err)
	}
	ctx := context.Background()

	// get storage providers list
	spLists, err := cli.ListStorageProviders(ctx, true)
	if err != nil {
		log.Fatalf("fail to list in service sps")
	}
	// choose the first sp to be the primary SP
	primarySP := spLists[6].GetOperatorAddress()
	log.Println(primarySP)
	// create bucket
	_, err = cli.CreateBucket(ctx, bucketName, primarySP, types.CreateBucketOptions{})
	handleErr(err, "CreateBucket")
	log.Printf("create bucket %s on SP: %s successfully \n", bucketName, spLists[0].Endpoint)

	// head bucket
	bucketInfo, err := cli.HeadBucket(ctx, bucketName)
	handleErr(err, "HeadBucket")
	log.Println("bucket info:", bucketInfo.String())

	// Create object content
	var buffer bytes.Buffer
	line := `0123456789`
	for i := 0; i < objectSize/10; i++ {
		buffer.WriteString(fmt.Sprintf("%s", line))
	}

	// create and put object
	txnHash, err := cli.CreateObject(ctx, bucketName, objectName, bytes.NewReader(buffer.Bytes()), types.CreateObjectOptions{})
	handleErr(err, "CreateObject")

	err = cli.PutObject(ctx, bucketName, objectName, int64(buffer.Len()),
		bytes.NewReader(buffer.Bytes()), types.PutObjectOptions{TxnHash: txnHash})
	handleErr(err, "PutObject")

	log.Printf("object: %s has been uploaded to SP\n", objectName)

	waitObjectSeal(cli, bucketName, objectName)
	time.Sleep(time.Second * 5)
	// get object
	reader, info, err := cli.GetObject(ctx, bucketName, objectName, types.GetObjectOptions{})
	handleErr(err, "GetObject")
	log.Printf("get object %s successfully, size %d \n", info.ObjectName, info.Size)
	handleErr(err, "GetObject")
	objectBytes, err := io.ReadAll(reader)
	if !bytes.Equal(objectBytes, buffer.Bytes()) {
		handleErr(errors.New("download content not same"), "GetObject")
	}

	// list object
	objects, err := cli.ListObjects(ctx, bucketName, types.ListObjectsOptions{
		true, "", "", "/", "", 10, httpsAddr, "",
	})
	log.Println("list objects result:")
	for _, obj := range objects.Objects {
		i := obj.ObjectInfo
		log.Printf("object: %s, status: %s\n", i.ObjectName, i.ObjectStatus)
	}

	// list object by object ids
	ids := []uint64{1, 2, 333}
	objects2, err := cli.ListObjectsByObjectID(ctx, ids, types.EndPointOptions{
		Endpoint:  httpsAddr,
		SPAddress: "",
	})
	log.Printf("list objects by ids result: %v\n", objects2)
	for _, object := range objects2.Objects {
		if object != nil {
			log.Printf("object: %s, status: %s\n", object.ObjectInfo.ObjectName, object.ObjectInfo.ObjectStatus)
		}
	}

	// list buckets by bucket ids
	buckets, err := cli.ListBucketsByBucketID(ctx, ids, types.EndPointOptions{
		Endpoint:  httpsAddr,
		SPAddress: "",
	})
	log.Printf("list buckets by ids result: %v\n", buckets)
	for _, bucket := range buckets.Buckets {
		if bucket != nil {
			log.Printf("bucket: %s, status: %s\n", bucket.BucketInfo.BucketName, bucket.BucketInfo.BucketStatus)
		}
	}
	log.Printf("object: %s has been deleted\n", objectName)
}
