package bindings

import (
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/noria-net/alliance/x/alliance/types"
)

var (
	ErrAllianceMsg = sdkerrors.Register(types.ModuleName, 1100, "invalid alliance message")
)
