package types

import (
	"github.com/noria-net/alliance/x/alliance/types"
)

type AllianceQuery struct {
	Alliance *AllianceSubQuery `json:"alliance,omitempty"`
}

type AllianceSubQuery struct {
	Alliance             *types.QueryAllianceRequest                  `json:"alliance"`
	Alliances            *types.QueryAlliancesRequest                 `json:"alliances"`
	AlliancesDelegations *types.QueryAllAlliancesDelegationsRequest   `json:"alliances_delegations"`
	Delegation           *types.QueryAllianceDelegationRequest        `json:"delegation"`
	DelegationRewards    *types.QueryAllianceDelegationRewardsRequest `json:"delegation_rewards"`
	Params               *types.QueryParamsRequest                    `json:"params"`
	Validator            *types.QueryAllianceValidatorRequest         `json:"validator"`
	Validators           *types.QueryAllAllianceValidatorsRequest     `json:"validators"`
}
