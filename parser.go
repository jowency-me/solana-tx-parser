package solanatxparser

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"sync/atomic"
	"time"

	solana "github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/rpc"
)

type Options struct {
	RPCTimeout time.Duration
}

func DefaultOptions() Options {
	return Options{
		RPCTimeout: 20 * time.Second,
	}
}

type Parser struct {
	rpc       *rpc.Client
	providers []Provider
	opts      Options
	rpcCalls  int64
}

// NewParser creates a parser with all built-in providers registered.
func NewParser(client *rpc.Client, opts Options) *Parser {
	if opts.RPCTimeout == 0 {
		opts.RPCTimeout = 20 * time.Second
	}
	p := &Parser{
		rpc:  client,
		opts: opts,
	}
	p.Register(newCxcxProvider())
	p.Register(newDefaultPersonProvider())
	p.Register(newStakeProvider())
	p.Register(newSwapProvider())

	return p
}

// Register appends a provider. Providers registered earlier have dedup priority.
// Panics if a provider with the same name is already registered.
func (p *Parser) Register(provider Provider) {
	name := provider.Name()
	for _, existing := range p.providers {
		if existing.Name() == name {
			panic("solanatxparser: provider already registered: " + string(name))
		}
	}
	p.providers = append(p.providers, provider)
}

func (p *Parser) RPCCallCount() int64 {
	return atomic.LoadInt64(&p.rpcCalls)
}

// ParseSignature fetches and parses a single transaction through all registered providers.
func (p *Parser) ParseSignature(ctx context.Context, signature string) (*UnifiedResult, error) {
	sig, err := solana.SignatureFromBase58(signature)
	if err != nil {
		return nil, fmt.Errorf("invalid signature: %w", err)
	}

	maxVer := uint64(0)
	rctx, cancel := context.WithTimeout(ctx, p.opts.RPCTimeout)
	defer cancel()

	atomic.AddInt64(&p.rpcCalls, 1)
	rawTx, err := p.rpc.GetTransaction(rctx, sig, &rpc.GetTransactionOpts{
		Encoding:                       solana.EncodingBase64,
		MaxSupportedTransactionVersion: &maxVer,
	})
	if err != nil {
		return nil, fmt.Errorf("rpc GetTransaction: %w", err)
	}
	if rawTx == nil {
		return nil, fmt.Errorf("transaction not found")
	}

	parsedTx, err := rawTx.Transaction.GetTransaction()
	if err != nil {
		return nil, fmt.Errorf("decode transaction: %w", err)
	}

	return p.parseDecoded(signature, rawTx, parsedTx)
}

// ParseDecoded runs all registered providers on pre-fetched data without any RPC calls.
func (p *Parser) ParseDecoded(signature string, rawTx *rpc.GetTransactionResult, tx *solana.Transaction) (*UnifiedResult, error) {
	if rawTx == nil || tx == nil {
		return nil, fmt.Errorf("nil input")
	}
	return p.parseDecoded(signature, rawTx, tx)
}

func (p *Parser) parseDecoded(signature string, rawTx *rpc.GetTransactionResult, parsedTx *solana.Transaction) (*UnifiedResult, error) {
	res := &UnifiedResult{
		Signature: signature,
		Slot:      rawTx.Slot,
		Providers: map[ProviderName]*ProviderStatus{},
	}
	if rawTx.BlockTime != nil {
		bt := int64(*rawTx.BlockTime)
		res.BlockTime = &bt
	}
	if len(parsedTx.Message.AccountKeys) > 0 {
		res.Signer = parsedTx.Message.AccountKeys[0].String()
	}
	if rawTx.Meta != nil {
		res.OnchainSuccess = rawTx.Meta.Err == nil
		if rawTx.Meta.Err != nil {
			res.OnchainError = rawTx.Meta.Err
		}
		res.Fee = rawTx.Meta.Fee
		if rawTx.Meta.ComputeUnitsConsumed != nil {
			res.ComputeUnits = *rawTx.Meta.ComputeUnitsConsumed
		}
	}

	dpTx, convErr := convertToDPTransaction(signature, rawTx.Slot, res.BlockTime, parsedTx, rawTx.Meta, rawTx.Version)

	pctx := &ParseContext{
		Signature: signature,
		Slot:      rawTx.Slot,
		BlockTime: res.BlockTime,
		Tx:        parsedTx,
		Meta:      rawTx.Meta,
		Version:   rawTx.Version,
		DPTx:      dpTx,
		DPConvErr: convErr,
	}

	for _, provider := range p.providers {
		status := provider.Parse(pctx, res)
		if status == nil {
			status = &ProviderStatus{Ran: true, Error: "returned nil status"}
		}
		res.Providers[provider.Name()] = status
	}

	dedupTrades(res)
	dedupLiquidities(res)
	annotateTransfers(res)
	sort.SliceStable(res.Trades, func(i, j int) bool { return idxLess(res.Trades[i].Idx, res.Trades[j].Idx) })
	sort.SliceStable(res.Liquidities, func(i, j int) bool { return idxLess(res.Liquidities[i].Idx, res.Liquidities[j].Idx) })

	return res, nil
}

const incineratorAddr = "1nc1nerator11111111111111111111111111111111"

func annotateTransfers(r *UnifiedResult) {
	for i := range r.Transfers {
		if r.Transfers[i].To == incineratorAddr {
			r.Transfers[i].IsBurn = true
		}
	}
}

func dedupTrades(r *UnifiedResult) {
	seen := map[string]bool{}
	var out []TradeEvent
	for _, t := range r.Trades {
		k := strings.Join([]string{t.ProgramID, t.Idx, t.AMM}, "|")
		if seen[k] {
			continue
		}
		seen[k] = true
		out = append(out, t)
	}
	r.Trades = out
}

func dedupLiquidities(r *UnifiedResult) {
	seen := map[string]bool{}
	var out []LiquidityEvent
	for _, l := range r.Liquidities {
		k := strings.Join([]string{l.ProgramID, l.Idx, l.AMM, l.Action}, "|")
		if seen[k] {
			continue
		}
		seen[k] = true
		out = append(out, l)
	}
	r.Liquidities = out
}

func idxLess(a, b string) bool {
	if a == b {
		return false
	}
	pa := strings.Split(a, ".")
	pb := strings.Split(b, ".")
	for i := 0; i < len(pa) && i < len(pb); i++ {
		if pa[i] == pb[i] {
			continue
		}
		na, ea := atoiSafe(pa[i])
		nb, eb := atoiSafe(pb[i])
		if ea && eb {
			return na < nb
		}
		return pa[i] < pb[i]
	}
	return len(pa) < len(pb)
}

func atoiSafe(s string) (int, bool) {
	n := 0
	if s == "" {
		return 0, false
	}
	for _, c := range s {
		if c < '0' || c > '9' {
			return 0, false
		}
		n = n*10 + int(c-'0')
	}
	return n, true
}
