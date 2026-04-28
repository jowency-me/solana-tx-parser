package solanatxparser_test

import (
	"path/filepath"
	"strings"
	"testing"

	solanatxparser "github.com/jowency-me/solana-tx-parser"
)

// fixturePaths returns all fixture JSON files in testdata/transactions.
func fixturePaths() ([]string, error) {
	return filepath.Glob("testdata/transactions/*.json")
}

// mustLoadFixture is a test helper that loads a fixture by base name.
func mustLoadFixture(t *testing.T, name string) (*fixture, *solanatxparser.UnifiedResult) {
	t.Helper()
	path := filepath.Join("testdata", "transactions", name+".json")
	f, raw, tx, err := loadFixture(path)
	if err != nil {
		t.Fatalf("load fixture %s: %v", name, err)
	}
	parser := solanatxparser.NewParser(nil, solanatxparser.DefaultOptions())
	res, err := parser.ParseDecoded(f.Signature, raw, tx)
	if err != nil {
		t.Fatalf("parse fixture %s: %v", name, err)
	}
	return f, res
}

// TestFixtureLoader verifies that every saved fixture can round-trip through
// JSON back into a parseable transaction.
func TestFixtureLoader(t *testing.T) {
	paths, err := fixturePaths()
	if err != nil {
		t.Fatalf("glob: %v", err)
	}
	if len(paths) == 0 {
		t.Fatal("no fixtures found in testdata/transactions")
	}

	for _, p := range paths {
		t.Run(filepath.Base(p), func(t *testing.T) {
			f, raw, tx, err := loadFixture(p)
			if err != nil {
				t.Fatalf("load: %v", err)
			}
			if f.Signature == "" {
				t.Fatal("missing signature")
			}
			if raw.Slot == 0 {
				t.Fatal("missing slot")
			}
			if tx == nil {
				t.Fatal("nil transaction")
			}
			if len(tx.Signatures) == 0 {
				t.Fatal("no signatures")
			}
		})
	}
}

// TestParseAllFixtures parses every fixture offline.
// Any transaction saved to testdata must be fully recognized: not only the
// event type, but also the key details (who, what token, how much, sender,
// receiver, etc.) must be present.
func TestParseAllFixtures(t *testing.T) {
	paths, err := fixturePaths()
	if err != nil {
		t.Fatalf("glob: %v", err)
	}

	parser := solanatxparser.NewParser(nil, solanatxparser.DefaultOptions())

	for _, p := range paths {
		name := strings.TrimSuffix(filepath.Base(p), ".json")
		t.Run(name, func(t *testing.T) {
			f, raw, tx, err := loadFixture(p)
			if err != nil {
				t.Fatalf("load: %v", err)
			}

			res, err := parser.ParseDecoded(f.Signature, raw, tx)
			if err != nil {
				t.Fatalf("parse: %v", err)
			}

			if !res.HasAnyEvent() {
				t.Fatalf("%s: no events detected (trades=%d liquidities=%d transfers=%d meme=%d stake=%d)",
					f.Description,
					len(res.Trades), len(res.Liquidities), len(res.Transfers),
					len(res.MemeEvents), len(res.StakeEvents))
			}

			for i, tr := range res.Trades {
				if tr.User == "" {
					t.Errorf("trade #%d: missing User", i)
				}
				if tr.ProgramID == "" {
					t.Errorf("trade #%d: missing ProgramID", i)
				}
				if tr.InputToken.Mint == "" {
					t.Errorf("trade #%d: missing InputToken.Mint", i)
				}
				if tr.InputToken.Amount == "" {
					t.Errorf("trade #%d: missing InputToken.Amount", i)
				}
				if tr.OutputToken.Mint == "" {
					t.Errorf("trade #%d: missing OutputToken.Mint", i)
				}
				if tr.OutputToken.Amount == "" {
					t.Errorf("trade #%d: missing OutputToken.Amount", i)
				}
			}

			for i, liq := range res.Liquidities {
				if liq.User == "" {
					t.Errorf("liquidity #%d: missing User", i)
				}
				if liq.Action == "" {
					t.Errorf("liquidity #%d: missing Action", i)
				}
				if liq.BaseToken.Mint == "" {
					t.Errorf("liquidity #%d: missing BaseToken.Mint", i)
				}
				if liq.QuoteToken.Mint == "" {
					t.Errorf("liquidity #%d: missing QuoteToken.Mint", i)
				}
			}

			for i, tf := range res.Transfers {
				if tf.From == "" {
					t.Errorf("transfer #%d: missing From", i)
				}
				if tf.To == "" {
					t.Errorf("transfer #%d: missing To", i)
				}
				if tf.Token.Mint == "" {
					t.Errorf("transfer #%d: missing Token.Mint", i)
				}
				if tf.Token.Amount == "" {
					t.Errorf("transfer #%d: missing Token.Amount", i)
				}
			}

			for i, m := range res.MemeEvents {
				if m.Protocol == "" {
					t.Errorf("meme #%d: missing Protocol", i)
				}
				if m.Action == "" {
					t.Errorf("meme #%d: missing Action", i)
				}
				if m.BaseMint == "" {
					t.Errorf("meme #%d: missing BaseMint", i)
				}
			}

			for i, s := range res.StakeEvents {
				if s.Instruction == "" {
					t.Errorf("stake #%d: missing Instruction", i)
				}
			}
		})
	}
}

// TestParseDedupe verifies deduplication for the dual-parser fixture.
func TestParseDedupe(t *testing.T) {
	_, res := mustLoadFixture(t, "dual_parser_trade")

	if len(res.Trades) == 0 {
		t.Fatal("expected at least 1 trade")
	}

	hasL1, hasL2 := false, false
	for _, tr := range res.Trades {
		if tr.Layer == solanatxparser.ProviderCxcx {
			hasL1 = true
		}
		if tr.Layer == solanatxparser.ProviderDefaultPerson {
			hasL2 = true
		}
	}
	if !hasL1 {
		t.Error("expected at least one cxcx trade")
	}
	if !hasL2 {
		t.Error("expected at least one defaultperson trade")
	}

	seen := map[string]bool{}
	for _, tr := range res.Trades {
		k := tr.ProgramID + "|" + tr.Idx + "|" + tr.AMM
		if seen[k] {
			t.Errorf("duplicate trade dedup key: %s", k)
		}
		seen[k] = true
	}
}
