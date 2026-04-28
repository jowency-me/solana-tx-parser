package solanatxparser

import (
	"fmt"
	"time"

	dexparser "github.com/DefaultPerson/solana-dex-parser-go"
	dpTypes "github.com/DefaultPerson/solana-dex-parser-go/types"
)

type defaultPersonProvider struct {
	inner *dexparser.DexParser
	cfg   *dpTypes.ParseConfig
}

func newDefaultPersonProvider() *defaultPersonProvider {
	cfg := dpTypes.DefaultParseConfig()
	return &defaultPersonProvider{
		inner: dexparser.NewDexParser(),
		cfg:   &cfg,
	}
}

func (p *defaultPersonProvider) Name() ProviderName { return ProviderDefaultPerson }

func (p *defaultPersonProvider) Parse(ctx *ParseContext, result *UnifiedResult) *ProviderStatus {
	status := &ProviderStatus{}
	start := time.Now()
	status.Ran = true

	defer func() {
		status.Elapsed = time.Since(start).String()
		if rec := recover(); rec != nil {
			status.Error = fmt.Sprintf("panic: %v", rec)
			result.Trades = nil
			result.Liquidities = nil
		}
	}()

	if ctx.DPTx == nil {
		if ctx.DPConvErr != nil {
			status.Error = "convert: " + ctx.DPConvErr.Error()
		} else {
			status.Error = "no pre-converted transaction data"
		}
		return status
	}

	parsed := p.inner.ParseAll(ctx.DPTx, p.cfg)
	if parsed == nil {
		return status
	}

	for _, t := range parsed.Trades {
		amm := t.AMM
		if amm == "" && len(t.AMMs) > 0 {
			amm = t.AMMs[0]
		}
		amm = NormalizeAMM(amm)
		if amm == "" || amm == "Unknown" {
			if v := ResolveAMM(t.ProgramId); v != "" {
				amm = v
			}
		}
		result.Trades = append(result.Trades, TradeEvent{
			Layer: ProviderDefaultPerson, Idx: t.Idx, Type: string(t.Type),
			User: t.User, AMM: amm,
			Route: t.Route, ProgramID: t.ProgramId,
			Pools: append([]string{}, t.Pool...),
			InputToken: TokenAmount{
				Mint: t.InputToken.Mint, Amount: t.InputToken.AmountRaw,
				UIAmount: t.InputToken.Amount, Decimals: t.InputToken.Decimals,
			},
			OutputToken: TokenAmount{
				Mint: t.OutputToken.Mint, Amount: t.OutputToken.AmountRaw,
				UIAmount: t.OutputToken.Amount, Decimals: t.OutputToken.Decimals,
			},
		})
	}

	for _, l := range parsed.Liquidities {
		amm := NormalizeAMM(l.AMM)
		if amm == "" || amm == "Unknown" {
			if v := ResolveAMM(l.ProgramId); v != "" {
				amm = v
			}
		}
		ev := LiquidityEvent{
			Layer: ProviderDefaultPerson, Idx: l.Idx,
			Action: string(l.Type), User: l.User,
			AMM: amm, ProgramID: l.ProgramId, PoolID: l.PoolId,
		}
		ev.BaseToken = TokenAmount{Mint: l.Token0Mint, Amount: l.Token0AmountRaw}
		if l.Token0Amount != nil {
			ev.BaseToken.UIAmount = *l.Token0Amount
		}
		if l.Token0Decimals != nil {
			ev.BaseToken.Decimals = *l.Token0Decimals
		}
		ev.QuoteToken = TokenAmount{Mint: l.Token1Mint, Amount: l.Token1AmountRaw}
		if l.Token1Amount != nil {
			ev.QuoteToken.UIAmount = *l.Token1Amount
		}
		if l.Token1Decimals != nil {
			ev.QuoteToken.Decimals = *l.Token1Decimals
		}
		result.Liquidities = append(result.Liquidities, ev)
	}

	return status
}
