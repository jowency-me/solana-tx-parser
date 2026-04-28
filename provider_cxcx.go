package solanatxparser

import (
	"fmt"
	"time"

	cxParsers "github.com/cxcx-ai/solana-parser-go/parsers"
	cxTypes "github.com/cxcx-ai/solana-parser-go/types"
)

type cxcxProvider struct {
	inner *cxParsers.DexParser
}

func newCxcxProvider() *cxcxProvider {
	cfg := &cxTypes.ParseConfig{
		ParseType: cxTypes.ParseType{
			AggregateTrade: true, Trade: true, Liquidity: true,
			Transfer: true, MemeEvent: true, AltEvent: true,
		},
		TryUnknowDEX: true,
	}
	return &cxcxProvider{inner: cxParsers.NewDexParser(cfg)}
}

func (p *cxcxProvider) Name() ProviderName { return ProviderCxcx }

func (p *cxcxProvider) Parse(ctx *ParseContext, result *UnifiedResult) *ProviderStatus {
	status := &ProviderStatus{}
	start := time.Now()
	status.Ran = true

	var blockTimeVal int64
	if ctx.BlockTime != nil {
		blockTimeVal = *ctx.BlockTime
	}

	defer func() {
		status.Elapsed = time.Since(start).String()
		if rec := recover(); rec != nil {
			status.Error = fmt.Sprintf("panic: %v", rec)
			result.Trades = nil
			result.Liquidities = nil
			result.Transfers = nil
			result.MemeEvents = nil
		}
	}()

	in := &cxTypes.SolanaTransaction{
		Slot: ctx.Slot, BlockTime: blockTimeVal,
		Transaction: ctx.Tx, Meta: ctx.Meta,
	}
	parsed := p.inner.Parse(in)

	for _, t := range parsed.Trades {
		amm := ""
		if len(t.Amms) > 0 {
			amm = t.Amms[0]
		}
		amm = NormalizeAMM(amm)
		if amm == "" || amm == "Unknown" {
			if v := ResolveAMM(t.ProgramID.String()); v != "" {
				amm = v
			}
		}
		result.Trades = append(result.Trades, TradeEvent{
			Layer:     ProviderCxcx,
			Idx:       t.Idx,
			Type:      string(t.Type),
			User:      t.User.String(),
			AMM:       amm,
			Route:     t.Route,
			ProgramID: t.ProgramID.String(),
			Pools:     t.GetPools(),
			InputToken: TokenAmount{
				Mint: pubkeyStr(t.InputToken.Mint), Amount: t.InputToken.Amount,
				UIAmount: t.InputToken.UiAmount, Decimals: t.InputToken.Decimals,
			},
			OutputToken: TokenAmount{
				Mint: pubkeyStr(t.OutputToken.Mint), Amount: t.OutputToken.Amount,
				UIAmount: t.OutputToken.UiAmount, Decimals: t.OutputToken.Decimals,
			},
		})
	}

	for _, l := range parsed.Liquidities {
		amm := NormalizeAMM(l.Amm)
		if amm == "" || amm == "Unknown" {
			if v := ResolveAMM(l.ProgramID.String()); v != "" {
				amm = v
			}
		}
		result.Liquidities = append(result.Liquidities, LiquidityEvent{
			Layer: ProviderCxcx, Idx: l.Idx,
			Action: string(l.Type), User: l.User.String(),
			AMM: amm, ProgramID: l.ProgramID.String(),
			PoolID: l.PoolID.String(),
			BaseToken: TokenAmount{
				Mint: pubkeyStr(l.BaseToken.Mint), Amount: l.BaseToken.Amount,
				UIAmount: l.BaseToken.UiAmount, Decimals: l.BaseToken.Decimals,
			},
			QuoteToken: TokenAmount{
				Mint: pubkeyStr(l.QuoteToken.Mint), Amount: l.QuoteToken.Amount,
				UIAmount: l.QuoteToken.UiAmount, Decimals: l.QuoteToken.Decimals,
			},
		})
	}

	for _, t := range parsed.Transfers {
		from := t.GetFrom()
		to := t.GetTo()
		result.Transfers = append(result.Transfers, TransferEvent{
			Layer: ProviderCxcx, Idx: t.Idx,
			From: from.String(), To: to.String(),
			IsSelf: !from.IsZero() && from.Equals(to),
			IsBurn: t.Type == "burn" || t.Type == "burnChecked",
			Token: TokenAmount{
				Mint: t.Info.Mint.String(), Amount: t.Info.TokenAmount.Amount,
				UIAmount: t.Info.TokenAmount.UiAmount, Decimals: t.Info.TokenAmount.Decimals,
			},
			IsFee: t.IsFee, ProgramID: t.ProgramID.String(),
		})
	}

	for _, m := range parsed.MemeEvents {
		ev := MemeEvent{
			Layer: ProviderCxcx, Idx: m.Idx,
			Action: string(m.Type), Protocol: m.Protocol,
		}
		if m.User != nil {
			ev.User = m.User.String()
		}
		if m.BaseMint != nil {
			ev.BaseMint = m.BaseMint.String()
		}
		if m.Name != nil {
			ev.Name = *m.Name
		}
		if m.Symbol != nil {
			ev.Symbol = *m.Symbol
		}
		result.MemeEvents = append(result.MemeEvents, ev)
	}

	return status
}
