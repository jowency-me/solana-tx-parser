// Package solanatxparser provides an extensible parser for Solana transactions.
//
// Parsers are registered in priority order and each contributes typed events
// to a UnifiedResult. The library ships with three built-in providers:
//
//   - "cxcx" (cxcx-ai/solana-parser-go)           — 50+ DEX protocols, Meme, Jupiter routing
//   - "defaultperson" (DefaultPerson/solana-dex-parser-go) — DLMM / DAMM v2 fallback
//   - "stake" (gagliardetto/solana-go/programs/stake)      — 18 Stake program instructions
//
// To add a new parser, implement the Provider interface and call Register.
package solanatxparser

import (
	"github.com/DefaultPerson/solana-dex-parser-go/adapter"
	solana "github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/rpc"
)

type EventKind string

const (
	EventTrade        EventKind = "TRADE"
	EventLiquidity    EventKind = "LIQUIDITY"
	EventTransfer     EventKind = "TRANSFER"
	EventSelfTransfer EventKind = "SELF_TRANSFER"
	EventMeme         EventKind = "MEME"
	EventStake        EventKind = "STAKE"
)

// ParserName identifies a registered provider.
type ProviderName string

const (
	ProviderCxcx          ProviderName = "cxcx"
	ProviderDefaultPerson ProviderName = "defaultperson"
	ProviderStake         ProviderName = "stake"

	ProviderSwap ProviderName = "swap"
)

type TokenAmount struct {
	Mint     string  `json:"mint"`
	Amount   string  `json:"amount"`
	UIAmount float64 `json:"uiAmount"`
	Decimals uint8   `json:"decimals"`
	Symbol   string  `json:"symbol,omitempty"`
}

type TradeEvent struct {
	Layer       ProviderName `json:"layer"`
	Idx         string       `json:"idx"`
	Type        string       `json:"type"`
	User        string       `json:"user"`
	AMM         string       `json:"amm"`
	Route       string       `json:"route"`
	ProgramID   string       `json:"programId"`
	Pools       []string     `json:"pools"`
	InputToken  TokenAmount  `json:"inputToken"`
	OutputToken TokenAmount  `json:"outputToken"`
}

type LiquidityEvent struct {
	Layer      ProviderName `json:"layer"`
	Idx        string       `json:"idx"`
	Action     string       `json:"action"`
	User       string       `json:"user"`
	AMM        string       `json:"amm"`
	ProgramID  string       `json:"programId"`
	PoolID     string       `json:"poolId"`
	BaseToken  TokenAmount  `json:"baseToken"`
	QuoteToken TokenAmount  `json:"quoteToken"`
}

type TransferEvent struct {
	Layer     ProviderName `json:"layer"`
	Idx       string       `json:"idx"`
	From      string       `json:"from"`
	To        string       `json:"to"`
	IsSelf    bool         `json:"isSelf"`
	IsBurn    bool         `json:"isBurn"`
	Token     TokenAmount  `json:"token"`
	IsFee     bool         `json:"isFee"`
	ProgramID string       `json:"programId"`
}

type MemeEvent struct {
	Layer    ProviderName `json:"layer"`
	Idx      string       `json:"idx"`
	Action   string       `json:"action"`
	Protocol string       `json:"protocol"`
	User     string       `json:"user"`
	BaseMint string       `json:"baseMint,omitempty"`
	Name     string       `json:"name,omitempty"`
	Symbol   string       `json:"symbol,omitempty"`
}

type StakeEvent struct {
	Layer        ProviderName `json:"layer"`
	Idx          string       `json:"idx"`
	Instruction  string       `json:"instruction"`
	StakeAccount string       `json:"stakeAccount,omitempty"`
	VoteAccount  string       `json:"voteAccount,omitempty"`
}

// ProviderStatus tracks whether a provider ran and how long it took.
type ProviderStatus struct {
	Ran     bool   `json:"ran"`
	Error   string `json:"error,omitempty"`
	Elapsed string `json:"elapsed,omitempty"`
}

type UnifiedResult struct {
	Signature string `json:"signature"`
	Slot      uint64 `json:"slot"`
	BlockTime *int64 `json:"blockTime,omitempty"`
	Signer    string `json:"signer"`

	OnchainSuccess bool        `json:"onchainSuccess"`
	OnchainError   interface{} `json:"onchainError,omitempty"`
	Fee            uint64      `json:"fee"`
	ComputeUnits   uint64      `json:"computeUnits"`

	Trades      []TradeEvent     `json:"trades"`
	Liquidities []LiquidityEvent `json:"liquidities"`
	Transfers   []TransferEvent  `json:"transfers"`
	MemeEvents  []MemeEvent      `json:"memeEvents"`
	StakeEvents []StakeEvent     `json:"stakeEvents"`

	Providers map[ProviderName]*ProviderStatus `json:"providers"`
}

// ParseContext holds the pre-fetched transaction data shared by all providers.
type ParseContext struct {
	Signature string
	Slot      uint64
	BlockTime *int64
	Tx        *solana.Transaction
	Meta      *rpc.TransactionMeta
	Version   interface{}
	DPTx      *adapter.SolanaTransaction
	DPConvErr error
}

// Provider contributes typed events to a UnifiedResult.
// Implementations must be safe for concurrent use.
type Provider interface {
	Name() ProviderName
	Parse(ctx *ParseContext, result *UnifiedResult) *ProviderStatus
}

func (r *UnifiedResult) HasAnyEvent() bool {
	return len(r.Trades)+len(r.Liquidities)+len(r.Transfers)+len(r.MemeEvents)+len(r.StakeEvents) > 0
}

func (r *UnifiedResult) SelfTransfers() []TransferEvent {
	var out []TransferEvent
	for _, t := range r.Transfers {
		if t.IsSelf {
			out = append(out, t)
		}
	}
	return out
}

func (r *UnifiedResult) EventCounts() map[EventKind]int {
	m := map[EventKind]int{
		EventTrade:     len(r.Trades),
		EventLiquidity: len(r.Liquidities),
		EventMeme:      len(r.MemeEvents),
		EventStake:     len(r.StakeEvents),
	}
	for _, t := range r.Transfers {
		m[EventTransfer]++
		if t.IsSelf {
			m[EventSelfTransfer]++
		}
	}
	return m
}

func pubkeyStr(p solana.PublicKey) string {
	if p.IsZero() {
		return ""
	}
	return p.String()
}
