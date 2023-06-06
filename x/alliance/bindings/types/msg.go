package types

import (
	"github.com/noria-net/alliance/x/alliance/types"
)

type AllianceMsg struct {
	Alliance *AllianceSubMsg `json:"alliance,omitempty"`
}

type AllianceSubMsg struct {
	Delegate               *types.MsgDelegate               `json:"delegate"`
	Undelegate             *types.MsgUndelegate             `json:"undelegate"`
	Redelegate             *types.MsgRedelegate             `json:"redelegate"`
	ClaimDelegationRewards *types.MsgClaimDelegationRewards `json:"claim_delegation_rewards"`
}
