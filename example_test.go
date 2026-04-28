package solanatxparser_test

import (
	"fmt"
	"testing"

	solanatxparser "github.com/jowency-me/solana-tx-parser"
)

func TestExample(t *testing.T) {
	f, raw, tx, err := loadFixture("testdata/transactions/pumpfun_create.json")
	if err != nil {
		fmt.Println("error:", err)
		return
	}

	parser := solanatxparser.NewParser(nil, solanatxparser.DefaultOptions())
	res, err := parser.ParseDecoded(f.Signature, raw, tx)
	if err != nil {
		fmt.Println("error:", err)
		return
	}

	fmt.Printf("Signature: %s\n", res.Signature)
	fmt.Printf("Signer: %s\n", res.Signer)
	fmt.Printf("Slot: %d\n", res.Slot)
	fmt.Printf("Fee: %d lamports\n", res.Fee)
	fmt.Printf("ComputeUnits: %d\n", res.ComputeUnits)
	fmt.Printf("OnchainSuccess: %v\n", res.OnchainSuccess)

	fmt.Printf("\nTrades: %d\n", len(res.Trades))
	for i, tr := range res.Trades {
		fmt.Printf("  Trade #%d: %s via %s (program: %s)\n", i+1, tr.Type, tr.AMM, tr.ProgramID)
		fmt.Printf("    User: %s\n", tr.User)
		fmt.Printf("    Input:  %s (mint: %s, amount: %s)\n", tr.InputToken.Symbol, tr.InputToken.Mint, tr.InputToken.Amount)
		fmt.Printf("    Output: %s (mint: %s, amount: %s)\n", tr.OutputToken.Symbol, tr.OutputToken.Mint, tr.OutputToken.Amount)
	}

	fmt.Printf("\nLiquidities: %d\n", len(res.Liquidities))
	for i, liq := range res.Liquidities {
		fmt.Printf("  Liquidity #%d: %s on %s\n", i+1, liq.Action, liq.AMM)
		fmt.Printf("    User: %s | Pool: %s\n", liq.User, liq.PoolID)
		fmt.Printf("    Base:  %s (mint: %s)\n", liq.BaseToken.Amount, liq.BaseToken.Mint)
		fmt.Printf("    Quote: %s (mint: %s)\n", liq.QuoteToken.Amount, liq.QuoteToken.Mint)
	}

	fmt.Printf("\nTransfers: %d\n", len(res.Transfers))
	for i, tf := range res.Transfers {
		fmt.Printf("  Transfer #%d: %s -> %s\n", i+1, tf.From, tf.To)
		fmt.Printf("    Token: %s (mint: %s, amount: %s)\n", tf.Token.Symbol, tf.Token.Mint, tf.Token.Amount)
		fmt.Printf("    IsFee: %v | IsSelf: %v\n", tf.IsFee, tf.IsSelf)
	}

	fmt.Printf("\nMemeEvents: %d\n", len(res.MemeEvents))
	for i, m := range res.MemeEvents {
		fmt.Printf("  Meme #%d: %s on %s\n", i+1, m.Action, m.Protocol)
		fmt.Printf("    Name: %s | Symbol: %s | Mint: %s\n", m.Name, m.Symbol, m.BaseMint)
	}

	fmt.Printf("\nStakeEvents: %d\n", len(res.StakeEvents))
	for i, s := range res.StakeEvents {
		fmt.Printf("  Stake #%d: %s\n", i+1, s.Instruction)
		fmt.Printf("    StakeAccount: %s | VoteAccount: %s\n", s.StakeAccount, s.VoteAccount)
	}

	fmt.Printf("\nProvider Status:\n")
	for name, st := range res.Providers {
		fmt.Printf("  [%s] ran=%v elapsed=%s\n", name, st.Ran, st.Elapsed)
	}
}
