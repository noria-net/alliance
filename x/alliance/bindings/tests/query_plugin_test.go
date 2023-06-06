package bindings_test

import (
	"encoding/json"
	"testing"
	"time"

	querytypes "github.com/cosmos/cosmos-sdk/types/query"

	"cosmossdk.io/math"
	minttypes "github.com/cosmos/cosmos-sdk/x/mint/types"

	tmproto "github.com/cometbft/cometbft/proto/tendermint/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	"github.com/noria-net/alliance/app"
	"github.com/noria-net/alliance/x/alliance/bindings"
	bindingtypes "github.com/noria-net/alliance/x/alliance/bindings/types"
	"github.com/noria-net/alliance/x/alliance/types"
)

func createTestContext(t *testing.T) (*app.App, sdk.Context, time.Time) {
	app := app.Setup(t)
	ctx := app.BaseApp.NewContext(false, tmproto.Header{})
	genesisTime := ctx.BlockTime()
	newAsset := types.NewAllianceAsset(AllianceDenom, sdk.NewDec(2), sdk.NewDec(1), sdk.ZeroDec(), sdk.NewDec(5), sdk.NewDec(0), genesisTime)
	app.AllianceKeeper.InitGenesis(ctx, &types.GenesisState{
		Params: types.DefaultParams(),
		Assets: []types.AllianceAsset{newAsset},
	})
	return app, ctx, genesisTime
}

var AllianceDenom = "alliance"

func TestParamsQuery(t *testing.T) {
	app, ctx, _ := createTestContext(t)

	querier := bindings.NewMockCustomQueryHandler(&app.AllianceKeeper)

	query := bindingtypes.AllianceQuery{
		Alliance: &bindingtypes.AllianceSubQuery{
			Params: &types.QueryParamsRequest{},
		},
	}

	qBz, err := json.Marshal(query)
	require.NoError(t, err)
	rBz, err := querier.TestQuery(ctx, qBz)
	require.NoError(t, err)

	var res types.QueryParamsResponse
	err = json.Unmarshal(rBz, &res)
	require.NoError(t, err)

	values := types.DefaultParams()
	expected := types.QueryParamsResponse{
		Params: types.Params{
			LastTakeRateClaimTime: values.LastTakeRateClaimTime,
			RewardDelayTime:       values.RewardDelayTime,
			TakeRateClaimInterval: values.TakeRateClaimInterval,
		},
	}

	require.Equal(t, expected, res)
}

func TestAssetQuery(t *testing.T) {
	app, ctx, genesisTime := createTestContext(t)

	querier := bindings.NewMockCustomQueryHandler(&app.AllianceKeeper)

	query := bindingtypes.AllianceQuery{
		Alliance: &bindingtypes.AllianceSubQuery{
			Alliance: &types.QueryAllianceRequest{
				Denom: AllianceDenom,
			},
		},
	}

	qBz, err := json.Marshal(query)
	require.NoError(t, err)
	rBz, err := querier.TestQuery(ctx, qBz)
	require.NoError(t, err)

	var res types.QueryAllianceResponse
	err = json.Unmarshal(rBz, &res)
	require.NoError(t, err)

	require.Equal(t, types.QueryAllianceResponse{
		Alliance: &types.AllianceAsset{
			Denom:                AllianceDenom,
			RewardWeight:         sdk.NewDec(2),
			ConsensusWeight:      sdk.NewDec(1),
			TakeRate:             sdk.NewDec(0),
			TotalTokens:          math.NewInt(0),
			TotalValidatorShares: sdk.NewDec(0),
			RewardStartTime:      genesisTime,
			RewardChangeRate:     sdk.MustNewDecFromStr("1"),
			RewardChangeInterval: 0,
			LastRewardChangeTime: genesisTime,
			RewardWeightRange:    types.RewardWeightRange{Min: sdk.NewDec(0), Max: sdk.NewDec(5)},
			IsInitialized:        false,
		},
	}, res)
}

func TestDelegationQuery(t *testing.T) {
	app, ctx, _ := createTestContext(t)

	querier := bindings.NewMockCustomQueryHandler(&app.AllianceKeeper)

	delegations := app.StakingKeeper.GetAllDelegations(ctx)
	require.Len(t, delegations, 1)
	// All the addresses needed
	delAddr, err := sdk.AccAddressFromBech32(delegations[0].DelegatorAddress)
	require.NoError(t, err)
	valAddr, err := sdk.ValAddressFromBech32(delegations[0].ValidatorAddress)
	require.NoError(t, err)
	val, err := app.AllianceKeeper.GetAllianceValidator(ctx, valAddr)
	require.NoError(t, err)

	// Mint alliance tokens
	err = app.BankKeeper.MintCoins(ctx, minttypes.ModuleName, sdk.NewCoins(sdk.NewCoin(AllianceDenom, sdk.NewInt(2000_000))))
	require.NoError(t, err)
	err = app.BankKeeper.SendCoinsFromModuleToAccount(ctx, minttypes.ModuleName, delAddr, sdk.NewCoins(sdk.NewCoin(AllianceDenom, sdk.NewInt(2000_000))))
	require.NoError(t, err)

	// Check current total staked tokens
	totalBonded := app.StakingKeeper.TotalBondedTokens(ctx)
	require.Equal(t, sdk.NewInt(1000_000), totalBonded)

	// Delegate
	_, err = app.AllianceKeeper.Delegate(ctx, delAddr, val, sdk.NewCoin(AllianceDenom, sdk.NewInt(1000_000)))
	require.NoError(t, err)

	delegationQuery := bindingtypes.AllianceQuery{
		Alliance: &bindingtypes.AllianceSubQuery{
			Delegation: &types.QueryAllianceDelegationRequest{
				DelegatorAddr: delAddr.String(),
				ValidatorAddr: val.GetOperator().String(),
				Denom:         AllianceDenom,
			},
		},
	}

	qBz, err := json.Marshal(delegationQuery)
	require.NoError(t, err)
	rBz, err := querier.TestQuery(ctx, qBz)
	require.NoError(t, err)

	var delegationResponse types.QueryAllianceDelegationResponse
	err = json.Unmarshal(rBz, &delegationResponse)
	require.NoError(t, err)

	require.Equal(t, types.QueryAllianceDelegationResponse{
		Delegation: types.DelegationResponse{
			Delegation: types.Delegation{
				DelegatorAddress:      delAddr.String(),
				ValidatorAddress:      valAddr.String(),
				Denom:                 AllianceDenom,
				Shares:                sdk.NewDec(1000000),
				RewardHistory:         nil,
				LastRewardClaimHeight: 0,
			},
			Balance: sdk.NewCoin(AllianceDenom, sdk.NewInt(1000_000)),
		},
	}, delegationResponse)
}

func TestDelegationRewardsQuery(t *testing.T) {
	app, ctx, _ := createTestContext(t)

	querier := bindings.NewMockCustomQueryHandler(&app.AllianceKeeper)

	delegations := app.StakingKeeper.GetAllDelegations(ctx)
	require.Len(t, delegations, 1)
	// All the addresses needed
	delAddr, err := sdk.AccAddressFromBech32(delegations[0].DelegatorAddress)
	require.NoError(t, err)
	valAddr, err := sdk.ValAddressFromBech32(delegations[0].ValidatorAddress)
	require.NoError(t, err)
	val, err := app.AllianceKeeper.GetAllianceValidator(ctx, valAddr)
	require.NoError(t, err)

	// Mint alliance tokens
	err = app.BankKeeper.MintCoins(ctx, minttypes.ModuleName, sdk.NewCoins(sdk.NewCoin(AllianceDenom, sdk.NewInt(2000_000))))
	require.NoError(t, err)
	err = app.BankKeeper.SendCoinsFromModuleToAccount(ctx, minttypes.ModuleName, delAddr, sdk.NewCoins(sdk.NewCoin(AllianceDenom, sdk.NewInt(2000_000))))
	require.NoError(t, err)

	// Check current total staked tokens
	totalBonded := app.StakingKeeper.TotalBondedTokens(ctx)
	require.Equal(t, sdk.NewInt(1000_000), totalBonded)

	// Delegate
	_, err = app.AllianceKeeper.Delegate(ctx, delAddr, val, sdk.NewCoin(AllianceDenom, sdk.NewInt(1000_000)))
	require.NoError(t, err)

	assets := app.AllianceKeeper.GetAllAssets(ctx)
	_, err = app.AllianceKeeper.RebalanceBondTokenWeights(ctx, assets)
	require.NoError(t, err)

	// Transfer to reward pool
	mintPoolAddr := app.AccountKeeper.GetModuleAddress(minttypes.ModuleName)
	err = app.BankKeeper.MintCoins(ctx, minttypes.ModuleName, sdk.NewCoins(sdk.NewCoin("stake", sdk.NewInt(4000_000))))
	require.NoError(t, err)
	err = app.AllianceKeeper.AddAssetsToRewardPool(ctx, mintPoolAddr, val, sdk.NewCoins(sdk.NewCoin("stake", sdk.NewInt(2000_000))))
	require.NoError(t, err)

	delegationQuery := bindingtypes.AllianceQuery{
		Alliance: &bindingtypes.AllianceSubQuery{
			DelegationRewards: &types.QueryAllianceDelegationRewardsRequest{
				DelegatorAddr: delAddr.String(),
				ValidatorAddr: val.GetOperator().String(),
				Denom:         AllianceDenom,
			},
		},
	}
	qBz, err := json.Marshal(delegationQuery)
	require.NoError(t, err)
	rBz, err := querier.TestQuery(ctx, qBz)
	require.NoError(t, err)

	var response types.QueryAllianceDelegationRewardsResponse
	err = json.Unmarshal(rBz, &response)
	require.NoError(t, err)

	require.Equal(t, types.QueryAllianceDelegationRewardsResponse{
		Rewards: []sdk.Coin{
			{
				Denom:  "stake",
				Amount: math.NewInt(2000000),
			},
		},
	}, response)
}

func TestAllAlliancesQuery(t *testing.T) {
	app, ctx, genesisTime := createTestContext(t)

	querier := bindings.NewMockCustomQueryHandler(&app.AllianceKeeper)

	query := bindingtypes.AllianceQuery{
		Alliance: &bindingtypes.AllianceSubQuery{
			Alliances: &types.QueryAlliancesRequest{
				Pagination: &querytypes.PageRequest{
					Offset:     0,
					Limit:      1,
					CountTotal: true,
				},
			},
		},
	}
	qBz, err := json.Marshal(query)
	require.NoError(t, err)
	rBz, err := querier.TestQuery(ctx, qBz)
	require.NoError(t, err)

	var response types.QueryAlliancesResponse
	err = json.Unmarshal(rBz, &response)
	require.NoError(t, err)

	require.Equal(t, types.QueryAlliancesResponse{
		Alliances: []types.AllianceAsset{
			{
				Denom:                AllianceDenom,
				RewardWeight:         sdk.NewDec(2),
				ConsensusWeight:      sdk.NewDec(1),
				TakeRate:             sdk.NewDec(0),
				TotalTokens:          math.NewInt(0),
				TotalValidatorShares: sdk.NewDec(0),
				RewardStartTime:      genesisTime,
				RewardChangeRate:     sdk.MustNewDecFromStr("1"),
				RewardChangeInterval: 0,
				LastRewardChangeTime: genesisTime,
				RewardWeightRange:    types.RewardWeightRange{Min: sdk.NewDec(0), Max: sdk.NewDec(5)},
				IsInitialized:        false,
			},
		},
		Pagination: &querytypes.PageResponse{
			Total: 1,
		},
	}, response)
}

func TestAllAllianceValidators(t *testing.T) {
	app, ctx, _ := createTestContext(t)

	querier := bindings.NewMockCustomQueryHandler(&app.AllianceKeeper)

	delegations := app.StakingKeeper.GetAllDelegations(ctx)
	require.Len(t, delegations, 1)
	// All the addresses needed
	delAddr, err := sdk.AccAddressFromBech32(delegations[0].DelegatorAddress)
	require.NoError(t, err)
	valAddr, err := sdk.ValAddressFromBech32(delegations[0].ValidatorAddress)
	require.NoError(t, err)
	val, err := app.AllianceKeeper.GetAllianceValidator(ctx, valAddr)
	require.NoError(t, err)

	// Mint alliance tokens
	err = app.BankKeeper.MintCoins(ctx, minttypes.ModuleName, sdk.NewCoins(sdk.NewCoin(AllianceDenom, sdk.NewInt(2000_000))))
	require.NoError(t, err)
	err = app.BankKeeper.SendCoinsFromModuleToAccount(ctx, minttypes.ModuleName, delAddr, sdk.NewCoins(sdk.NewCoin(AllianceDenom, sdk.NewInt(2000_000))))
	require.NoError(t, err)

	// Check current total staked tokens
	totalBonded := app.StakingKeeper.TotalBondedTokens(ctx)
	require.Equal(t, sdk.NewInt(1000_000), totalBonded)

	// Delegate
	_, err = app.AllianceKeeper.Delegate(ctx, delAddr, val, sdk.NewCoin(AllianceDenom, sdk.NewInt(1000_000)))
	require.NoError(t, err)

	query := bindingtypes.AllianceQuery{
		Alliance: &bindingtypes.AllianceSubQuery{
			Validators: &types.QueryAllAllianceValidatorsRequest{
				Pagination: &querytypes.PageRequest{
					Offset:     0,
					Limit:      1,
					CountTotal: true,
				},
			},
		},
	}

	qBz, err := json.Marshal(query)
	require.NoError(t, err)
	rBz, err := querier.TestQuery(ctx, qBz)
	require.NoError(t, err)

	var res types.QueryAllianceValidatorsResponse
	err = json.Unmarshal(rBz, &res)
	require.NoError(t, err)

	require.Equal(t, types.QueryAllianceValidatorsResponse{
		Validators: []types.QueryAllianceValidatorResponse{
			{
				ValidatorAddr:         valAddr.String(),
				TotalDelegationShares: sdk.NewDecCoins(sdk.NewDecCoin(AllianceDenom, sdk.NewInt(1000_000))),
				ValidatorShares:       sdk.NewDecCoins(sdk.NewDecCoin(AllianceDenom, sdk.NewInt(1000_000))),
				TotalStaked:           sdk.NewDecCoins(sdk.NewDecCoin(AllianceDenom, sdk.NewInt(1000_000))),
			},
		},
		Pagination: &querytypes.PageResponse{
			Total: 1,
		},
	}, res)
}

func TestAllianceValidator(t *testing.T) {
	app, ctx, _ := createTestContext(t)

	querier := bindings.NewMockCustomQueryHandler(&app.AllianceKeeper)

	delegations := app.StakingKeeper.GetAllDelegations(ctx)
	require.Len(t, delegations, 1)
	// All the addresses needed
	delAddr, err := sdk.AccAddressFromBech32(delegations[0].DelegatorAddress)
	require.NoError(t, err)
	valAddr, err := sdk.ValAddressFromBech32(delegations[0].ValidatorAddress)
	require.NoError(t, err)
	val, err := app.AllianceKeeper.GetAllianceValidator(ctx, valAddr)
	require.NoError(t, err)

	// Mint alliance tokens
	err = app.BankKeeper.MintCoins(ctx, minttypes.ModuleName, sdk.NewCoins(sdk.NewCoin(AllianceDenom, sdk.NewInt(2000_000))))
	require.NoError(t, err)
	err = app.BankKeeper.SendCoinsFromModuleToAccount(ctx, minttypes.ModuleName, delAddr, sdk.NewCoins(sdk.NewCoin(AllianceDenom, sdk.NewInt(2000_000))))
	require.NoError(t, err)

	// Check current total staked tokens
	totalBonded := app.StakingKeeper.TotalBondedTokens(ctx)
	require.Equal(t, sdk.NewInt(1000_000), totalBonded)

	// Delegate
	_, err = app.AllianceKeeper.Delegate(ctx, delAddr, val, sdk.NewCoin(AllianceDenom, sdk.NewInt(1000_000)))
	require.NoError(t, err)

	query := bindingtypes.AllianceQuery{
		Alliance: &bindingtypes.AllianceSubQuery{
			Validator: &types.QueryAllianceValidatorRequest{
				ValidatorAddr: valAddr.String(),
			},
		},
	}

	qBz, err := json.Marshal(query)
	require.NoError(t, err)
	rBz, err := querier.TestQuery(ctx, qBz)
	require.NoError(t, err)

	var res types.QueryAllianceValidatorResponse
	err = json.Unmarshal(rBz, &res)
	require.NoError(t, err)

	require.Equal(t, types.QueryAllianceValidatorResponse{
		ValidatorAddr:         valAddr.String(),
		TotalDelegationShares: sdk.NewDecCoins(sdk.NewDecCoin(AllianceDenom, sdk.NewInt(1000_000))),
		ValidatorShares:       sdk.NewDecCoins(sdk.NewDecCoin(AllianceDenom, sdk.NewInt(1000_000))),
		TotalStaked:           sdk.NewDecCoins(sdk.NewDecCoin(AllianceDenom, sdk.NewInt(1000_000))),
	}, res)
}

func TestAlliancesDelegations(t *testing.T) {
	app, ctx, _ := createTestContext(t)

	querier := bindings.NewMockCustomQueryHandler(&app.AllianceKeeper)

	delegations := app.StakingKeeper.GetAllDelegations(ctx)
	require.Len(t, delegations, 1)
	// All the addresses needed
	delAddr, err := sdk.AccAddressFromBech32(delegations[0].DelegatorAddress)
	require.NoError(t, err)
	valAddr, err := sdk.ValAddressFromBech32(delegations[0].ValidatorAddress)
	require.NoError(t, err)
	val, err := app.AllianceKeeper.GetAllianceValidator(ctx, valAddr)
	require.NoError(t, err)

	// Mint alliance tokens
	err = app.BankKeeper.MintCoins(ctx, minttypes.ModuleName, sdk.NewCoins(sdk.NewCoin(AllianceDenom, sdk.NewInt(2000_000))))
	require.NoError(t, err)
	err = app.BankKeeper.SendCoinsFromModuleToAccount(ctx, minttypes.ModuleName, delAddr, sdk.NewCoins(sdk.NewCoin(AllianceDenom, sdk.NewInt(2000_000))))
	require.NoError(t, err)

	// Check current total staked tokens
	totalBonded := app.StakingKeeper.TotalBondedTokens(ctx)
	require.Equal(t, sdk.NewInt(1000_000), totalBonded)

	// Delegate
	_, err = app.AllianceKeeper.Delegate(ctx, delAddr, val, sdk.NewCoin(AllianceDenom, sdk.NewInt(1000_000)))
	require.NoError(t, err)

	query := bindingtypes.AllianceQuery{
		Alliance: &bindingtypes.AllianceSubQuery{
			AlliancesDelegations: &types.QueryAllAlliancesDelegationsRequest{
				Pagination: &querytypes.PageRequest{
					Offset:     0,
					Limit:      1,
					CountTotal: true,
				},
			},
		},
	}

	qBz, err := json.Marshal(query)
	require.NoError(t, err)
	rBz, err := querier.TestQuery(ctx, qBz)
	require.NoError(t, err)

	var res types.QueryAlliancesDelegationsResponse
	err = json.Unmarshal(rBz, &res)
	require.NoError(t, err)

	require.Equal(t, types.QueryAlliancesDelegationsResponse{
		Delegations: []types.DelegationResponse{
			{
				Delegation: types.Delegation{
					DelegatorAddress:      delAddr.String(),
					ValidatorAddress:      val.OperatorAddress,
					Denom:                 AllianceDenom,
					Shares:                sdk.NewDec(1000000),
					RewardHistory:         nil,
					LastRewardClaimHeight: 0,
				},
				Balance: sdk.NewCoin(AllianceDenom, math.NewInt(1000_000)),
			},
		},
		Pagination: &querytypes.PageResponse{
			Total: 1,
		},
	}, res)
}
