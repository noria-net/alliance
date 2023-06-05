package alliance

import (
	"fmt"
	"time"

	"github.com/noria-net/alliance/x/alliance/keeper"
	"github.com/noria-net/alliance/x/alliance/types"

	abci "github.com/cometbft/cometbft/abci/types"
	"github.com/cosmos/cosmos-sdk/telemetry"
	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
)

// EndBlocker
func EndBlocker(ctx sdk.Context, k keeper.Keeper) []abci.ValidatorUpdate {
	defer telemetry.ModuleMeasureSince(types.ModuleName, ctx.BlockTime(), telemetry.MetricKeyEndBlocker)
	defer telemetry.ModuleMeasureSince(stakingtypes.ModuleName, time.Now(), telemetry.MetricKeyEndBlocker)

	valUpdates := k.BlockValidatorUpdates(ctx)
	if len(valUpdates) > 0 {
		k.QueueAssetRebalanceEvent(ctx)
	}

	k.CompleteRedelegations(ctx)
	if err := k.CompleteUndelegations(ctx); err != nil {
		panic(fmt.Errorf("failed to complete undelegations from x/alliance module: %s", err))
	}

	assets := k.GetAllAssets(ctx)
	k.InitializeAllianceAssets(ctx, assets)
	if _, err := k.DeductAssetsHook(ctx, assets); err != nil {
		panic(fmt.Errorf("failed to deduct take rate from alliance in x/alliance module: %s", err))
	}
	if err := k.RewardWeightChangeHook(ctx, assets); err != nil {
		panic(fmt.Errorf("failed to update assets reward weights in x/alliance module: %s", err))
	}

	newUpdates, err := k.RebalanceHook(ctx, assets)
	if err != nil {
		panic(fmt.Errorf("failed to rebalance assets in x/alliance module: %s", err))
	}

	if len(valUpdates) > 0 && len(newUpdates) == 0 && len(assets) > 0 {
		// fmt.Printf("\n########### Already processed changes in previous block\n")
		return []abci.ValidatorUpdate{}
	}

	// fmt.Printf("\n########### Native Voting Power Changes: %v\n", valUpdates)
	// fmt.Printf("\n########### Weighted Voting Power Changes: %v\n", newUpdates)

	// merge validator updates
	mergedUpdates := mergeValidatorUpdates(valUpdates, newUpdates)
	// fmt.Printf("\n########### Merged Validator Updates: %v\n\n", mergedUpdates)

	return mergedUpdates
}

func mergeValidatorUpdates(a, b []abci.ValidatorUpdate) []abci.ValidatorUpdate {
	for _, v := range b {
		found := false
		for i, u := range a {
			if u.PubKey.Equal(v.PubKey) {
				a[i] = v
				found = true
				break
			}
		}
		if !found {
			a = append(a, v)
		}
	}
	return a
}
