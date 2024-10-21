package e2e

import (
	"encoding/hex"
	"testing"
	"time"

	"cosmossdk.io/math"
	"github.com/cometbft/cometbft/crypto/tmhash"
	govTypesV1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
	types2 "github.com/evmos/evmos/v12/sdk/types"
	spTypes "github.com/evmos/evmos/v12/x/sp/types"
	types3 "github.com/evmos/evmos/v12/x/sp/types"
	"github.com/stretchr/testify/suite"
	"github.com/zkMeLabs/mechain-go-sdk/e2e/basesuite"
	"github.com/zkMeLabs/mechain-go-sdk/types"
)

type SPTestSuite struct {
	basesuite.BaseSuite
	OperatorAcc    *types.Account
	FundingAcc     *types.Account
	SealAcc        *types.Account
	ApprovalAcc    *types.Account
	GcAcc          *types.Account
	MaintenanceAcc *types.Account
	BlsAcc         *types.Account
}

func (s *SPTestSuite) SetupSuite() {
	s.BaseSuite.SetupSuite()
	var err error
	s.OperatorAcc, _, err = types.NewAccount("operator")
	s.Require().NoError(err)
	s.FundingAcc, _, err = types.NewAccount("funding")
	s.Require().NoError(err)
	s.SealAcc, _, err = types.NewAccount("seal")
	s.Require().NoError(err)
	s.ApprovalAcc, _, err = types.NewAccount("approval")
	s.Require().NoError(err)
	s.GcAcc, _, err = types.NewAccount("gc")
	s.Require().NoError(err)
	s.MaintenanceAcc, _, err = types.NewAccount("maintenance")
	s.Require().NoError(err)
	s.BlsAcc, _, err = types.NewBlsAccount("bls")
	s.Require().NoError(err)
	s.T().Logf("FundingAddr: %s, sealAddr: %s, approvalAddr: %s, operatorAddr: %s, gcAddr: %s, maintenanceAddr: %s, blsPubKey: %s",
		s.FundingAcc.GetAddress().String(),
		s.SealAcc.GetAddress().String(),
		s.ApprovalAcc.GetAddress().String(),
		s.OperatorAcc.GetAddress().String(),
		s.GcAcc.GetAddress().String(),
		s.MaintenanceAcc.GetAddress().String(),
		s.BlsAcc.GetKeyManager().PubKey().String(),
	)
}

func (s *SPTestSuite) Test_CreateStorageProvider() {
	txHash, err := s.Client.Transfer(s.ClientContext, s.FundingAcc.GetAddress().String(), math.NewIntWithDecimal(10001, types2.DecimalZKME), types2.TxOption{})
	s.Require().NoError(err)
	_, err = s.Client.WaitForTx(s.ClientContext, txHash)
	s.Require().NoError(err)
	fundingBalance, err := s.Client.GetAccountBalance(s.ClientContext, s.FundingAcc.GetAddress().String())
	s.Require().NoError(err)
	s.T().Logf("funding validatorAccount balance: %s", fundingBalance.String())

	txHash, err = s.Client.Transfer(s.ClientContext, s.SealAcc.GetAddress().String(), math.NewIntWithDecimal(1, types2.DecimalZKME), types2.TxOption{})
	s.Require().NoError(err)
	_, err = s.Client.WaitForTx(s.ClientContext, txHash)
	s.Require().NoError(err)
	sealBalance, err := s.Client.GetAccountBalance(s.ClientContext, s.SealAcc.GetAddress().String())
	s.Require().NoError(err)
	s.T().Logf("seal validatorAccount balance: %s", sealBalance.String())

	txHash, err = s.Client.Transfer(s.ClientContext, s.OperatorAcc.GetAddress().String(), math.NewIntWithDecimal(1000, types2.DecimalZKME), types2.TxOption{})
	s.Require().NoError(err)
	_, err = s.Client.WaitForTx(s.ClientContext, txHash)
	s.Require().NoError(err)
	operatorBalance, err := s.Client.GetAccountBalance(s.ClientContext, s.OperatorAcc.GetAddress().String())
	s.Require().NoError(err)
	s.T().Logf("operator validatorAccount balance: %s", operatorBalance.String())

	s.Client.SetDefaultAccount(s.FundingAcc)
	txHash, err = s.Client.GrantDepositForStorageProvider(s.ClientContext, s.OperatorAcc.GetAddress().String(), math.NewIntWithDecimal(10000, types2.DecimalZKME), types.GrantDepositForStorageProviderOptions{})
	s.Require().NoError(err)
	_, err = s.Client.WaitForTx(s.ClientContext, txHash)
	s.Require().NoError(err)

	s.Client.SetDefaultAccount(s.OperatorAcc)

	blsProofBz, err := s.BlsAcc.GetKeyManager().Sign(tmhash.Sum(s.BlsAcc.GetKeyManager().PubKey().Bytes()))
	s.Require().NoError(err)
	proposalID, txHash, err := s.Client.CreateStorageProvider(s.ClientContext, s.FundingAcc.GetAddress().String(), s.SealAcc.GetAddress().String(), s.ApprovalAcc.GetAddress().String(), s.GcAcc.GetAddress().String(), s.MaintenanceAcc.GetAddress().String(),
		hex.EncodeToString(s.BlsAcc.GetKeyManager().PubKey().Bytes()), hex.EncodeToString(blsProofBz),
		"https://sp0.mechain.io",
		math.NewIntWithDecimal(10000, types2.DecimalZKME),
		types3.Description{Moniker: "test"},
		types.CreateStorageProviderOptions{ProposalMetaData: "create", ProposalTitle: "test", ProposalSummary: "test"})
	s.Require().NoError(err)

	createTx, err := s.Client.WaitForTx(s.ClientContext, txHash)
	s.Require().NoError(err)
	s.T().Log(createTx.TxResult.String())

	for {
		p, err := s.Client.GetProposal(s.ClientContext, proposalID)
		s.T().Logf("Proposal: %d, %s, %s, %s", p.Id, p.Status, p.VotingEndTime.String(), p.FinalTallyResult.String())
		s.Require().NoError(err)
		if p.Status == govTypesV1.ProposalStatus_PROPOSAL_STATUS_VOTING_PERIOD {
			break
		}
		time.Sleep(1 * time.Second)
	}

	s.Client.SetDefaultAccount(s.DefaultAccount)
	voteTxHash, err := s.Client.VoteProposal(s.ClientContext, proposalID, govTypesV1.OptionYes, types.VoteProposalOptions{})
	s.Require().NoError(err)

	tx, err := s.Client.WaitForTx(s.ClientContext, voteTxHash)
	s.Require().NoError(err)
	s.T().Logf("VoteTx: %s", hex.EncodeToString(tx.Hash))

	for {
		p, err := s.Client.GetProposal(s.ClientContext, proposalID)
		s.T().Logf("Proposal: %d, %s, %s, %s", p.Id, p.Status, p.VotingEndTime.String(), p.FinalTallyResult.String())
		s.Require().NoError(err)
		if p.Status == govTypesV1.ProposalStatus_PROPOSAL_STATUS_PASSED {
			break
		} else if p.Status == govTypesV1.ProposalStatus_PROPOSAL_STATUS_FAILED {
			s.Require().True(false)
		}
		time.Sleep(1 * time.Second)
	}

	info, err := s.Client.GetStorageProviderInfo(s.ClientContext, s.OperatorAcc.GetAddress())
	s.Require().NoError(err)
	s.Require().Equal(info.Status, spTypes.STATUS_IN_MAINTENANCE)

	// sp activate itself
	s.Client.SetDefaultAccount(s.OperatorAcc)

	updateStatusTxHash, err := s.Client.UpdateSpStatus(s.ClientContext,
		s.OperatorAcc.GetAddress().String(),
		spTypes.STATUS_IN_SERVICE,
		0,
		types2.TxOption{},
	)
	s.Require().NoError(err)
	_, err = s.Client.WaitForTx(s.ClientContext, updateStatusTxHash)
	s.Require().NoError(err)

	info, err = s.Client.GetStorageProviderInfo(s.ClientContext, s.OperatorAcc.GetAddress())
	s.Require().NoError(err)
	s.Require().Equal(info.Status, spTypes.STATUS_IN_SERVICE)
}

func TestSPTestSuite(t *testing.T) {
	suite.Run(t, new(SPTestSuite))
}
