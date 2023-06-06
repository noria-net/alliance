package bindings

import (
	"encoding/json"

	sdkerrors "cosmossdk.io/errors"
	wasmkeeper "github.com/CosmWasm/wasmd/x/wasm/keeper"
	wasmvmtypes "github.com/CosmWasm/wasmvm/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	bindings "github.com/noria-net/alliance/x/alliance/bindings/types"
	alliancekeeper "github.com/noria-net/alliance/x/alliance/keeper"
)

// CustomMessageDecorator returns decorator for custom CosmWasm bindings messages
func CustomMessageDecorator(alliance *alliancekeeper.Keeper) func(wasmkeeper.Messenger) wasmkeeper.Messenger {
	return func(old wasmkeeper.Messenger) wasmkeeper.Messenger {
		return &CustomMessenger{
			wrapped:  old,
			alliance: alliance,
		}
	}
}

type CustomMessenger struct {
	wrapped  wasmkeeper.Messenger
	alliance *alliancekeeper.Keeper
}

var _ wasmkeeper.Messenger = (*CustomMessenger)(nil)

// DispatchMsg executes on the contractMsg.
func (m *CustomMessenger) DispatchMsg(ctx sdk.Context, contractAddr sdk.AccAddress, contractIBCPortID string, msg wasmvmtypes.CosmosMsg) ([]sdk.Event, [][]byte, error) {
	if msg.Custom != nil {
		// only handle the happy path where this is really delegating / undelegating / redelegating / claiming rewards ...
		// leave everything else for the wrapped version
		var contractMsg bindings.AllianceMsg
		if err := json.Unmarshal(msg.Custom, &contractMsg); err != nil {
			return nil, nil, sdkerrors.Wrap(ErrAllianceMsg, "requires 'alliance' field")
		}
		if contractMsg.Alliance == nil {
			return m.wrapped.DispatchMsg(ctx, contractAddr, contractIBCPortID, msg)
		}
		allianceMsg := contractMsg.Alliance
		msgServer := alliancekeeper.NewMsgServerImpl(*m.alliance)

		if allianceMsg.Delegate != nil {

			if err := contractMsg.Alliance.Delegate.ValidateBasic(); err != nil {
				return nil, nil, sdkerrors.Wrapf(ErrAllianceDelegateMsg, "failed validating MsgDelegate: %s", err.Error())
			}

			resp, err := msgServer.Delegate(
				sdk.WrapSDKContext(ctx),
				allianceMsg.Delegate,
			)
			if err != nil {
				return nil, nil, sdkerrors.Wrap(ErrAllianceDelegateMsg, err.Error())
			}

			bz, err := json.Marshal(resp)
			if err != nil {
				return nil, nil, sdkerrors.Wrap(ErrAllianceDelegateMsg, err.Error())
			}

			return nil, [][]byte{bz}, err
		}
		if allianceMsg.Undelegate != nil {
			if err := contractMsg.Alliance.Undelegate.ValidateBasic(); err != nil {
				return nil, nil, sdkerrors.Wrapf(ErrAllianceUndelegateMsg, "failed validating MsgUndelegate: %s", err.Error())
			}

			resp, err := msgServer.Undelegate(
				sdk.WrapSDKContext(ctx),
				allianceMsg.Undelegate,
			)
			if err != nil {
				return nil, nil, sdkerrors.Wrap(ErrAllianceUndelegateMsg, err.Error())
			}

			bz, err := json.Marshal(resp)
			if err != nil {
				return nil, nil, sdkerrors.Wrap(ErrAllianceUndelegateMsg, err.Error())
			}

			return nil, [][]byte{bz}, err
		}
		if allianceMsg.Redelegate != nil {
			if err := contractMsg.Alliance.Redelegate.ValidateBasic(); err != nil {
				return nil, nil, sdkerrors.Wrapf(ErrAllianceRedelegateMsg, "failed validating MsgRedelegate: %s", err.Error())
			}

			resp, err := msgServer.Redelegate(
				sdk.WrapSDKContext(ctx),
				allianceMsg.Redelegate,
			)
			if err != nil {
				return nil, nil, sdkerrors.Wrap(ErrAllianceRedelegateMsg, err.Error())
			}

			bz, err := json.Marshal(resp)
			if err != nil {
				return nil, nil, sdkerrors.Wrap(ErrAllianceRedelegateMsg, err.Error())
			}

			return nil, [][]byte{bz}, err
		}
		if allianceMsg.ClaimDelegationRewards != nil {
			if err := contractMsg.Alliance.ClaimDelegationRewards.ValidateBasic(); err != nil {
				return nil, nil, sdkerrors.Wrapf(ErrAllianceClaimRewardsMsg, "failed validating MsgClaimRewards: %s", err.Error())
			}

			resp, err := msgServer.ClaimDelegationRewards(
				sdk.WrapSDKContext(ctx),
				allianceMsg.ClaimDelegationRewards,
			)
			if err != nil {
				return nil, nil, sdkerrors.Wrap(ErrAllianceClaimRewardsMsg, err.Error())
			}

			bz, err := json.Marshal(resp)
			if err != nil {
				return nil, nil, sdkerrors.Wrap(ErrAllianceClaimRewardsMsg, err.Error())
			}

			return nil, [][]byte{bz}, err
		}
	}
	return m.wrapped.DispatchMsg(ctx, contractAddr, contractIBCPortID, msg)
}
