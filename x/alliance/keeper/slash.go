package keeper

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	"github.com/noria-net/alliance/x/alliance/types"
)

/*
If there is a slash receiver defined, the slashed tokens are sent to the slash receiver.
*/
func (k Keeper) RedirectSlashedCoins(ctx sdk.Context, valAddr sdk.ValAddress, coins sdk.Coins) {
	slashReceieverParam := k.SlashReceiver(ctx)
	fmt.Printf("\n #### Slash Receiver Address: %v\n", slashReceieverParam)
	if len(slashReceieverParam) > 0 {
		receiver, err := sdk.AccAddressFromBech32(slashReceieverParam)
		if err != nil {
			panic("Invalid slash_receiver alliance parameter")
		}

		fmt.Printf("\n #### Redirecting slashed alliance funds to %s: %v\n\n", receiver, coins)
		k.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, receiver, coins)
	}
}

func (k Keeper) getNativeSlashingInfo(ctx sdk.Context, val types.AllianceValidator, fraction sdk.Dec) (sdk.Dec, sdk.Dec) {
	currentBondedAmount := sdk.NewDec(0)
	moduleAddr := k.accountKeeper.GetModuleAddress(types.ModuleName)
	delegation, found := k.stakingKeeper.GetDelegation(ctx, moduleAddr, val.GetOperator())
	if found {
		currentBondedAmount = val.TokensFromShares(delegation.GetShares())
	}
	fmt.Printf("\n\n #### Alliance Staking Amount: %v\n", currentBondedAmount)

	nativeBondAmountForValidator := sdk.NewDec(val.Validator.Tokens.Int64()).Sub(currentBondedAmount)
	originalFraction := fraction

	if currentBondedAmount.GT(sdk.NewDec(0)) {
		powerReduction := sdk.NewDecFromInt(k.stakingKeeper.PowerReduction(ctx))
		// nativePower := (nativeBondAmountForValidator.Quo(powerReduction))
		originalFraction = sdk.NewDecFromInt(val.Validator.Tokens).Mul(fraction).Quo(sdk.NewDec(val.AllianceValidatorInfo.VotingPower).Mul(powerReduction))
	}

	fmt.Printf("\n #### Native Staking Amount: %v\n", nativeBondAmountForValidator)
	fmt.Printf("\n #### Original Slashing Fraction: %v\n", originalFraction)
	return originalFraction, nativeBondAmountForValidator
}

func (k Keeper) MintAndRedirectSlashedNativeCoins(ctx sdk.Context, valAddr sdk.ValAddress, fraction sdk.Dec) error {
	slashReceieverParam := k.SlashReceiver(ctx)
	if len(slashReceieverParam) > 0 {
		receiver, err := sdk.AccAddressFromBech32(slashReceieverParam)
		if err != nil {
			panic("Invalid slash_receiver alliance parameter")
		}

		fmt.Printf("\n\n #### Minting and redirecting slashed funds to %s: (effectiveFraction): %v\n\n", receiver, fraction)

		if fraction.LTE(sdk.ZeroDec()) || fraction.GT(sdk.OneDec()) {
			return fmt.Errorf("slashed fraction must be greater than 0 and less than or equal to 1: %d", fraction)
		}

		val, err := k.GetAllianceValidator(ctx, valAddr)
		if err != nil {
			return err
		}

		originalSlashingFraction, nativeBondAmountForValidator := k.getNativeSlashingInfo(ctx, val, fraction)
		amountToSlash := nativeBondAmountForValidator.Mul(originalSlashingFraction)
		fmt.Printf("\n #### Native Amount Slashed to send to Receiver: %v\n", amountToSlash.TruncateInt())
		originalStakingSlashAmount := sdk.NewDecFromInt(val.Validator.Tokens).Mul(fraction)
		fmt.Printf("\n #### Slash Amount Burned By Staking Module: %v\n", originalStakingSlashAmount.TruncateInt())
		totalAmountToSlash := sdk.NewDecFromInt(val.Validator.Tokens).Mul(originalSlashingFraction)
		fmt.Printf("\n #### Total Amount That Should Be Burned for that slash: %v\n", totalAmountToSlash.TruncateInt())
		missingAmountToBurn := totalAmountToSlash.Sub(originalStakingSlashAmount)
		fmt.Printf("\n #### Missing Amount To Burn By Alliance: %v\n\n", missingAmountToBurn.TruncateInt())

		bondDenom := k.stakingKeeper.BondDenom(ctx)

		if missingAmountToBurn.GT(sdk.ZeroDec()) {
			// If the slashing amount is greater than the amount that the staking module will burn, we need to burn the difference
			coinsToBurn := sdk.NewCoins(sdk.NewCoin(bondDenom, missingAmountToBurn.TruncateInt()))
			k.bankKeeper.BurnCoins(ctx, stakingtypes.BondedPoolName, coinsToBurn)
		}

		coinsToSlash := sdk.NewCoins(sdk.NewCoin(bondDenom, amountToSlash.TruncateInt()))
		k.bankKeeper.MintCoins(ctx, types.ModuleName, coinsToSlash)
		k.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, receiver, coinsToSlash)

		fmt.Printf("\n #### Redirecting slashed native funds to %s: %v\n\n", receiver, coinsToSlash)
	}
	return nil
}

// SlashValidator works by reducing the amount of validator shares for all alliance assets by a `fraction`
// This effectively reallocates tokens from slashed validators to good validators
// On top of slashing currently bonded delegations, we also slash re-delegations and un-delegations
// that are still in the progress of unbonding
func (k Keeper) SlashValidator(ctx sdk.Context, valAddr sdk.ValAddress, fraction sdk.Dec) error {
	// Slashing must be checked otherwise we can end up slashing incorrect amounts

	fmt.Print("\n\n #### ALLIANCE BEFORE SLASH HOOK \n\n")

	if fraction.LTE(sdk.ZeroDec()) || fraction.GT(sdk.OneDec()) {
		return fmt.Errorf("slashed fraction must be greater than 0 and less than or equal to 1: %d", fraction)
	}

	val, err := k.GetAllianceValidator(ctx, valAddr)
	if err != nil {
		return err
	}
	originalSlashingFraction, _ := k.getNativeSlashingInfo(ctx, val, fraction)

	// slashedValidatorShares accumulates the final validator shares after slashing
	slashedValidatorShares := sdk.NewDecCoins()
	for _, share := range val.ValidatorShares {
		sharesToSlash := share.Amount.Mul(fraction)

		k.RedirectSlashedCoins(ctx, valAddr, sdk.NewCoins(sdk.NewCoin(share.Denom, share.Amount.Mul(originalSlashingFraction).TruncateInt())))

		sharesAfterSlashing := sdk.NewDecCoinFromDec(share.Denom, share.Amount.Sub(sharesToSlash))
		slashedValidatorShares = slashedValidatorShares.Add(sharesAfterSlashing)
		asset, found := k.GetAssetByDenom(ctx, share.Denom)
		if !found {
			return types.ErrUnknownAsset
		}
		asset.TotalValidatorShares = asset.TotalValidatorShares.Sub(sharesToSlash)
		k.SetAsset(ctx, asset)
	}
	val.ValidatorShares = slashedValidatorShares
	k.SetValidator(ctx, val)

	err = k.slashRedelegations(ctx, valAddr, fraction)
	if err != nil {
		return err
	}

	err = k.slashUndelegations(ctx, valAddr, fraction)
	if err != nil {
		return err
	}
	return nil
}

func (k Keeper) slashRedelegations(ctx sdk.Context, valAddr sdk.ValAddress, fraction sdk.Dec) error {
	store := ctx.KVStore(k.storeKey)
	// Slash all immature re-delegations
	redelegationIterator := k.IterateRedelegationsBySrcValidator(ctx, valAddr)
	for ; redelegationIterator.Valid(); redelegationIterator.Next() {
		redelegationKey, completion, err := types.ParseRedelegationIndexForRedelegationKey(redelegationIterator.Key())
		if err != nil {
			return err
		}
		// Skip if redelegation is already mature
		if completion.Before(ctx.BlockTime()) {
			continue
		}
		b := store.Get(redelegationKey)
		var redelegation types.Redelegation
		k.cdc.MustUnmarshal(b, &redelegation)

		delAddr, err := sdk.AccAddressFromBech32(redelegation.DelegatorAddress)
		if err != nil {
			return err
		}
		dstValAddr, err := sdk.ValAddressFromBech32(redelegation.DstValidatorAddress)
		if err != nil {
			return err
		}
		dstVal, err := k.GetAllianceValidator(ctx, dstValAddr)
		if err != nil {
			return err
		}

		_, err = k.ClaimDelegationRewards(ctx, delAddr, dstVal, redelegation.Balance.Denom)
		if err != nil {
			return err
		}

		delegation, found := k.GetDelegation(ctx, delAddr, dstVal.GetOperator(), redelegation.Balance.Denom)
		if !found {
			continue
		}

		asset, found := k.GetAssetByDenom(ctx, redelegation.Balance.Denom)
		if !found {
			continue
		}

		// Slash delegation shares
		tokensToSlash := fraction.MulInt(redelegation.Balance.Amount).TruncateInt()
		sharesToSlash, err := k.ValidateDelegatedAmount(delegation, sdk.NewCoin(redelegation.Balance.Denom, tokensToSlash), dstVal, asset)
		if err != nil {
			return err
		}
		k.RedirectSlashedCoins(ctx, valAddr, sdk.NewCoins(sdk.NewCoin(redelegation.Balance.Denom, sharesToSlash.TruncateInt())))
		dstVal.TotalDelegatorShares = sdk.DecCoins(dstVal.TotalDelegatorShares).Sub(sdk.NewDecCoins(sdk.NewDecCoinFromDec(asset.Denom, sharesToSlash)))
		k.SetValidator(ctx, dstVal)

		delegation.Shares = delegation.Shares.Sub(sharesToSlash)
		k.SetDelegation(ctx, delAddr, dstVal.GetOperator(), asset.Denom, delegation)
	}
	return nil
}

func (k Keeper) slashUndelegations(ctx sdk.Context, valAddr sdk.ValAddress, fraction sdk.Dec) error {
	store := ctx.KVStore(k.storeKey)
	// Slash all immature re-delegations
	undelegationIterator := k.IterateUndelegationsBySrcValidator(ctx, valAddr)
	for ; undelegationIterator.Valid(); undelegationIterator.Next() {
		undelegationKey, completion, err := types.ParseUnbondingIndexKeyToUndelegationKey(undelegationIterator.Key())
		if err != nil {
			return err
		}
		// Skip if undelegation is already mature
		if completion.Before(ctx.BlockTime()) {
			continue
		}
		b := store.Get(undelegationKey)
		var undelegations types.QueuedUndelegation
		k.cdc.MustUnmarshal(b, &undelegations)

		// Slash undelegations by sending slashed tokens to fee pool
		for _, entry := range undelegations.Entries {
			tokensToSlash := fraction.MulInt(entry.Balance.Amount).TruncateInt()
			entry.Balance = sdk.NewCoin(entry.Balance.Denom, entry.Balance.Amount.Sub(tokensToSlash))
			coinToSlash := sdk.NewCoin(entry.Balance.Denom, tokensToSlash)
			k.RedirectSlashedCoins(ctx, valAddr, sdk.NewCoins(coinToSlash))
		}
		b = k.cdc.MustMarshal(&undelegations)
		store.Set(undelegationKey, b)
	}
	return nil
}
