package types

type AllianceQuery struct {
	Alliance *AllianceSubQuery `json:"alliance,omitempty"`
}

type AllianceSubQuery struct {
	Alliance          *Alliance          `json:"alliance"`
	Delegation        *Delegation        `json:"delegation"`
	DelegationRewards *DelegationRewards `json:"delegation_rewards"`
	Params            *ParamsRequest     `json:"params"`
}

type Alliance struct {
	Denom string `json:"denom"`
}

type Delegation struct {
	Denom     string `json:"denom"`
	Delegator string `json:"delegator"`
	Validator string `json:"validator"`
}

type DelegationRewards struct {
	Denom     string `json:"denom"`
	Delegator string `json:"delegator"`
	Validator string `json:"validator"`
}

type ParamsRequest struct {
}
