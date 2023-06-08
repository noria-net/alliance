package types

type AllianceResponse struct {
	Denom                string            `json:"denom"`
	RewardWeight         string            `json:"reward_weight"`
	ConsensusWeight      string            `json:"consensus_weight"`
	ConsensusCap         string            `json:"consensus_cap"`
	TakeRate             string            `json:"take_rate"`
	TotalTokens          string            `json:"total_tokens"`
	TotalValidatorShares string            `json:"total_validator_shares"`
	RewardStartTime      uint64            `json:"reward_start_time"`
	RewardChangeRate     string            `json:"reward_change_rate"`
	LastRewardChangeTime uint64            `json:"last_reward_change_time"`
	RewardWeightRange    RewardWeightRange `json:"reward_weight_range"`
	IsInitialized        bool              `json:"is_initialized"`
}

type RewardWeightRange struct {
	Min string `json:"min"`
	Max string `json:"max"`
}

type Coin struct {
	Denom  string `json:"denom"`
	Amount string `json:"amount"`
}

type DelegationResponse struct {
	Delegator string `json:"delegator"`
	Validator string `json:"validator"`
	Denom     string `json:"denom"`
	Amount    Coin   `json:"amount"`
}

type DelegationRewardsResponse struct {
	Rewards []Coin `json:"rewards"`
}
