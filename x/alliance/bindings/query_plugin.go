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
		res, err := querier.Alliance(ctx, req.Alliance)
		if err != nil {
			return nil, sdkerrors.Wrap(ErrAllianceMsg, "failed to get alliance")
		}
		bz, err := json.Marshal(res)
		if err != nil {
			return nil, sdkerrors.Wrap(ErrAllianceMsg, "failed to marshal alliance")
		}
		return bz, nil
	case req.Delegation != nil:
		res, err := querier.AllianceDelegation(ctx, req.Delegation)
		if err != nil {
			return nil, sdkerrors.Wrap(ErrAllianceMsg, "failed to get delegation")
		}
		bz, err := json.Marshal(res)
		if err != nil {
			return nil, sdkerrors.Wrap(ErrAllianceMsg, "failed to marshal alliance delegations")
		}
		return bz, nil
	case req.DelegationRewards != nil:
		res, err := querier.AllianceDelegationRewards(ctx, req.DelegationRewards)
		if err != nil {
			return nil, sdkerrors.Wrap(ErrAllianceMsg, "failed to get delegation rewards")
		}
		bz, err := json.Marshal(res)
		if err != nil {
			return nil, sdkerrors.Wrap(ErrAllianceMsg, "failed to marshal alliance delegation rewards")
		}
		return bz, nil
	case req.Alliances != nil:
		res, err := querier.Alliances(ctx, req.Alliances)
		if err != nil {
			return nil, sdkerrors.Wrap(ErrAllianceMsg, "failed to get all alliances")
		}
		bz, err := json.Marshal(res)
		if err != nil {
			return nil, sdkerrors.Wrap(ErrAllianceMsg, "failed to marshal all alliances")
		}
		return bz, nil
	case req.Validators != nil:
		res, err := querier.AllAllianceValidators(ctx, req.Validators)
		if err != nil {
			return nil, sdkerrors.Wrap(ErrAllianceMsg, "failed to get all alliance validators")
		}
		bz, err := json.Marshal(res)
		if err != nil {
			return nil, sdkerrors.Wrap(ErrAllianceMsg, "failed to marshal all alliance validators")
		}
		return bz, nil
	case req.Validator != nil:
		res, err := querier.AllianceValidator(ctx, req.Validator)
		if err != nil {
			return nil, sdkerrors.Wrap(ErrAllianceMsg, "failed to get alliance validator")
		}
		bz, err := json.Marshal(res)
		if err != nil {
			return nil, sdkerrors.Wrap(ErrAllianceMsg, "failed to marshal alliance validator")
		}
		return bz, nil
	case req.AlliancesDelegations != nil:
		res, err := querier.AllAlliancesDelegations(ctx, req.AlliancesDelegations)
		if err != nil {
			return nil, sdkerrors.Wrap(ErrAllianceMsg, "failed to get all alliances delegations")
		}
		bz, err := json.Marshal(res)
		if err != nil {
			return nil, sdkerrors.Wrap(ErrAllianceMsg, "failed to marshal all alliances delegations")
		}
		return bz, nil
	case req.AlliancesDelegationByValidator != nil:
		res, err := querier.AlliancesDelegationByValidator(ctx, req.AlliancesDelegationByValidator)
		if err != nil {
			return nil, sdkerrors.Wrap(ErrAllianceMsg, "failed to get all alliances delegations by validator")
		}
		bz, err := json.Marshal(res)
		if err != nil {
			return nil, sdkerrors.Wrap(ErrAllianceMsg, "failed to marshal all alliances delegations by validator")
		}
		return bz, nil
	default:
		return nil, wasmvmtypes.UnsupportedRequest{Kind: "unknown alliance query variant"}
	}
}
