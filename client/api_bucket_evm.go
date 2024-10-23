package client

import (
	"context"
	"encoding/xml"
	"fmt"
	"io"
	"math/big"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/tx"
	gnfdsdk "github.com/evmos/evmos/v12/sdk/types"
	"github.com/evmos/evmos/v12/testutil/storage"
	gnfdTypes "github.com/evmos/evmos/v12/types"
	storageTypes "github.com/evmos/evmos/v12/x/storage/types"
	"github.com/rs/zerolog/log"
	"github.com/zkMeLabs/mechain-go-sdk/pkg/utils"
	"github.com/zkMeLabs/mechain-go-sdk/types"
)

func (c *Client) createBucketEvm(ctx context.Context, bucketName string, primaryAddr string, opts types.CreateBucketOptions) (string, error) {
	address, err := sdk.AccAddressFromHexUnsafe(primaryAddr)
	if err != nil {
		return "", err
	}

	var visibility storageTypes.VisibilityType
	if opts.Visibility == storageTypes.VISIBILITY_TYPE_UNSPECIFIED {
		visibility = storageTypes.VISIBILITY_TYPE_PRIVATE // set default visibility type
	} else {
		visibility = opts.Visibility
	}

	var paymentAddr sdk.AccAddress
	if opts.PaymentAddress != "" {
		paymentAddr, err = sdk.AccAddressFromHexUnsafe(opts.PaymentAddress)
		if err != nil {
			return "", err
		}
	}

	createBucketMsg := storageTypes.NewMsgCreateBucket(c.MustGetDefaultAccount().GetAddress(), bucketName, visibility, address, paymentAddr, 0, nil, opts.ChargedQuota)

	err = createBucketMsg.ValidateBasic()
	if err != nil {
		return "", err
	}

	accAddress, err := sdk.AccAddressFromHexUnsafe(primaryAddr)
	if err != nil {
		return "", err
	}

	sp, err := c.GetStorageProviderInfo(ctx, accAddress)
	if err != nil {
		return "", err
	}

	familyID, err := c.GetRecommendedVirtualGroupFamilyIDBySPID(ctx, sp.Id)
	if err != nil {
		log.Error().Msg(fmt.Sprintf("failed to query sp vgf:  %s", err.Error()))
		var signedMsg *storageTypes.MsgCreateBucket
		signedMsg, err = c.GetCreateBucketApproval(ctx, createBucketMsg)
		if err != nil {
			return "", err
		}
		familyID = signedMsg.PrimarySpApproval.GlobalVirtualGroupFamilyId
	}

	// evm
	nonce, err := c.chainClient.GetNonce(context.Background())
	if err != nil {
		return "", err
	}
	var (
		hexPrivateKey string
		chain         *big.Int
		gasLimit      uint64
	)
	txOpts, err := CreateTxOpts(ctx, c.evmClient, hexPrivateKey, chain, gasLimit, nonce)
	if err != nil {
		return "", err
	}

	session, err := CreateStorageSession(c.evmClient, *txOpts, mtypes.StorageAddress)
	if err != nil {
		return "", err
	}
	approval := storage.Approval{
		exp
	}
	txRsp, err := session.CreateBucket(bucketName, uint8(visibility), common.Address(paymentAddr), common.Address(address))
	// evm end
	createBucketMsg.PrimarySpApproval.GlobalVirtualGroupFamilyId = familyID

	// set the default txn broadcast mode as block mode
	if opts.TxOpts == nil {
		broadcastMode := tx.BroadcastMode_BROADCAST_MODE_SYNC
		opts.TxOpts = &gnfdsdk.TxOption{Mode: &broadcastMode}
	}
	msgs := []sdk.Msg{createBucketMsg}

	if opts.Tags != nil {
		// Set tag
		grn := gnfdTypes.NewBucketGRN(bucketName)
		msgSetTag := storageTypes.NewMsgSetTag(c.MustGetDefaultAccount().GetAddress(), grn.String(), opts.Tags)
		msgs = append(msgs, msgSetTag)
	}
	resp, err := c.BroadcastTx(ctx, msgs, opts.TxOpts)
	if err != nil {
		return "", err
	}
	txnHash := resp.TxResponse.TxHash
	if !opts.IsAsyncMode {
		ctxTimeout, cancel := context.WithTimeout(ctx, types.ContextTimeout)
		defer cancel()
		txnResponse, err := c.WaitForTx(ctxTimeout, txnHash)
		if err != nil {
			return txnHash, fmt.Errorf("the transaction has been submitted, please check it later:%v", err)
		}
		if txnResponse.TxResult.Code != 0 {
			return txnHash, fmt.Errorf("the createBucket txn has failed with response code: %d, codespace:%s", txnResponse.TxResult.Code, txnResponse.TxResult.Codespace)
		}
	}
	return txnHash, nil
}

// ListBuckets - Lists the bucket info of the user.
//
// If the opts.Account is not set, the user is default set as the sender.
//
// - ctx: Context variables for the current API call.
//
// - opts: The options to set the meta to list the bucket
//
// - ret1: The result of list bucket under specific user address
//
// - ret2: Return error when the request failed, otherwise return nil.
func (c *Client) listBucketsEvm(ctx context.Context, opts types.ListBucketsOptions) (types.ListBucketsResult, error) {
	params := url.Values{}
	params.Set("include-removed", strconv.FormatBool(opts.ShowRemovedBucket))

	account := opts.Account
	if account == "" {
		acc, err := c.GetDefaultAccount()
		if err != nil {
			log.Error().Msg(fmt.Sprintf("failed to get default account:  %s", err.Error()))
			return types.ListBucketsResult{}, err
		}
		account = acc.GetAddress().String()
	} else {
		_, err := sdk.AccAddressFromHexUnsafe(account)
		if err != nil {
			return types.ListBucketsResult{}, err
		}
	}

	reqMeta := requestMeta{
		urlValues:     params,
		contentSHA256: types.EmptyStringSHA256,
		userAddress:   account,
	}

	sendOpt := sendOptions{
		method:           http.MethodGet,
		disableCloseBody: true,
	}

	endpoint, err := c.getEndpointByOpt(&types.EndPointOptions{
		Endpoint:  opts.Endpoint,
		SPAddress: opts.SPAddress,
	})
	if err != nil {
		log.Error().Msg(fmt.Sprintf("get endpoint by option failed %s", err.Error()))
		return types.ListBucketsResult{}, err
	}

	resp, err := c.sendReq(ctx, reqMeta, &sendOpt, endpoint)
	if err != nil {
		log.Error().Msg("the list of user's buckets failed: " + err.Error())
		return types.ListBucketsResult{}, err
	}
	defer utils.CloseResponse(resp)

	listBucketsResult := types.ListBucketsResult{}
	// unmarshal the json content from response body
	buf := new(strings.Builder)
	_, err = io.Copy(buf, resp.Body)
	if err != nil {
		log.Error().Msg("the list of user's buckets failed: " + err.Error())
		return types.ListBucketsResult{}, err
	}

	bufStr := buf.String()
	err = xml.Unmarshal([]byte(bufStr), &listBucketsResult)
	// TODO(annie) remove tolerance for unmarshal err after structs got stabilized
	if err != nil {
		return types.ListBucketsResult{}, err
	}

	return listBucketsResult, nil
}
