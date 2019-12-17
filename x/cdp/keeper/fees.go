package keeper

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/kava-labs/kava/x/cdp/types"
)

// CalculateFees returns the fees accumulated since fees were last calculated
func (k Keeper) CalculateFees(ctx sdk.Context, cdp types.CDP) sdk.Coins {
	newFees := sdk.NewCoins()
	timeElapsed := sdk.NewInt(ctx.BlockTime().Unix() - cdp.FeesUpdated.Unix())
	for _, pc := range cdp.Principal {
		feePerSecond := k.GetFeeRate(ctx, pc.Denom)
		feesDec := sdk.NewDecFromInt(pc.Amount).Mul(feePerSecond.Mul(sdk.NewDecFromInt(timeElapsed)))
		// TODO this will always round down, causing precision loss between the sum of all fees in CDPs and surplus coins in liquidator account
		newFees.Add(sdk.NewCoins(sdk.NewCoin(pc.Denom, feesDec.TruncateInt())))
	}
	return newFees
}

// IncrementTotalPrincipal increments the total amount of debt that has been drawn with that collateral type
func (k Keeper) IncrementTotalPrincipal(ctx sdk.Context, collateralDenom string, principal sdk.Coins) {
	for _, pc := range principal {
		total := k.GetTotalPrincipal(ctx, collateralDenom, pc.Denom)
		total = total.Add(pc.Amount)
		k.SetTotalPrincipal(ctx, collateralDenom, pc.Denom, total)
	}
}

// GetTotalPrincipal returns the total amount of principal that has been drawn for that collateral
func (k Keeper) GetTotalPrincipal(ctx sdk.Context, collateralDenom string, principalDenom string) (total sdk.Int) {
	store := prefix.NewStore(ctx.KVStore(k.key), types.PrincipalKeyPrefix)
	bz := store.Get([]byte(collateralDenom + principalDenom))
	if bz == nil {
		panic(fmt.Sprintf("total principal of %s for %s collateral not set in genesis", principalDenom, collateralDenom))
	}
	k.cdc.MustUnmarshalBinaryLengthPrefixed(bz, &total)
	return total
}

// SetTotalPrincipal sets the total amount of principal that has been draws for the input collateral
func (k Keeper) SetTotalPrincipal(ctx sdk.Context, collateralDenom string, principalDenom string, total sdk.Int) {
	store := prefix.NewStore(ctx.KVStore(k.key), types.PrincipalKeyPrefix)
	store.Set([]byte(collateralDenom+principalDenom), k.cdc.MustMarshalBinaryLengthPrefixed(total))
}

// GetFeeRate returns the per second fee rate for the input denom
func (k Keeper) GetFeeRate(ctx sdk.Context, denom string) (fee sdk.Dec) {
	store := prefix.NewStore(ctx.KVStore(k.key), types.AccumulatorKeyPrefix)
	bz := store.Get([]byte(denom))
	if bz == nil {
		panic(fmt.Sprintf("fee rate for %s not set in genesis", denom))
	}
	k.cdc.MustUnmarshalBinaryLengthPrefixed(bz, &fee)
	return fee
}

// SetFeeRate sets the per second fee rate for the input denom
func (k Keeper) SetFeeRate(ctx sdk.Context, denom string, fee sdk.Dec) {
	store := prefix.NewStore(ctx.KVStore(k.key), types.AccumulatorKeyPrefix)
	store.Set([]byte(denom), k.cdc.MustMarshalBinaryLengthPrefixed(fee))
}