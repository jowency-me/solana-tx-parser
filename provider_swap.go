package solanatxparser

import (
	"encoding/binary"
	"fmt"
	"math"
	"math/big"
	"time"

	solana "github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/rpc"
)

// ---------------------------------------------------------------------------
// Swap provider — transfer-based swap detection for known DEX programs
// ---------------------------------------------------------------------------

var swapProtocols = []string{
	"ALPHAQmeA7bjrVuccPsYPiCvsi428SNwte66Srvs4pHA",
	"FLUXubRmkEi2q6K3Y9kBPg9248ggaZVsoSFhtJHSrm1X",
	"REALQqNEomY6cQGZJUGwywTBD2UmDT32rZcNnfxQ5N2",
	"T1TANpTeScyeqVzzgNViGDNrkQ6qHz9KrSBS4aNXvGT",
	"SV2EYYJyRz2YhfXwXnhNAevDEui5Q6yrfyo13WtupPF",
	"MNFSTqtC93rEfYHB6hF82sKdZpUDFWkViLByLd1k1Ms",
	"FrLmhMwyYQJVivt2EsVZwEEVspxoVhYj22Cxxst4hWVf",
	"TessVdML9pBGgG9yGks7o4HewRaXVAMuoVj4x83GLQH",
	"ExA6GYhHAeRNMWVNLrDir1SKPJZcZA2oaPq6uriSmxfJ",
	"proVF4pMXVaYqmy4NjniPh4pqKNfMmsihgd4wdkCX3u",
	"61DFfeTKM7trxYcPQCM78bJ794ddZprZpAwAnLiwTpYH",
	"NinafKYvKDCH26v6uEpfDjyuDjjpbdfhiPjrJV6FTFs",
}

type swapProvider struct{}

func newSwapProvider() *swapProvider { return &swapProvider{} }

func (p *swapProvider) Name() ProviderName { return ProviderSwap }

func (p *swapProvider) Parse(ctx *ParseContext, result *UnifiedResult) *ProviderStatus {
	status := &ProviderStatus{}
	start := time.Now()
	status.Ran = true
	defer func() {
		status.Elapsed = time.Since(start).String()
		if rec := recover(); rec != nil {
			status.Error = fmt.Sprintf("panic: %v", rec)
		}
	}()

	for _, pidStr := range swapProtocols {
		pid := solana.MustPublicKeyFromBase58(pidStr)
		for _, cInst := range getClassifiedInstructions(ctx.Tx, ctx.Meta, pid) {
			if trade := getTransferSwapTrade(ctx, cInst); trade != nil {
				trade.Layer = ProviderSwap
				trade.AMM = ResolveAMM(pidStr)
				result.Trades = append(result.Trades, *trade)
			}
		}
	}

	return status
}

// ---------------------------------------------------------------------------
// BinaryReader
// ---------------------------------------------------------------------------

type BinaryReader struct {
	buffer []byte
	offset int
}

func NewBinaryReader(buffer []byte) *BinaryReader {
	return &BinaryReader{buffer: buffer, offset: 0}
}

func (r *BinaryReader) checkBounds(length int) {
	if r.offset+length > len(r.buffer) {
		panic(fmt.Sprintf("buffer overflow: read %d at offset %d, len %d",
			length, r.offset, len(r.buffer)))
	}
}

func (r *BinaryReader) ReadBytes(n int) []byte {
	r.checkBounds(n)
	v := r.buffer[r.offset : r.offset+n]
	r.offset += n
	return v
}

func (r *BinaryReader) ReadU128() *big.Int {
	data := r.ReadBytes(16)
	rev := make([]byte, 16)
	for i := 0; i < 16; i++ {
		rev[i] = data[15-i]
	}
	return new(big.Int).SetBytes(rev)
}

func (r *BinaryReader) ReadPubkey() solana.PublicKey {
	return solana.PublicKeyFromBytes(r.ReadBytes(32))
}

// ---------------------------------------------------------------------------
// classifiedInstruction
// ---------------------------------------------------------------------------

type classifiedInstruction struct {
	ix         solana.CompiledInstruction
	programID  solana.PublicKey
	outerIndex int
	innerIndex *int
}

func getIdxString(c classifiedInstruction) string {
	if c.innerIndex != nil {
		return fmt.Sprintf("%d.%d", c.outerIndex, *c.innerIndex)
	}
	return fmt.Sprintf("%d", c.outerIndex)
}

func getClassifiedInstructions(tx *solana.Transaction, meta *rpc.TransactionMeta, programID solana.PublicKey) []classifiedInstruction {
	var out []classifiedInstruction
	for i, ix := range tx.Message.Instructions {
		if int(ix.ProgramIDIndex) < len(tx.Message.AccountKeys) &&
			tx.Message.AccountKeys[ix.ProgramIDIndex].Equals(programID) {
			out = append(out, classifiedInstruction{
				ix: ix, programID: programID, outerIndex: i,
			})
		}
	}
	if meta != nil {
		for _, set := range meta.InnerInstructions {
			for j, ci := range set.Instructions {
				if int(ci.ProgramIDIndex) < len(tx.Message.AccountKeys) &&
					tx.Message.AccountKeys[ci.ProgramIDIndex].Equals(programID) {
					innerIdx := j
					out = append(out, classifiedInstruction{
						ix: solana.CompiledInstruction{
							ProgramIDIndex: ci.ProgramIDIndex,
							Accounts:       ci.Accounts,
							Data:           ci.Data,
						},
						programID:  programID,
						outerIndex: int(set.Index),
						innerIndex: &innerIdx,
					})
				}
			}
		}
	}
	return out
}

// ---------------------------------------------------------------------------
// Token balance resolution
// ---------------------------------------------------------------------------

type tokenAccountInfo struct {
	mint     solana.PublicKey
	decimals uint8
}

func buildTokenAccountMap(balances []rpc.TokenBalance) map[uint16]tokenAccountInfo {
	m := make(map[uint16]tokenAccountInfo, len(balances))
	for _, tb := range balances {
		dec := uint8(0)
		if tb.UiTokenAmount != nil {
			dec = tb.UiTokenAmount.Decimals
		}
		m[tb.AccountIndex] = tokenAccountInfo{mint: tb.Mint, decimals: dec}
	}
	return m
}

// ---------------------------------------------------------------------------
// Transfer-based swap detection
// ---------------------------------------------------------------------------

var (
	splTokenPID     = solana.MustPublicKeyFromBase58("TokenkegQfeZyiNwAJbNbGKPFXCWuBvf9Ss623VQ5DA")
	splToken2022PID = solana.MustPublicKeyFromBase58("TokenzQdBNbLqP5VEhdkAS6EPFLC1PHnBqCXEpPxuEb")
	wrappedSOL      = solana.MustPublicKeyFromBase58("So11111111111111111111111111111111111111112")
)

type transferEntry struct {
	mint     solana.PublicKey
	amount   uint64
	decimals uint8
	from     solana.PublicKey
	to       solana.PublicKey
}

func getTransfersForInstruction(tx *solana.Transaction, meta *rpc.TransactionMeta, outerIndex int) []transferEntry {
	if meta == nil {
		return nil
	}

	preMap := buildTokenAccountMap(meta.PreTokenBalances)

	var out []transferEntry
	for _, set := range meta.InnerInstructions {
		if int(set.Index) != outerIndex {
			continue
		}
		for _, ci := range set.Instructions {
			if int(ci.ProgramIDIndex) >= len(tx.Message.AccountKeys) {
				continue
			}
			pid := tx.Message.AccountKeys[ci.ProgramIDIndex]
			if !pid.Equals(splTokenPID) && !pid.Equals(splToken2022PID) {
				continue
			}
			if len(ci.Data) == 0 {
				continue
			}

			switch ci.Data[0] {
			case 3: // transfer
				if len(ci.Data) < 9 || len(ci.Accounts) < 3 {
					continue
				}
				amount := binary.LittleEndian.Uint64(ci.Data[1:9])
				from := resolveAccountKey(tx, ci.Accounts[0])
				to := resolveAccountKey(tx, ci.Accounts[1])
				mint, dec := resolveMint(ci.Accounts[0], ci.Accounts[1], preMap)
				out = append(out, transferEntry{mint: mint, amount: amount, decimals: dec, from: from, to: to})

			case 12: // transferChecked
				if len(ci.Data) < 10 || len(ci.Accounts) < 4 {
					continue
				}
				amount := binary.LittleEndian.Uint64(ci.Data[1:9])
				dec := ci.Data[9]
				from := resolveAccountKey(tx, ci.Accounts[0])
				mint := resolveAccountKey(tx, ci.Accounts[1])
				to := resolveAccountKey(tx, ci.Accounts[2])
				out = append(out, transferEntry{mint: mint, amount: amount, decimals: dec, from: from, to: to})
			}
		}
	}
	return out
}

func resolveAccountKey(tx *solana.Transaction, idx uint16) solana.PublicKey {
	if int(idx) < len(tx.Message.AccountKeys) {
		return tx.Message.AccountKeys[idx]
	}
	return solana.PublicKey{}
}

func resolveMint(fromIdx, toIdx uint16, preMap map[uint16]tokenAccountInfo) (solana.PublicKey, uint8) {
	if info, ok := preMap[fromIdx]; ok && !info.mint.IsZero() {
		return info.mint, info.decimals
	}
	if info, ok := preMap[toIdx]; ok && !info.mint.IsZero() {
		return info.mint, info.decimals
	}
	return solana.PublicKey{}, 0
}

func processSwapData(transfers []transferEntry, signer solana.PublicKey) (TokenAmount, TokenAmount, bool) {
	type mintInfo struct {
		mint     solana.PublicKey
		amount   uint64
		decimals uint8
		source   solana.PublicKey
	}
	seen := make(map[solana.PublicKey]mintInfo)
	for _, t := range transfers {
		if t.mint.Equals(wrappedSOL) || t.mint.IsZero() {
			continue
		}
		if existing, ok := seen[t.mint]; !ok || t.amount > existing.amount {
			seen[t.mint] = mintInfo{mint: t.mint, amount: t.amount, decimals: t.decimals, source: t.from}
		}
	}
	if len(seen) < 2 {
		return TokenAmount{}, TokenAmount{}, false
	}

	unique := make([]mintInfo, 0, len(seen))
	for _, info := range seen {
		unique = append(unique, info)
	}

	input := unique[0]
	output := unique[len(unique)-1]

	if output.source.Equals(signer) {
		input, output = output, input
	}

	return TokenAmount{
			Mint:     input.mint.String(),
			Amount:   fmt.Sprintf("%d", input.amount),
			UIAmount: float64(input.amount) / math.Pow(10, float64(input.decimals)),
			Decimals: input.decimals,
		}, TokenAmount{
			Mint:     output.mint.String(),
			Amount:   fmt.Sprintf("%d", output.amount),
			UIAmount: float64(output.amount) / math.Pow(10, float64(output.decimals)),
			Decimals: output.decimals,
		}, true
}

func getTransferSwapTrade(ctx *ParseContext, cInst classifiedInstruction) *TradeEvent {
	signer := ctx.Tx.Message.AccountKeys[0]

	transfers := getTransfersForInstruction(ctx.Tx, ctx.Meta, cInst.outerIndex)
	if len(transfers) < 2 {
		return nil
	}

	inputToken, outputToken, ok := processSwapData(transfers, signer)
	if !ok {
		return nil
	}

	return &TradeEvent{
		Layer:       "",
		Idx:         getIdxString(cInst),
		Type:        "SWAP",
		User:        signer.String(),
		InputToken:  inputToken,
		OutputToken: outputToken,
		ProgramID:   cInst.programID.String(),
	}
}
