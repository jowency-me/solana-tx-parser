package solanatxparser

import (
	"fmt"
	"time"

	solana "github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/programs/stake"
	"github.com/gagliardetto/solana-go/rpc"
)

type stakeProvider struct{}

func newStakeProvider() *stakeProvider { return &stakeProvider{} }

func (p *stakeProvider) Name() ProviderName { return ProviderStake }

func (p *stakeProvider) Parse(ctx *ParseContext, result *UnifiedResult) *ProviderStatus {
	status := &ProviderStatus{}
	start := time.Now()
	status.Ran = true
	defer func() {
		status.Elapsed = time.Since(start).String()
		if rec := recover(); rec != nil {
			status.Error = fmt.Sprintf("panic: %v", rec)
			result.StakeEvents = nil
		}
	}()

	events, err := decodeStakeEvents(ctx.Tx, ctx.Meta)
	if err != nil {
		status.Error = err.Error()
		return status
	}
	result.StakeEvents = append(result.StakeEvents, events...)
	return status
}

func decodeStakeEvents(tx *solana.Transaction, meta *rpc.TransactionMeta) ([]StakeEvent, error) {
	if tx == nil {
		return nil, fmt.Errorf("nil transaction")
	}
	var out []StakeEvent

	for i, ix := range tx.Message.Instructions {
		if ev, ok := decodeStake(tx, ix, fmt.Sprintf("%d", i)); ok {
			out = append(out, ev)
		}
	}

	if meta != nil {
		for _, set := range meta.InnerInstructions {
			for j, ci := range set.Instructions {
				solIx := solana.CompiledInstruction{
					ProgramIDIndex: ci.ProgramIDIndex,
					Accounts:       ci.Accounts,
					Data:           ci.Data,
				}
				if ev, ok := decodeStake(tx, solIx, fmt.Sprintf("%d.%d", set.Index, j)); ok {
					out = append(out, ev)
				}
			}
		}
	}

	return out, nil
}

func decodeStake(tx *solana.Transaction, ix solana.CompiledInstruction, idx string) (ev StakeEvent, ok bool) {
	defer func() {
		if rec := recover(); rec != nil {
			ok = false
		}
	}()

	if int(ix.ProgramIDIndex) >= len(tx.Message.AccountKeys) {
		return StakeEvent{}, false
	}
	programID := tx.Message.AccountKeys[ix.ProgramIDIndex]
	if !programID.Equals(solana.StakeProgramID) {
		return StakeEvent{}, false
	}

	accounts := make([]*solana.AccountMeta, 0, len(ix.Accounts))
	accountKeys := make([]solana.PublicKey, 0, len(ix.Accounts))
	for _, acIdx := range ix.Accounts {
		if int(acIdx) >= len(tx.Message.AccountKeys) {
			continue
		}
		k := tx.Message.AccountKeys[acIdx]
		accounts = append(accounts, &solana.AccountMeta{PublicKey: k})
		accountKeys = append(accountKeys, k)
	}

	inst, err := stake.DecodeInstruction(accounts, ix.Data)
	if err != nil {
		return StakeEvent{}, false
	}

	ev = StakeEvent{
		Layer:       ProviderStake,
		Idx:         idx,
		Instruction: stakeInstructionName(inst),
	}
	if len(accountKeys) > 0 {
		ev.StakeAccount = accountKeys[0].String()
	}
	if len(accountKeys) > 1 {
		switch inst.TypeID.Uint32() {
		case stake.Instruction_DelegateStake, stake.Instruction_Redelegate:
			ev.VoteAccount = accountKeys[1].String()
		}
	}
	return ev, true
}

func stakeInstructionName(inst *stake.Instruction) string {
	switch inst.TypeID.Uint32() {
	case stake.Instruction_Initialize:
		return "Initialize"
	case stake.Instruction_Authorize:
		return "Authorize"
	case stake.Instruction_DelegateStake:
		return "DelegateStake"
	case stake.Instruction_Split:
		return "Split"
	case stake.Instruction_Withdraw:
		return "Withdraw"
	case stake.Instruction_Deactivate:
		return "Deactivate"
	case stake.Instruction_SetLockup:
		return "SetLockup"
	case stake.Instruction_Merge:
		return "Merge"
	case stake.Instruction_AuthorizeWithSeed:
		return "AuthorizeWithSeed"
	case stake.Instruction_InitializeChecked:
		return "InitializeChecked"
	case stake.Instruction_AuthorizeChecked:
		return "AuthorizeChecked"
	case stake.Instruction_AuthorizeCheckedWithSeed:
		return "AuthorizeCheckedWithSeed"
	case stake.Instruction_SetLockupChecked:
		return "SetLockupChecked"
	case stake.Instruction_GetMinimumDelegation:
		return "GetMinimumDelegation"
	case stake.Instruction_DeactivateDelinquent:
		return "DeactivateDelinquent"
	case stake.Instruction_Redelegate:
		return "Redelegate"
	case stake.Instruction_MoveStake:
		return "MoveStake"
	case stake.Instruction_MoveLamports:
		return "MoveLamports"
	}
	return fmt.Sprintf("Unknown(%d)", inst.TypeID.Uint32())
}
