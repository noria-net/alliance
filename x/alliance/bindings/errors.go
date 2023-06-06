package bindings

import (
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/noria-net/alliance/x/alliance/types"
)

var (
	ErrAllianceMsg             = sdkerrors.Register(types.ModuleName, 1100, "invalid alliance message")
	ErrAllianceDelegateMsg     = sdkerrors.Register(types.ModuleName, 1101, "error with alliance Delegate message")
	ErrAllianceUndelegateMsg   = sdkerrors.Register(types.ModuleName, 1102, "error with alliance Undelegate message")
	ErrAllianceRedelegateMsg   = sdkerrors.Register(types.ModuleName, 1103, "error with alliance Redelegate message")
	ErrAllianceClaimRewardsMsg = sdkerrors.Register(types.ModuleName, 1104, "error with alliance ClaimRewards message")
)
