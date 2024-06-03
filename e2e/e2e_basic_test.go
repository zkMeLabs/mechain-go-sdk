package e2e

import (
	"encoding/hex"
	"fmt"
	"testing"

	"github.com/stretchr/testify/suite"

	"cosmossdk.io/math"
	"github.com/bnb-chain/greenfield-go-sdk/e2e/basesuite"
	"github.com/bnb-chain/greenfield-go-sdk/types"
	types2 "github.com/evmos/evmos/v12/sdk/types"
)

type BasicTestSuite struct {
	basesuite.BaseSuite
}

func (s *BasicTestSuite) SetupSuite() {
	s.BaseSuite.SetupSuite()
}

func (s *BasicTestSuite) Test_Basic() {
	_, _, err := s.Client.GetNodeInfo(s.ClientContext)
	s.Require().NoError(err)

	latestBlock, err := s.Client.GetLatestBlock(s.ClientContext)
	s.Require().NoError(err)
	fmt.Println(latestBlock.String())

	heightBefore := latestBlock.Header.Height
	err = s.Client.WaitForBlockHeight(s.ClientContext, heightBefore+10)
	s.Require().NoError(err)
	height, err := s.Client.GetLatestBlockHeight(s.ClientContext)
	s.Require().NoError(err)
	s.Require().GreaterOrEqual(height, heightBefore+10)

	syncing, err := s.Client.GetSyncing(s.ClientContext)
	s.Require().NoError(err)
	s.Require().False(syncing)

	blockByHeight, err := s.Client.GetBlockByHeight(s.ClientContext, heightBefore)
	s.Require().NoError(err)
	s.Require().Equal(blockByHeight.Header.Hash(), latestBlock.Header.Hash())
}

func (s *BasicTestSuite) Test_Account() {
	balance, err := s.Client.GetAccountBalance(s.ClientContext, s.DefaultAccount.GetAddress().String())
	s.Require().NoError(err)
	s.T().Logf("Balance: %s", balance.String())

	account1, _, err := types.NewAccount("test2")
	s.Require().NoError(err)
	transferTxHash, err := s.Client.Transfer(s.ClientContext, account1.GetAddress().String(), math.NewIntFromUint64(1), types2.TxOption{})
	s.Require().NoError(err)
	s.T().Logf("Transfer response: %s", transferTxHash)

	waitForTx, err := s.Client.WaitForTx(s.ClientContext, transferTxHash)
	s.Require().NoError(err)
	s.T().Logf("Wair for tx: %s", hex.EncodeToString(waitForTx.Hash))

	balance, err = s.Client.GetAccountBalance(s.ClientContext, account1.GetAddress().String())
	s.Require().NoError(err)
	s.T().Logf("Balance: %s", balance.String())
	s.Require().True(balance.Amount.Equal(math.NewInt(1)))

	acc, err := s.Client.GetAccount(s.ClientContext, account1.GetAddress().String())
	s.Require().NoError(err)
	s.T().Logf("Acc: %s", acc.String())
	s.Require().Equal(acc.GetAddress(), account1.GetAddress())
	s.Require().Equal(acc.GetSequence(), uint64(0))

	paymentAccountsByOwnerBefore, err := s.Client.GetPaymentAccountsByOwner(s.ClientContext, s.DefaultAccount.GetAddress().String())
	numPaymentAccounts := 0
	if err == nil {
		numPaymentAccounts = len(paymentAccountsByOwnerBefore)
	}

	txHash, err := s.Client.CreatePaymentAccount(s.ClientContext, s.DefaultAccount.GetAddress().String(), types2.TxOption{})
	s.Require().NoError(err)
	s.T().Logf("Acc: %s", txHash)
	waitForTx, err = s.Client.WaitForTx(s.ClientContext, txHash)
	s.Require().NoError(err)
	s.T().Logf("Wair for tx: %s", hex.EncodeToString(waitForTx.Hash))

	paymentAccountsByOwner, err := s.Client.GetPaymentAccountsByOwner(s.ClientContext, s.DefaultAccount.GetAddress().String())
	s.Require().NoError(err)
	s.Require().Equal(len(paymentAccountsByOwner), numPaymentAccounts+1)
}

func (s *BasicTestSuite) Test_MultiTransfer() {
	transferDetails := make([]types.TransferDetail, 0)
	totalSendAmount := math.NewInt(0)

	receiver1, _, err := types.NewAccount("receiver1")
	s.Require().NoError(err)
	receiver1Amount := math.NewInt(1000)
	totalSendAmount = totalSendAmount.Add(receiver1Amount)
	transferDetails = append(transferDetails, types.TransferDetail{
		ToAddress: receiver1.GetAddress().String(),
		Amount:    receiver1Amount,
	})

	receiver2, _, err := types.NewAccount("receiver2")
	s.Require().NoError(err)
	receiver2Amount := math.NewInt(1000)
	totalSendAmount = totalSendAmount.Add(receiver2Amount)
	transferDetails = append(transferDetails, types.TransferDetail{
		ToAddress: receiver2.GetAddress().String(),
		Amount:    receiver2Amount,
	})

	receiver3, _, err := types.NewAccount("receiver3")
	s.Require().NoError(err)
	receiver3Amount := math.NewInt(1000)
	totalSendAmount = totalSendAmount.Add(receiver3Amount)
	transferDetails = append(transferDetails, types.TransferDetail{
		ToAddress: receiver3.GetAddress().String(),
		Amount:    receiver3Amount,
	})

	s.T().Logf("totally sending %s", totalSendAmount.String())

	txHash, err := s.Client.MultiTransfer(s.ClientContext, transferDetails, types2.TxOption{})
	s.Require().NoError(err)
	s.T().Log(txHash)

	_, err = s.Client.WaitForTx(s.ClientContext, txHash)
	s.Require().NoError(err)

	balance1, err := s.Client.GetAccountBalance(s.ClientContext, receiver1.GetAddress().String())
	s.Require().NoError(err)
	s.Assertions.Equal(receiver1Amount, balance1.Amount)

	balance2, err := s.Client.GetAccountBalance(s.ClientContext, receiver2.GetAddress().String())
	s.Require().NoError(err)
	s.Assertions.Equal(receiver2Amount, balance2.Amount)

	balance3, err := s.Client.GetAccountBalance(s.ClientContext, receiver3.GetAddress().String())
	s.Require().NoError(err)
	s.Assertions.Equal(receiver3Amount, balance3.Amount)
}

func (s *BasicTestSuite) Test_Payment() {
	account := s.DefaultAccount
	cli := s.Client
	t := s.T()
	ctx := s.ClientContext

	paymentAccountsBeforeCreate, err := cli.GetPaymentAccountsByOwner(ctx, account.GetAddress().String())
	s.Require().NoError(err)
	txHash, err := cli.CreatePaymentAccount(ctx, account.GetAddress().String(), types2.TxOption{})
	s.Require().NoError(err)
	t.Logf("Acc: %s", txHash)
	waitForTx, err := cli.WaitForTx(ctx, txHash)
	s.Require().NoError(err)
	t.Logf("Wair for tx: %s", hex.EncodeToString(waitForTx.Hash))

	paymentAccountsByOwnerAfterCreate, err := cli.GetPaymentAccountsByOwner(ctx, account.GetAddress().String())
	s.Require().NoError(err)
	s.Require().Equal(len(paymentAccountsByOwnerAfterCreate)-len(paymentAccountsBeforeCreate), 1)

	// deposit
	paymentAddr := paymentAccountsByOwnerAfterCreate[len(paymentAccountsByOwnerAfterCreate)-1].Addr
	depositAmount := math.NewIntFromUint64(100)
	depositTxHash, err := cli.Deposit(ctx, paymentAddr, depositAmount, types2.TxOption{})
	s.Require().NoError(err)
	t.Logf("deposit tx: %s", depositTxHash)
	waitForTx, err = cli.WaitForTx(ctx, depositTxHash)
	s.Require().NoError(err)
	t.Logf("Wair for tx: %s", hex.EncodeToString(waitForTx.Hash))

	// get stream record
	streamRecord, err := cli.GetStreamRecord(ctx, paymentAddr)
	s.Require().NoError(err)
	s.Require().Equal(streamRecord.StaticBalance.String(), depositAmount.String())

	// withdraw
	withdrawAmount := math.NewIntFromUint64(50)
	withdrawTxHash, err := cli.Withdraw(ctx, paymentAddr, withdrawAmount, types2.TxOption{})
	s.Require().NoError(err)
	t.Logf("withdraw tx: %s", withdrawTxHash)
	waitForTx, err = cli.WaitForTx(ctx, withdrawTxHash)
	s.Require().NoError(err)
	t.Logf("Wair for tx: %s", hex.EncodeToString(waitForTx.Hash))
	streamRecordAfterWithdraw, err := cli.GetStreamRecord(ctx, paymentAddr)
	s.Require().NoError(err)
	s.Require().Equal(streamRecordAfterWithdraw.StaticBalance.String(), depositAmount.Sub(withdrawAmount).String())

	// disable refund
	paymentAccountBeforeDisableRefund, err := cli.GetPaymentAccount(ctx, paymentAddr)
	s.Require().NoError(err)
	s.Require().True(paymentAccountBeforeDisableRefund.Refundable)
	disableRefundTxHash, err := cli.DisableRefund(ctx, paymentAddr, types2.TxOption{})
	s.Require().NoError(err)
	t.Logf("disable refund tx: %s", disableRefundTxHash)
	waitForTx, err = cli.WaitForTx(ctx, disableRefundTxHash)
	s.Require().NoError(err)
	t.Logf("Wair for tx: %s", hex.EncodeToString(waitForTx.Hash))
	paymentAccountAfterDisableRefund, err := cli.GetPaymentAccount(ctx, paymentAddr)
	s.Require().NoError(err)
	s.Require().False(paymentAccountAfterDisableRefund.Refundable)
}

func TestBasicTestSuite(t *testing.T) {
	suite.Run(t, new(BasicTestSuite))
}
