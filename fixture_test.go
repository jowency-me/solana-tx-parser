package solanatxparser_test

import (
	"encoding/json"
	"fmt"
	"os"

	solana "github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/rpc"
)

type fixtureExpected struct {
	Trade        bool     `json:"trade"`
	Liquidity    bool     `json:"liquidity"`
	Meme         bool     `json:"meme"`
	SelfTransfer bool     `json:"selfTransfer"`
	Stake        []string `json:"stake"`
}

type fixture struct {
	Description string          `json:"description"`
	Category    string          `json:"category"`
	Signature   string          `json:"signature"`
	Slot        uint64          `json:"slot"`
	BlockTime   *int64          `json:"blockTime,omitempty"`
	ExplorerURL string          `json:"explorerURL"`
	Expected    fixtureExpected `json:"expected"`
	Raw         json.RawMessage `json:"raw"` // raw GetTransactionResult JSON
}

// loadFixture reads a fixture JSON file and returns the parsed components.
func loadFixture(path string) (*fixture, *rpc.GetTransactionResult, *solana.Transaction, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("read: %w", err)
	}

	var f fixture
	if err := json.Unmarshal(data, &f); err != nil {
		return nil, nil, nil, fmt.Errorf("unmarshal wrapper: %w", err)
	}

	var raw rpc.GetTransactionResult
	if err := json.Unmarshal(f.Raw, &raw); err != nil {
		return nil, nil, nil, fmt.Errorf("unmarshal raw tx: %w", err)
	}

	if raw.Transaction == nil {
		return nil, nil, nil, fmt.Errorf("raw tx missing transaction envelope")
	}

	tx, err := raw.Transaction.GetTransaction()
	if err != nil {
		return nil, nil, nil, fmt.Errorf("get transaction: %w", err)
	}

	return &f, &raw, tx, nil
}
