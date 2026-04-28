# Solana TX Parser

A production-grade Solana transaction parser with **single RPC call** design and an **extensible parser registry**.

## Supported Transaction Types

| Type | Details |
|------|---------|
| **Swap (Trade)** | DEX swaps across 50+ protocols including Raydium V4/CPMM/CL/Launchpad, Orca Whirlpool, Meteora DLMM/DAMM v2, Pumpfun, Pumpswap, Bonk, Jupiter V6 routing, GoonFi, ZeroFi, 1Dex, FusionAMM |
| **Liquidity** | Add/remove liquidity on Meteora DLMM/DAMM v2, Raydium, Orca, and other supported DEXes |
| **Transfer** | SPL token transfers including self-transfers, fee transfers, and **burn detection** (SPL Token `Burn`/`BurnChecked` instructions, plus transfers to the incinerator address) |
| **Meme Token Events** | Token creation and launch events on Pumpfun and supported meme protocols |
| **Stake** | Native Stake program instructions: Initialize, DelegateStake, Withdraw, Deactivate, Merge, and 13 others |

### What Gets Extracted

Every parsed transaction returns:
- Signature, slot, block time, signer
- On-chain success/failure status, fee, compute units consumed
- **Trades**: input/output token (mint + amount), user address, AMM name, program ID
- **Liquidity**: action type, pool ID, base/quote token amounts
- **Transfers**: from/to addresses, token mint + amount, burn/fee/self flags
- **Meme events**: protocol, action, token name/symbol/mint
- **Stake events**: instruction type, stake account, vote account

## Supported Protocols

The parser recognizes transactions from the following programs. Program IDs are listed for verification and integration.

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
| **OKX Router** | `HV1KXxWFaSeriyFvXyx48FqG9BoFbfinB8njCJonqP7K` |
| **Raydium Route** | `routeUGWgWzqBWFcrCfv8tritsqukccJPu3q5GPP3xS` |
| **Sanctum** | `stkitrT1Uoy18Dk1fTrgPw8W6MVzoCfYoAFT4MLsmhq` |
| **Photon** | `BSfD6SHZigAfDWSjzD5Q41jw8LmKwtmjskPH9XW1mrRW` |

### Major DEX

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
| **Phoenix** | `PhoeNiXZ8ByJGLkxNfZRnkUfjvmuYqLR89jjFHGqdXY` |
| **Openbook** | `opnb2LAfJYbRMAHHvqjCwQxanZn7ReEHp1k81EohpZb` |
| **Meteora DLMM** | `LBUZKhRxPF3XUpBCjp4YzTKgLccjZhTSDM9YuVaPwxo` |
| **Meteora DAMM** | `Eo7WjKq67rjJQSZxS6z3YkapzY3eMj6Xy8X5EQVn5UaB` |
| **Meteora DAMM v2** | `cpamdpZCGKUy5JxQXB4dcpGPiikHawvSWAd6mEn1sGG` |
| **Meteora DBC** | `dbcij3LWUppWqq96dh6gJWwBifmcGfLSB5D4DuSMaqN` |
| **Serum V3** | `9xQeWvG816bUx9EPjHmaT23yvVM2ZWbrrpZb9PusVFin` |
| **Kamino Lending** | `KLend2g3cP87fffoy8q1mQqGKjrxjC8boSyAYavgmjD` |
| **Kamino Yield Vault** | `6LtLpnUFNByNXLyCoK9wA2MykKAmQNZKBdY8s47dehDc` |
| **Obri V2** | `obriQD1zbpyLz95G5n7nJe6a4DPjpFwa5XYPoNm113y` |

### Meme / Launchpad

| Protocol | Program ID |
|----------|------------|
| **Pumpfun** | `6EF8rrecthR5Dkzon8Nwu78hRvfCKubJ14M5uBEwF6P` |
| **Pumpswap** | `pAMMBay6oceH9fJKBRHGP5D4bD4sWpmSwMn52FMfXEA` |
| **Boopfun** | `boop8hVGQGqehUK2iVEMEnMrL5RbjywRzHKBmBE7ry4` |

### Other DEX

| Protocol | Program ID |
|----------|------------|
| **1Dex** | `DEXYosS6oEGvk8uCDayvwEZz4qEyDJRf9nFgYCaqPMTm` |
| **Aldrin** | `AMM55ShdkoGRB5jVYPjWziwk8m5MpwyDgsMWHaMSQWH6` |
| **Aldrin V2** | `CURVGoZn8zycx6FXwwevgBTB2gVvdbGTEpvMJDbgs2t4` |
| **Crema** | `CLMM9tUoggJu2wagPkkqs9eFG4BWhVBZWkP1qv3Sp7tR` |
| **GooseFX GAMMA** | `GAMMA7meSFWaBXF25oSUgmGRwaW6sCMFLmBNiMSdbHVT` |
| **GoonFi** | `goonuddtQRrWqqn5nFyczVKaie28f3kDkHWkHtURSLE` |
| **Heaven** | `HEAVENoP2qxoeuF8Dj2oT1GHEnu49U5mJYkdeC8BAX2o` |
| **Lifinity** | `EewxydAPCCVuNEyrVN68PuSYdQ7wKn27V9Gjeoi8dy3S` |
| **Lifinity V2** | `2wT8Yq49kHgDzXuPxZSaeLaH1qbmGXtEyPy64bL7aD3c` |
| **Mercurial** | `MERLuDFBMmsHnsBPZw2sDQZHvXFMwp8EdjudcU2HKky` |
| **Moonit** | `MoonCVVNZFSYkqNXP6bxHLPL6QQJiMagDL3qcqUQTrG` |
| **Saber** | `SSwpkEEcbUqx4vtoEByFjSkhKdCT862DNVb52nZg1UZ` |
| **Saros** | `SSwapUtytfBdBn1b9NUGG6foMVPtcWgpRU32HToDUZr` |
| **SolFi** | `SoLFiHG9TfgtdUXUjWAxi3LtvYuFyDLVhBWxdMZxyCe` |
| **Stabble** | `swapNyd8XiQwJ6ianp9snpu4brUqFxadzvHebnAXjJZ` |
| **Stabble Weight** | `swapFpHZwjELNnjvThjajtiVmkz3yPQEHjLtka2fwHW` |
| **Sugar** | `deus4Bvftd5QKcEkE5muQaWGWDoma8GrySvPFrBPjhS` |
| **ZeroFi** | `ZERor4xhbUycZ6gb9ntrhqscUcZmAbQDjEAtCf4hbZY` |
| **FusionAMM** | `fUSioN9YKKSa3CUC2YUc4tPkHJ5Y6XW1yz8y6F7qWz9` |

### Trading Bots

| Protocol | Program ID |
|----------|------------|
| **BananaGun** | `BANANAjs7FJiPQqJTGFzkZJndT9o7UmKiYYGaJz6frGu` |
| **Mintech** | `minTcHYRLVPubRK8nt6sqe2ZpWrGDLQoNLipDJCGocY` |
| **Bloom** | `b1oomGGqPKGD6errbyfbVMBuzSC8WtAAYo8MwNafWW1` |
| **Maestro** | `MaestroAAe9ge5HTc64VbBQZ6fP77pwvrhM8i1XWSAx` |
| **Nova** | `NoVA1TmDUqksaj2hB1nayFkPysjJbFiU76dT4qPw2wm` |
| **Apepro** | `JSW99DKmxNyREQM14SQLDykeBvEUG63TeohrvmofEiw` |

### Native Programs

| Program | Program ID | Parsed Events |
|---------|------------|---------------|
| **SPL Token** | `TokenkegQfeZyiNwAJbNbGKPFXCWuBvf9Ss623VQ5DA` | Transfers, Burn |
| **SPL Token 2022** | `TokenzQdBNbLqP5VEhdkAS6EPFLC1PHnBqCXEpPxuEb` | Transfers |
| **Native Stake** | `Stake11111111111111111111111111111111111111` | Delegate, Withdraw, Deactivate, Merge, etc. |

## Unsupported Protocols

The following protocols are not yet supported. Transactions involving only these programs will produce zero events:

| Protocol | Program ID | Type |
|----------|------------|------|
| **AlphaQ** | `ALPHAQmeA7bjrVuccPsYPiCvsi428SNwte66Srvs4pHA` | Swap |
| **FluxBeam** | `FLUXubRmkEi2q6K3Y9kBPg9248ggaZVsoSFhtJHSrm1X` | DEX |
| **REALQ** | `REALQqNEomY6cQGZJUGwywTBD2UmDT32rZcNnfxQ5N2` | Unidentified swap program |
| **T1TAN** | `T1TANpTeScyeqVzzgNViGDNrkQ6qHz9KrSBS4aNXvGT` | Unidentified |
| **SV2EYYJy** | `SV2EYYJyRz2YhfXwXnhNAevDEui5Q6yrfyo13WtupPF` | Unidentified |
| **MNFSTqtC** | `MNFSTqtC93rEfYHB6hF82sKdZpUDFWkViLByLd1k1Ms` | Unidentified |
| **FrLmh** | `FrLmhMwyYQJVivt2EsVZwEEVspxoVhYj22Cxxst4hWVf` | Unidentified |
| **TessVdML9** | `TessVdML9pBGgG9yGks7o4HewRaXVAMuoVj4x83GLQH` | Unidentified |
| **ExA6GYh** | `ExA6GYhHAeRNMWVNLrDir1SKPJZcZA2oaPq6uriSmxfJ` | Unidentified |
| **proVF4p** | `proVF4pMXVaYqmy4NjniPh4pqKNfMmsihgd4wdkCX3u` | Unidentified |
| **61DFfe** | `61DFfeTKM7trxYcPQCM78bJ794ddZprZpAwAnLiwTpYH` | Unidentified |
| **Nina** | `NinafKYvKDCH26v6EpdfhiPjrJV6FTFs` | Unidentified |

When Jupiter V6 routes through an unrecognized pool, the trade is detected but the target AMM may be reported as "Unknown".

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

Each event includes a `layer` field tracing which provider produced it.

## Design

1. **Single RPC call** — fetch base64-encoded transaction once; all parsers share pre-fetched data.
2. **Parser registry** — register parsers in priority order. First-registered wins on dedup.
3. **AMM normalization** — alias map + program-ID fallback ensures consistent naming across upstream libraries.
4. **Panic recovery** — each parser runs in its own recovery boundary.
5. **Burn detection** — SPL Token `Burn`/`BurnChecked` instructions and transfers to the incinerator address (`1nc1nerator11111111111111111111111111111111`) are flagged as `IsBurn` in unified post-processing.

## License

MIT
