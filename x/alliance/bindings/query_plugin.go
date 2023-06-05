package bindings

import (
	"encoding/json"

	sdkerrors "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"

	wasmkeeper "github.com/CosmWasm/wasmd/x/wasm/keeper"
	wasmvmtypes "github.com/CosmWasm/wasmvm/types"
	"github.com/noria-net/alliance/x/alliance/bindings/types"
	"github.com/noria-net/alliance/x/alliance/keeper"
	alliancetypes "github.com/noria-net/alliance/x/alliance/types"
)

type QueryPlugin struct {
	allianceKeeper keeper.Keeper
}

func CustomQueryDecorator(allianceKeeper *keeper.Keeper) func(wasmkeeper.WasmVMQueryHandler) wasmkeeper.WasmVMQueryHandler {
	return func(old wasmkeeper.WasmVMQueryHandler) wasmkeeper.WasmVMQueryHandler {
		return &CustomQueryHandler{
			wrapped:        old,
			allianceKeeper: allianceKeeper,
		}
	}
}

type MockQueryHandler struct {
}

func (m *MockQueryHandler) HandleQuery(ctx sdk.Context, caller sdk.AccAddress, request wasmvmtypes.QueryRequest) ([]byte, error) {
	return nil, nil
}

func NewMockCustomQueryHandler(allianceKeeper *keeper.Keeper) *CustomQueryHandler {
	return &CustomQueryHandler{
		wrapped:        &MockQueryHandler{},
		allianceKeeper: allianceKeeper,
	}
}

func (m *CustomQueryHandler) TestQuery(ctx sdk.Context, payload []byte) ([]byte, error) {
	req := wasmvmtypes.QueryRequest{
		Custom: payload,
	}
	addr := sdk.AccAddress("addr1")
	return m.HandleQuery(ctx, addr, req)
}

type CustomQueryHandler struct {
	wrapped        wasmkeeper.WasmVMQueryHandler
	allianceKeeper *keeper.Keeper
}

func NewAllianceQueryPlugin(keeper keeper.Keeper) *QueryPlugin {
	return &QueryPlugin{
		allianceKeeper: keeper,
	}
}

func (m *CustomQueryHandler) HandleQuery(ctx sdk.Context, caller sdk.AccAddress, request wasmvmtypes.QueryRequest) ([]byte, error) {
	if request.Custom == nil {
		return m.wrapped.HandleQuery(ctx, caller, request)
	}
	customQuery := request.Custom

	var allianceQuery types.AllianceQuery
	if err := json.Unmarshal(customQuery, &allianceQuery); err != nil {
		return nil, sdkerrors.Wrap(ErrAllianceMsg, "requires 'alliance' field")
	}
	if allianceQuery.Alliance == nil {
		return m.wrapped.HandleQuery(ctx, caller, request)
	}

	req := allianceQuery.Alliance

	var querier alliancetypes.QueryServer = keeper.QueryServer{
		Keeper: *m.allianceKeeper,
	}

	switch {
	case req.Params != nil:
		res, err := querier.Params(ctx, (*alliancetypes.QueryParamsRequest)(req.Params))
		if err != nil {
			return nil, sdkerrors.Wrap(ErrAllianceMsg, "failed to get alliance params")
		}
		bz, err := json.Marshal(res)
		if err != nil {
			return nil, sdkerrors.Wrap(ErrAllianceMsg, "failed to marshal alliance params")
		}
		return bz, nil
	case req.Alliance != nil:
		res, err := m.GetAlliance(ctx, req.Alliance.Denom)
		if err != nil {
			return nil, sdkerrors.Wrap(ErrAllianceMsg, "failed to get alliance")
		}
		return res, nil
	case req.Delegation != nil:
		res, err := m.GetDelegation(ctx, req.Delegation.Denom, req.Delegation.Delegator, req.Delegation.Validator)
		if err != nil {
			return nil, sdkerrors.Wrap(ErrAllianceMsg, "failed to get delegation")
		}
		return res, nil
	case req.DelegationRewards != nil:
		res, err := m.GetDelegationRewards(ctx, req.DelegationRewards.Denom, req.DelegationRewards.Delegator, req.DelegationRewards.Validator)
		if err != nil {
			return nil, sdkerrors.Wrap(ErrAllianceMsg, "failed to get delegation rewards")
		}
		return res, nil
	default:
		return nil, wasmvmtypes.UnsupportedRequest{Kind: "unknown alliance query variant"}
	}
}

func CustomQuerier(q *QueryPlugin) func(ctx sdk.Context, request json.RawMessage) (result []byte, err error) {
	return func(ctx sdk.Context, request json.RawMessage) (result []byte, err error) {
		var AllianceRequest types.AllianceQuery
		err = json.Unmarshal(request, &AllianceRequest)
		if err != nil {
			return
		}
		// if AllianceRequest.Alliance != nil {
		// 	return q.GetAlliance(ctx, AllianceRequest.Alliance.Denom)
		// }
		// if AllianceRequest.Delegation != nil {
		// 	return q.GetDelegation(ctx, AllianceRequest.Delegation.Denom, AllianceRequest.Delegation.Delegator, AllianceRequest.Delegation.Validator)
		// }
		// if AllianceRequest.DelegationRewards != nil {
		// 	return q.GetDelegationRewards(ctx, AllianceRequest.DelegationRewards.Denom, AllianceRequest.DelegationRewards.Delegator, AllianceRequest.DelegationRewards.Validator)
		// }
		return nil, nil
	}
}

func (q *CustomQueryHandler) GetParams(ctx sdk.Context) (res []byte, err error) {
	params := q.allianceKeeper.Params(ctx)
	res, err = json.Marshal(alliancetypes.QueryParamsResponse{
		Params: alliancetypes.Params{
			LastTakeRateClaimTime: params.LastTakeRateClaimTime,
			RewardDelayTime:       params.RewardDelayTime,
			TakeRateClaimInterval: params.TakeRateClaimInterval,
		},
	})
	return
}

func (q *CustomQueryHandler) GetAlliance(ctx sdk.Context, denom string) (res []byte, err error) {
	asset, found := q.allianceKeeper.GetAssetByDenom(ctx, denom)
	if !found {
		return nil, alliancetypes.ErrUnknownAsset
	}
	res, err = json.Marshal(alliancetypes.QueryAllianceResponse{
		Alliance: &alliancetypes.AllianceAsset{
			Denom:                asset.Denom,
			RewardWeight:         asset.RewardWeight,
			TakeRate:             asset.TakeRate,
			TotalTokens:          asset.TotalTokens,
			TotalValidatorShares: asset.TotalValidatorShares,
			RewardStartTime:      asset.RewardStartTime,
			RewardChangeRate:     asset.RewardChangeRate,
			RewardChangeInterval: 0,
			LastRewardChangeTime: asset.LastRewardChangeTime,
			RewardWeightRange:    alliancetypes.RewardWeightRange{Min: asset.RewardWeightRange.Min, Max: asset.RewardWeightRange.Max},
			IsInitialized:        asset.IsInitialized,
			ConsensusWeight:      asset.ConsensusWeight,
		},
	})
	return
}

func (q *CustomQueryHandler) GetDelegation(ctx sdk.Context, denom string, delegator string, validator string) (res []byte, err error) {
	delegatorAddr, err := sdk.AccAddressFromBech32(delegator)
	if err != nil {
		return
	}
	validatorAddr, err := sdk.ValAddressFromBech32(validator)
	if err != nil {
		return
	}
	delegation, found := q.allianceKeeper.GetDelegation(ctx, delegatorAddr, validatorAddr, denom)
	if !found {
		return nil, alliancetypes.ErrDelegationNotFound
	}
	asset, found := q.allianceKeeper.GetAssetByDenom(ctx, denom)
	if !found {
		return nil, alliancetypes.ErrUnknownAsset
	}

	allianceValidator, err := q.allianceKeeper.GetAllianceValidator(ctx, validatorAddr)
	if err != nil {
		return nil, err
	}
	balance := alliancetypes.GetDelegationTokens(delegation, allianceValidator, asset)
	res, err = json.Marshal(alliancetypes.QueryAllianceDelegationResponse{
		Delegation: alliancetypes.DelegationResponse{
			Delegation: delegation,
			Balance:    balance,
		},
	})
	return res, err
}

func (q *CustomQueryHandler) GetDelegationRewards(ctx sdk.Context, denom string, delegator string, validator string) (res []byte, err error) {
	delegatorAddr, err := sdk.AccAddressFromBech32(delegator)
	if err != nil {
		return
	}
	validatorAddr, err := sdk.ValAddressFromBech32(validator)
	if err != nil {
		return
	}
	delegation, found := q.allianceKeeper.GetDelegation(ctx, delegatorAddr, validatorAddr, denom)
	if !found {
		return nil, alliancetypes.ErrDelegationNotFound
	}
	allianceValidator, err := q.allianceKeeper.GetAllianceValidator(ctx, validatorAddr)
	if err != nil {
		return nil, err
	}
	asset, found := q.allianceKeeper.GetAssetByDenom(ctx, denom)
	if !found {
		return nil, alliancetypes.ErrUnknownAsset
	}

	rewards, _, err := q.allianceKeeper.CalculateDelegationRewards(ctx, delegation, allianceValidator, asset)
	if err != nil {
		return
	}

	var coins []sdk.Coin
	for _, coin := range rewards {
		coins = append(coins, sdk.NewCoin(coin.Denom, coin.Amount))
	}

	res, err = json.Marshal(alliancetypes.QueryAllianceDelegationRewardsResponse{
		Rewards: coins,
	})
	return res, err
}
