package bindings

import (
	"github.com/CosmWasm/wasmd/x/wasm"

	wasmkeeper "github.com/CosmWasm/wasmd/x/wasm/keeper"

	alliancekeeper "github.com/noria-net/alliance/x/alliance/keeper"
)

func RegisterCustomPlugins(
	keeper *alliancekeeper.Keeper,
) []wasmkeeper.Option {

	queryPluginOpt := wasmkeeper.WithQueryHandlerDecorator(CustomQueryDecorator(keeper))

	// messengerDecoratorOpt := wasmkeeper.WithMessageHandlerDecorator(
	// 	CustomMessageDecorator(keeper),
	// )

	return []wasm.Option{
		queryPluginOpt,
		// messengerDecoratorOpt,
	}
}
