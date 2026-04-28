# Solana TX Parser

A production-grade Solana transaction parser with **single RPC call** design and an **extensible parser registry**.

## Supported Transaction Types

| Type | Details |
|------|---------|
| **Swap (Trade)** | DEX swaps across 60+ protocols including Raydium V4/CPMM/CL/Launchpad, Orca Whirlpool/V1/V2, Meteora DLMM/DAMM/DAMM v2, Pumpfun, Pumpswap, Jupiter V2/V4/V6 routing, and more |
| **Liquidity** | Add/remove liquidity on Meteora DLMM/DAMM v2, Raydium, and other supported DEXes |
| **Transfer** | SPL token transfers including self-transfers, fee transfers, and **burn detection** (SPL Token `Burn`/`BurnChecked` instructions, plus transfers to the incinerator address) |
| **Meme Token Events** | Token creation and launch events on Pumpfun and supported meme protocols |
| **Stake** | Native Stake program instructions: Initialize, DelegateStake, Withdraw, Deactivate, Merge, and 13 others |
| **Bridge** | Cross-chain bridge transfers (e.g. Wormhole) |

### What Gets Extracted

Every parsed transaction returns:
- Signature, slot, block time, signer
- On-chain success/failure status, fee, compute units consumed
- **Trades**: input/output token (mint + amount), user address, AMM name, program ID
- **Liquidity**: action type, pool ID, base/quote token amounts
- **Transfers**: from/to addresses, token mint + amount, burn/fee/self flags
- **Meme events**: protocol, action, token name/symbol/mint
- **Stake events**: instruction type, stake account, vote account

## Parser Architecture

The parser combines four sources:

1. **cxcx** (`cxcx-ai/solana-parser-go`) — 50+ DEX protocols, Meme, Jupiter routing
2. **defaultperson** (`DefaultPerson/solana-dex-parser-go`) — DLMM / DAMM v2 fallback
3. **swap** — Local transfer-based swap detection for DEX programs not covered by upstream libraries
4. **stake** (`gagliardetto/solana-go/programs/stake`) — 18 Stake program instructions

Results are deduplicated across providers.

## Supported Protocols

All protocols below have been tested with **real on-chain transaction data** (fixtures in `testdata/transactions/`).

### DEX Aggregators

| Protocol | Program ID |
|----------|------------|
| **Jupiter V6** | `JUP6LkbZbjS1jKKwapdHNy74zcZ3tLUZoi5QNyVTaV4` |
| **Jupiter V4** | `JUP4Fb2cqiRUcaTHdrPC8h2gNsA2ETXiPDD33WcGuJB` |
| **Jupiter V2** | `JUP2jxvXaqu7NQY1GmNF4m1vodw12LVXYxbFL2uJvfo` |
| **Jupiter DCA** | `DCA265Vj8a9CEuX1eb1LWRnDT7uK6q1xMipnNyatn23M` |
| **Jupiter Limit** | `jupoNjAxXgZ4rjzxzPMP4oxduvQsQtZzyknqvzYNrNu` |
| **Jupiter Limit V2** | `j1o2qRpjcyUwEvwtcfhEQefh773ZgjxcVRry7LDqg5X` |
| **Jupiter VA** | `VALaaymxQh2mNy2trH9jUqHT1mTow76wpTcGmSWSwJe` |
| **OKX DEX** | `6m2CDdhRgxpH4WjvdzxAYbGxwdGUz5MziiL5jek2kBma` |
| **OKX Router** * | `HV1KXxWFaSeriyFvXyx48FqG9BoFfbinB8njCJonqP7K` |
| **Raydium Route** | `routeUGWgWzqBWFcrCfv8tritsqukccJPu3q5GPP3xS` |
| **Sanctum** | `stkitrT1Uoy18Dk1fTrgPw8W6MVzoCfYoAFT4MLsmhq` |
| **Photon** | `BSfD6SHZigAfDWSjzD5Q41jw8LmKwtmjskPH9XW1mrRW` |
| **DFlow** | `DF1ow4tspfHX9JwWJsAb9epbkA8hmpSEAtxXy1V27QBH` |

### DEX

| Protocol | Program ID |
|----------|------------|
| **Raydium V4** | `675kPX9MHTjS2zt1qfr1NYHuzeLXfQM9H24wFSUt1Mp8` |
| **Raydium AMM** | `5quBtoiQqxF9Jv6KYKctB59NT3gtJD2Y65kdnB1Uev3h` |
| **Raydium CPMM** | `CPMMoo8L3F4NbTegBCKVNunggL7H1ZpdTHKxQB5qKP1C` |
| **Raydium CL** | `CAMMCzo5YL8w4VFF8KVHrK22GGUsp5VTaW7grrKgrWqK` |
| **Raydium Launchpad** | `LanMV9sAd7wArD4vJFi2qDdfnVhFxYSUg6eADduJ3uj` |
| **Orca Whirlpool** | `whirLbMiicVdio4qvUfM5KAg6Ct8VwpYzGff3uctyCc` |
| **Orca V2** | `9W959DqEETiGZocYWCQPaJ6sBmUzgfxXfqGeTEdp3aQP` |
| **Orca V1** | `DjVE6JNiYqPL2QXyCUUh8rNjHrbz9hXHNYt99MQ59qw1` |
| **Meteora DLMM** | `LBUZKhRxPF3XUpBCjp4YzTKgLccjZhTSDM9YuVaPwxo` |
| **Meteora DAMM** | `Eo7WjKq67rjJQSZxS6z3YkapzY3eMj6Xy8X5EQVn5UaB` |
| **Meteora DAMM v2** | `cpamdpZCGKUy5JxQXB4dcpGPiikHawvSWAd6mEn1sGG` |
| **Meteora DBC** * | `dbcij3LWUppWqq96dh6gJWwBifmcGfLSB5D4DuSMaqN` |
| **MeteoraVault** | `24Uqj9JCLxUeoC3hGfh5W3s9FM9uCHDS2SG3LYwBpyTi` |
| **Pumpfun** | `6EF8rrecthR5Dkzon8Nwu78hRvfCKubJ14M5uBEwF6P` |
| **Pumpswap** | `pAMMBay6oceH9fJKBRHGP5D4bD4sWpmSwMn52FMfXEA` |
| **Phoenix** | `PhoeNiXZ8ByJGLkxNfZRnkUfjvmuYqLR89jjFHGqdXY` |
| **Serum V3** | `9xQeWvG816bUx9EPjHmaT23yvVM2ZWbrrpZb9PusVFin` |
| **Openbook** | `opnb2LAfJYbRMAHHvqjCwQxanZn7ReEHp1k81EohpZb` |
| **Kamino Lending** | `KLend2g3cP87fffoy8q1mQqGKjrxjC8boSyAYavgmjD` |
| **Kamino Yield Vault** | `6LtLpnUFNByNXLyCoK9wA2MykKAmQNZKBdY8s47dehDc` |
| **1Dex** | `DEXYosS6oEGvk8uCDayvwEZz4qEyDJRf9nFgYCaqPMTm` |
| **Aldrin** | `AMM55ShdkoGRB5jVYPjWziwk8m5MpwyDgsMWHaMSQWH6` |
| **Aldrin V2** | `CURVGoZn8zycx6FXwwevgBTB2gVvdbGTEpvMJDbgs2t4` |
| **GooseFX GAMMA** | `GAMMA7meSFWaBXF25oSUgmGRwaW6sCMFLmBNiMSdbHVT` |
| **Heaven** | `HEAVENoP2qxoeuF8Dj2oT1GHEnu49U5mJYkdeC8BAX2o` |
| **Lifinity V2** | `2wT8Yq49kHgDzXuPxZSaeLaH1qbmGXtEyPy64bL7aD3c` |
| **Lifinity** | `EewxydAPCCVuNEyrVN68PuSYdQ7wKn27V9Gjeoi8dy3S` |
| **Mercurial** | `MERLuDFBMmsHnsBPZw2sDQZHvXFMwp8EdjudcU2HKky` |
| **Moonit** | `MoonCVVNZFSYkqNXP6bxHLPL6QQJiMagDL3qcqUQTrG` |
| **Saber** | `SSwpkEEcbUqx4vtoEByFjSkhKdCT862DNVb52nZg1UZ` |
| **Saros** | `SSwapUtytfBdBn1b9NUGG6foMVPtcWgpRU32HToDUZr` |
| **Stabble** | `swapNyd8XiQwJ6ianp9snpu4brUqFxadzvHebnAXjJZ` |
| **StabbleVault** | `vo1tWgqZMjG61Z2T9qUaMYKqZ75CYzMuaZ2LZP1n7HV` |
| **Stabble Weight** | `swapFpHZwjELNjhvThjajtiVmkz3yPQEHjLtka2fwHW` |
| **Sugar** | `deus4Bvftd5QKcEkE5muQaWGWDoma8GrySvPFrBPjhS` |
| **FusionAMM** | `fUSioN9YKKSa3CUC2YUc4tPkHJ5Y6XW1yz8y6F7qWz9` |
| **HumidiFi** | `9H6tua7jkLhdm3w8BvgpTn5LZNU7g4ZynDmCiNN3q6Rp` |
| **Boopfun** | `boop8hVGQGqehUK2iVEMEnMrL5RbjywRzHKBmBE7ry4` |
| **GoonFi** | `goonuddtQRrWqqn5nFyczVKaie28f3kDkHWkHtURSLE` |
| **GoonFi (DP)** | `goonERTdGsjnkZqWuVjs73BZ3Pb9qoCUdBUL17BnS5j` |
| **SolFi** | `SoLFiHG9TfgtdUXUjWAxi3LtvYuFyDLVhBWxdMZxyCe` |
| **ZeroFi** * | `ZERor4xhbUycZ6gb9ntrhqscUcZmAbQDjEAtCf4hbZY` |
| **Obri V2** * | `obriQD1zbpyLz95G5n7nJe6a4DPjpFwa5XYPoNm113y` |
| **AlphaQ** | `ALPHAQmeA7bjrVuccPsYPiCvsi428SNwte66Srvs4pHA` |
| **FluxBeam** | `FLUXubRmkEi2q6K3Y9kBPg9248ggaZVsoSFhtJHSrm1X` |
| **Byreal** | `REALQqNEomY6cQGZJUGwywTBD2UmDT32rZcNnfxQ5N2` |
| **Titan** | `T1TANpTeScyeqVzzgNViGDNrkQ6qHz9KrSBS4aNXvGT` |
| **SolFiV2** | `SV2EYYJyRz2YhfXwXnhNAevDEui5Q6yrfyo13WtupPF` |
| **Manifest** | `MNFSTqtC93rEfYHB6hF82sKdZpUDFWkViLByLd1k1Ms` |
| **FrLmh** | `FrLmhMwyYQJVivt2EsVZwEEVspxoVhYj22Cxxst4hWVf` |
| **TesseraV** | `TessVdML9pBGgG9yGks7o4HewRaXVAMuoVj4x83GLQH` |
| **ExA6GYh** | `ExA6GYhHAeRNMWVNLrDir1SKPJZcZA2oaPq6uriSmxfJ` |
| **OKXLabs2** | `proVF4pMXVaYqmy4NjniPh4pqKNfMmsihgd4wdkCX3u` |
| **JupiterOrderEngine** | `61DFfeTKM7trxYcPQCM78bJ794ddZprZpAwAnLiwTpYH` |
| **Nina** | `NinafKYvKDCH26v6uEpfDjyuDjjpbdfhiPjrJV6FTFs` |

\* Resolved via AMM name mapping, tested indirectly through Jupiter-routed transactions.

### Trading Bots

| Protocol | Program ID |
|----------|------------|
| **BananaGun** | `BANANAjs7FJiPQqJTGFzkZJndT9o7UmKiYYGaJz6frGu` |
| **Mintech** | `minTcHYRLVPubRK8nt6sqe2ZpWrGDLQoNLipDJCGocY` |
| **Bloom** | `b1oomGGqPKGD6errbyfbVMBuzSC8WtAAYo8MwNafWW1` |
| **Maestro** | `MaestroAAe9ge5HTc64VbBQZ6fP77pwvrhM8i1XWSAx` |
| **Apepro** | `JSW99DKmxNyREQM14SQLDykeBvEUG63TeohrvmofEiw` |
| **Nova** | `NoVA1TmDUqksaj2hB1nayFkPysjJbFiU76dT4qPw2wm` |

### Native Programs

| Program | Program ID | Parsed Events |
|---------|------------|---------------|
| **SPL Token** | `TokenkegQfeZyiNwAJbNbGKPFXCWuBvf9Ss623VQ5DA` | Transfers, Burn |
| **SPL Token 2022** | `TokenzQdBNbLqP5VEhdkAS6EPFLC1PHnBqCXEpPxuEb` | Transfers, Burn |
| **Native Stake** | `Stake11111111111111111111111111111111111111` | Delegate, Withdraw, Deactivate, Merge, etc. |
| **Wormhole** | `worm2ZoG2kUd4vFXhvjh93UUH596ayRfgQ2MgjNMTth` | Bridge transfers |

## Unsupported Protocols

The following protocols are listed in upstream libraries but could **not be verified** with real on-chain swap transactions.

| Protocol | Program ID | Type |
|----------|------------|------|
| **Crema** | `CLMM9tUoggJu2wagPkkqs9eFG4BWhVBZWkP1qv3Sp7tR` | CLMM |
| **HeavenStore** | `HEvSKofvBgfaexv23kMabbYqxasxU3mQ4ibBMEmJWHny` | Oracle |

## Installation

```bash
go get github.com/jowency-me/solana-tx-parser
```

## Usage

### Parse by Signature (with RPC)

```go
import (
    "context"
    "github.com/gagliardetto/solana-go/rpc"
    solanatxparser "github.com/jowency-me/solana-tx-parser"
)

client := rpc.New("https://solana-rpc.publicnode.com")
parser := solanatxparser.NewParser(client, solanatxparser.DefaultOptions())

res, err := parser.ParseSignature(context.Background(), "5zYnEBQ...")
if err != nil {
    panic(err)
}

fmt.Printf("Trades=%d Liquidities=%d Transfers=%d Stake=%d\n",
    len(res.Trades), len(res.Liquidities), len(res.Transfers), len(res.StakeEvents))
```

### Parse Pre-fetched Data (no RPC)

```go
res, err := parser.ParseDecoded(signature, rawTx, decodedTx)
```

### Custom Parsers

```go
type MyParser struct{}

func (p *MyParser) Name() solanatxparser.ProviderName { return "my-parser" }

func (p *MyParser) Parse(ctx *solanatxparser.ParseContext, result *solanatxparser.UnifiedResult) *solanatxparser.ProviderStatus {
    return &solanatxparser.ProviderStatus{Ran: true}
}

parser.Register(&MyParser{})
```

## Output Structure

```go
type UnifiedResult struct {
    Signature      string
    Slot           uint64
    BlockTime      *int64
    Signer         string
    OnchainSuccess bool
    OnchainError   any
    Fee            uint64
    ComputeUnits   uint64

    Trades      []TradeEvent
    Liquidities []LiquidityEvent
    Transfers   []TransferEvent
    MemeEvents  []MemeEvent
    StakeEvents []StakeEvent

    Providers map[ProviderName]*ProviderStatus
}
```

## Design

1. **Single RPC call** — fetch base64-encoded transaction once; all parsers share pre-fetched data.
2. **Parser registry** — register parsers in priority order. First-registered wins on dedup.
3. **AMM normalization** — alias map + program-ID fallback ensures consistent naming across upstream libraries.
4. **Panic recovery** — each parser runs in its own recovery boundary.
5. **Burn detection** — SPL Token `Burn`/`BurnChecked` instructions and transfers to the incinerator address (`1nc1nerator11111111111111111111111111111111`) are flagged as `IsBurn` in unified post-processing.

## License

MIT
