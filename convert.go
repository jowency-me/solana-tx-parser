package solanatxparser

import (
	"fmt"

	dpAdapter "github.com/DefaultPerson/solana-dex-parser-go/adapter"
	dpTypes "github.com/DefaultPerson/solana-dex-parser-go/types"
	solana "github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/rpc"
)

// convertToDPTransaction converts decoded solana.Transaction + rpc.TransactionMeta
// into the DefaultPerson adapter format, avoiding a second RPC call or JSON re-encoding.
func convertToDPTransaction(
	signature string,
	slot uint64,
	blockTime *int64,
	tx *solana.Transaction,
	meta *rpc.TransactionMeta,
	versionTag interface{},
) (*dpAdapter.SolanaTransaction, error) {
	if tx == nil {
		return nil, fmt.Errorf("nil transaction")
	}

	out := &dpAdapter.SolanaTransaction{
		Slot:      slot,
		BlockTime: blockTime,
		Version:   versionTag,
	}

	sigs := make([]string, 0, len(tx.Signatures))
	for _, s := range tx.Signatures {
		sigs = append(sigs, s.String())
	}
	if len(sigs) == 0 && signature != "" {
		sigs = append(sigs, signature)
	}
	out.Transaction.Signatures = sigs

	msg := tx.Message
	isV0 := msg.IsVersioned()

	staticKeys := make([]string, 0, len(msg.AccountKeys))
	for _, k := range msg.AccountKeys {
		staticKeys = append(staticKeys, k.String())
	}

	if isV0 {
		out.Transaction.Message.Header = &dpAdapter.MessageHeader{
			NumRequiredSignatures:       int(msg.Header.NumRequiredSignatures),
			NumReadonlySignedAccounts:   int(msg.Header.NumReadonlySignedAccounts),
			NumReadonlyUnsignedAccounts: int(msg.Header.NumReadonlyUnsignedAccounts),
		}
		out.Transaction.Message.StaticAccountKeys = staticKeys
		out.Transaction.Message.CompiledInstructions = make([]dpAdapter.CompiledInstruction, 0, len(msg.Instructions))
		for _, ix := range msg.Instructions {
			out.Transaction.Message.CompiledInstructions = append(
				out.Transaction.Message.CompiledInstructions,
				compiledInsToDP(ix),
			)
		}
	} else {
		akObjs := make([]dpAdapter.AccountKey, 0, len(staticKeys))
		signersN := int(msg.Header.NumRequiredSignatures)
		readonlySignedN := int(msg.Header.NumReadonlySignedAccounts)
		readonlyUnsignedN := int(msg.Header.NumReadonlyUnsignedAccounts)
		total := len(staticKeys)
		writableUnsignedEnd := total - readonlyUnsignedN
		for i, pk := range staticKeys {
			signer := i < signersN
			writable := false
			if signer {
				writable = i < (signersN - readonlySignedN)
			} else {
				writable = i < writableUnsignedEnd
			}
			akObjs = append(akObjs, dpAdapter.AccountKey{
				Pubkey: pk, Signer: signer, Writable: writable,
			})
		}
		out.Transaction.Message.AccountKeys = akObjs
		instructionsIfaces := make([]interface{}, 0, len(msg.Instructions))
		for _, ix := range msg.Instructions {
			instructionsIfaces = append(instructionsIfaces, compiledInsToDP(ix))
		}
		out.Transaction.Message.Instructions = instructionsIfaces
	}

	if len(msg.AddressTableLookups) > 0 {
		atl := make([]dpAdapter.AddressTableLookup, 0, len(msg.AddressTableLookups))
		for _, lk := range msg.AddressTableLookups {
			wi := make([]int, 0, len(lk.WritableIndexes))
			for _, x := range lk.WritableIndexes {
				wi = append(wi, int(x))
			}
			ri := make([]int, 0, len(lk.ReadonlyIndexes))
			for _, x := range lk.ReadonlyIndexes {
				ri = append(ri, int(x))
			}
			atl = append(atl, dpAdapter.AddressTableLookup{
				AccountKey:      lk.AccountKey.String(),
				WritableIndexes: wi,
				ReadonlyIndexes: ri,
			})
		}
		out.Transaction.Message.AddressTableLookups = atl
	}

	if meta == nil {
		return out, nil
	}

	dpMeta := &dpAdapter.TransactionMeta{
		Fee:          meta.Fee,
		PreBalances:  append([]uint64{}, meta.PreBalances...),
		PostBalances: append([]uint64{}, meta.PostBalances...),
		LogMessages:  append([]string{}, meta.LogMessages...),
	}
	if meta.Err != nil {
		dpMeta.Err = meta.Err
	}
	if meta.ComputeUnitsConsumed != nil {
		v := *meta.ComputeUnitsConsumed
		dpMeta.ComputeUnitsConsumed = &v
	}

	dpLA := &dpAdapter.LoadedAddresses{}
	for _, k := range meta.LoadedAddresses.Writable {
		dpLA.Writable = append(dpLA.Writable, k.String())
	}
	for _, k := range meta.LoadedAddresses.ReadOnly {
		dpLA.Readonly = append(dpLA.Readonly, k.String())
	}
	dpMeta.LoadedAddresses = dpLA

	dpMeta.PreTokenBalances = convertTokenBalances(meta.PreTokenBalances)
	dpMeta.PostTokenBalances = convertTokenBalances(meta.PostTokenBalances)

	if len(meta.InnerInstructions) > 0 {
		inner := make([]dpAdapter.InnerInstructionSet, 0, len(meta.InnerInstructions))
		for _, set := range meta.InnerInstructions {
			ifaces := make([]interface{}, 0, len(set.Instructions))
			for _, ci := range set.Instructions {
				ifaces = append(ifaces, dpAdapter.CompiledInstruction{
					ProgramIdIndex: int(ci.ProgramIDIndex),
					Accounts:       toIntSlice(ci.Accounts),
					Data:           ci.Data.String(),
				})
			}
			inner = append(inner, dpAdapter.InnerInstructionSet{
				Index:        int(set.Index),
				Instructions: ifaces,
			})
		}
		dpMeta.InnerInstructions = inner
	}
	out.Meta = dpMeta

	return out, nil
}

func compiledInsToDP(ix solana.CompiledInstruction) dpAdapter.CompiledInstruction {
	return dpAdapter.CompiledInstruction{
		ProgramIdIndex: int(ix.ProgramIDIndex),
		Accounts:       toIntSlice(ix.Accounts),
		Data:           ix.Data.String(),
	}
}

func toIntSlice(in []uint16) []int {
	out := make([]int, len(in))
	for i, v := range in {
		out[i] = int(v)
	}
	return out
}

func convertTokenBalances(in []rpc.TokenBalance) []dpAdapter.TokenBalance {
	if len(in) == 0 {
		return nil
	}
	out := make([]dpAdapter.TokenBalance, 0, len(in))
	for _, tb := range in {
		var ui *float64
		if tb.UiTokenAmount != nil && tb.UiTokenAmount.UiAmount != nil {
			v := *tb.UiTokenAmount.UiAmount
			ui = &v
		}
		amt := ""
		dec := uint8(0)
		if tb.UiTokenAmount != nil {
			amt = tb.UiTokenAmount.Amount
			dec = tb.UiTokenAmount.Decimals
		}
		owner := ""
		if tb.Owner != nil {
			owner = tb.Owner.String()
		}
		out = append(out, dpAdapter.TokenBalance{
			AccountIndex: int(tb.AccountIndex),
			Mint:         tb.Mint.String(),
			Owner:        owner,
			UiTokenAmount: dpTypes.TokenAmount{
				Amount:   amt,
				Decimals: dec,
				UIAmount: ui,
			},
		})
	}
	return out
}
