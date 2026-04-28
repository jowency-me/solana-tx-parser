package solanatxparser

import "strings"

// programIDToAMM maps known program IDs to AMM names for cases where upstream
// libraries recognize the trade but output an empty or Unknown AMM string.
var programIDToAMM = map[string]string{
	"goonuddtQRrWqqn5nFyczVKaie28f3kDkHWkHtURSLE":  "GoonFi",
	"dbcij3LWUppWqq96dh6gJWwBifmcGfLSB5D4DuSMaqN":  "MeteoraDBC",
	"ZERor4xhbUycZ6gb9ntrhqscUcZmAbQDjEAtCf4hbZY":  "ZeroFi",
	"Eo7WjKq67rjJQSZxS6z3YkapzY3eMj6Xy8X5EQVn5UaB": "MeteoraDamm",
	"DEXYosS6oEGvk8uCDayvwEZz4qEyDJRf9nFgYCaqPMTm": "1Dex",
	"fUSioN9YKKSa3CUC2YUc4tPkHJ5Y6XW1yz8y6F7qWz9":  "FusionAMM",
}

// ResolveAMM looks up a program ID when the AMM name is empty or Unknown.
func ResolveAMM(programID string) string {
	if v, ok := programIDToAMM[programID]; ok {
		return v
	}
	return ""
}

// ammAliasMap normalizes AMM naming differences across upstream libraries.
var ammAliasMap = map[string]string{
	"meteora":         "MeteoraDLMM",
	"meteora-dlmm":    "MeteoraDLMM",
	"dlmm":            "MeteoraDLMM",
	"meteora-damm":    "MeteoraDamm",
	"meteora-damm-v2": "MeteoraDammV2",
	"damm":            "MeteoraDamm",
	"damm-v2":         "MeteoraDammV2",

	"raydium":      "RaydiumV4",
	"raydiumv4":    "RaydiumV4",
	"raydium-v4":   "RaydiumV4",
	"raydium-cpmm": "RaydiumCPMM",
	"raydium-cl":   "RaydiumCL",
	"raydium-clmm": "RaydiumCL",

	"pump-fun":  "Pumpfun",
	"pumpfun":   "Pumpfun",
	"pump-swap": "Pumpswap",
	"pumpswap":  "Pumpswap",

	"whirlpool":      "OrcaWhirlpool",
	"orca":           "OrcaWhirlpool",
	"orca-whirlpool": "OrcaWhirlpool",

	"jupiter":   "Jupiter",
	"jupiterv6": "JupiterV6",
}

func NormalizeAMM(name string) string {
	if name == "" {
		return ""
	}
	key := strings.ToLower(strings.ReplaceAll(name, "_", "-"))
	if v, ok := ammAliasMap[key]; ok {
		return v
	}
	return name
}
