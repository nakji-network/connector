package chain

import (
	"context"
	"strings"
	"time"

	"github.com/nakji-network/connector/config"

	"github.com/avast/retry-go"
	bin "github.com/gagliardetto/binary"
	token_metadata "github.com/gagliardetto/metaplex-go/clients/token-metadata"
	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/programs/system"
	"github.com/gagliardetto/solana-go/programs/token"
	"github.com/gagliardetto/solana-go/rpc"
	"github.com/gagliardetto/solana-go/rpc/ws"
	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
)

const chain = "solana"

// Config yaml(local.yaml) file must include the following
//	rpcs:
//	  solana:
//		full:
//		  -wss://rpcprovider.com/
//		  -https://rpcprovider.com/

type (
	rpcClient = *rpc.Client
	wsClient  = *ws.Client
)

type SolanaClient struct {
	rpcClient
	wsClient
	Config *viper.Viper
}

func initConfig() *viper.Viper {
	conf := config.GetConfig()

	conf.SetDefault("solana.maxRetries", 10)
	conf.SetDefault("solana.retryDelay", 500)
	conf.SetDefault("solana.lowPrioRetryDelay", 3)
	conf.SetDefault("solana.lowPrioMaxRetries", 100)

	return conf
}

// TODO: The ws client is not being used. Need to look into it.
func (c Clients) Solana() *SolanaClient {
	conf := initConfig()

	rpcs := c.rpcMap[chain].Full

	var wsURL, httpURL string
	for _, u := range rpcs {
		if strings.HasPrefix(u, "ws") {
			wsURL = u
		} else {
			httpURL = u
		}
	}
	log.Info().
		Str("chain", chain).
		Str("ws url", wsURL).
		Str("http url", httpURL).
		Msg("connecting to RPC")

	rpcCli := rpc.New(httpURL)

	wsCli, err := ws.Connect(context.Background(), wsURL)
	if err != nil {
		log.Fatal().Err(err).Msg("Solana: WS Client connection error")
	}

	return &SolanaClient{rpcCli, wsCli, conf}
}

// ProgramListener listens for Signatures from the latest Solana block that mention the given `programId`
// and gets all Transaction data related for each of those signatures. `processTransaction`
// is where custom logic comes in for the specific program connector. This usually gets
// and decodes all instructions data from the transaction and emits/publishes events
// based on those decoded instructions.
// NOTE THAT IT IS NOT USING WS CLIENT !!!
func (sc *SolanaClient) ProgramListener(ctx context.Context, programId solana.PublicKey, namespace string, processTransaction func(*rpc.GetTransactionResult, *solana.Transaction, solana.Signature)) {
	var lastTxSig solana.Signature
	limit := 50
	opts := &rpc.GetSignaturesForAddressOpts{Limit: &limit}

	for {
		if !lastTxSig.IsZero() {
			opts.Until = lastTxSig
		}
		signatures, _ := sc.GetSignaturesForAddressWithOpts(context.Background(), programId, opts)

		if len(signatures) > 0 {
			lastTxSig = signatures[0].Signature
		}

		for _, txSignature := range signatures {
			signature := txSignature.Signature
			txResult, tx := sc.GetTransaction(ctx, signature)

			log.Debug().Msgf("%s : processing transaction %s", namespace, signature.String())
			processTransaction(txResult, tx, signature)
		}

		time.Sleep(time.Millisecond * 500)
	}
}

// ProgramBackfiller backfills all events/instructions using checkpointing. We use the last processed transaction's signature as
// the checkpoint. On next start of the program, we continue backfilling from that checkpoint. This also backfills
// until `backfillTxLimit` to backfill any missed transactions when the connector was down.
//func (sc *SolanaClient) ProgramBackfiller(ctx context.Context, programId solana.PublicKey, namespace string, processTransaction func(*rpc.GetTransactionResult, *solana.Transaction, solana.Signature), backfillLimit int) {
//	_, backfillCompleted := sc.getSignaturesForAddressOpts(namespace)
//
//	if backfillCompleted {
//		log.Info().Msgf("backfill already completed for %s", namespace)
//		return
//	}
//
//	var once sync.Once
//
//	go once.Do(func() {
//		log.Debug().Msgf("backfilling last %d transactions", backfillLimit)
//		opts := &rpc.GetSignaturesForAddressOpts{Limit: &backfillLimit}
//		sc.backfill(ctx, programId, namespace, processTransaction, false, opts)
//	})
//
//	for opts, backfillCompleted := sc.getSignaturesForAddressOpts(namespace); !backfillCompleted; {
//		sc.backfill(ctx, programId, namespace, processTransaction, true, opts)
//	}
//}
//
//func (sc *SolanaClient) backfill(ctx context.Context, programId solana.PublicKey, namespace string, processTransaction func(*rpc.GetTransactionResult, *solana.Transaction, solana.Signature), storeCheckpoint bool, opts *rpc.GetSignaturesForAddressOpts) {
//	var txSignatures []*rpc.TransactionSignature
//
//	for retries := 0; retries < 5 && len(txSignatures) == 0; retries++ {
//		var err error
//		txSignatures, err = sc.GetSignaturesForAddressWithOpts(ctx, programId, opts)
//		if err != nil {
//			log.Fatal().Err(err).Str("program", programId.String()).Msg("failed getting signatures for program in backfill")
//		}
//
//		if len(txSignatures) > 0 {
//			break
//		}
//
//		time.Sleep(500 * time.Millisecond)
//	}
//
//	// We want to end the loop when there are no signatures/transactions left
//	if len(txSignatures) == 0 {
//		sc.Db.UpdateBackfillCompleted(namespace, true)
//		return
//	}
//
//	for _, txSignature := range txSignatures {
//		signature := txSignature.Signature
//		log.Debug().Msgf("%s : backfilling transaction %s", namespace, signature.String())
//		txResult, tx := sc.GetTransaction(ctx, signature)
//		processTransaction(txResult, tx, signature)
//		if storeCheckpoint {
//			// We store the last signature we published events for so we can use as checkpoint later on.
//			sc.Db.UpsertLastTxSignatureQueried(namespace, signature.String())
//		}
//	}
//}
//
//func (sc *SolanaClient) getSignaturesForAddressOpts(namespace string) (*rpc.GetSignaturesForAddressOpts, bool) {
//	lastSignature := sc.Db.GetLastTxSignatureQueried(namespace)
//	limit := 500
//	opts := &rpc.GetSignaturesForAddressOpts{Limit: &limit}
//
//	if lastSignature != nil {
//		log.Info().Msgf("starting backfill from checkpoint %s for %s", lastSignature.Signature, namespace)
//		opts.Before = solana.MustSignatureFromBase58(lastSignature.Signature)
//		return opts, lastSignature.BackfillCompleted
//	}
//
//	return opts, false
//}

func (sc *SolanaClient) GetTransaction(ctx context.Context, signature solana.Signature) (*rpc.GetTransactionResult, *solana.Transaction) {
	opts := &rpc.GetTransactionOpts{Commitment: rpc.CommitmentFinalized, Encoding: solana.EncodingBase64}
	var txResult *rpc.GetTransactionResult

	err := sc.withRetry(
		func() error {
			var err error
			txResult, err = sc.rpcClient.GetTransaction(ctx, signature, opts)
			return err
		}, "GetTransaction", map[string]string{"txSignature": signature.String()},
	)

	if err != nil {
		log.Fatal().Err(err).Msgf("failed getting transaction: %s", signature)
	}

	decoder := bin.NewBinDecoder(txResult.Transaction.GetBinary())
	tx, err := solana.TransactionFromDecoder(decoder)
	if err != nil {
		log.Fatal().Err(err).Msg("failed decoding transaction")
	}

	return txResult, tx
}

func logHasError(logResult *ws.LogResult) bool {
	return HasError(logResult.Value.Err)
}

func HasError(err interface{}) bool {
	switch err.(type) {
	case map[string]interface{}:
		return true
	case nil:
		return false
	default:
		return false
	}
}

// Get and decode Mint data given a Mint address
func (sc *SolanaClient) GetMint(ctx context.Context, mintPubKey solana.PublicKey) *token.Mint {
	var mint token.Mint

	err := sc.withRetry(
		func() error {
			err := sc.GetAccountDataBorshInto(ctx, mintPubKey, &mint)
			return err
		}, "GetMint", map[string]string{"mint": mintPubKey.String()},
	)

	if err != nil {
		log.Fatal().Interface("mintPubKey", mintPubKey).Err(err).Msg("failed getting token mint account")
	}

	return &mint
}

// Get and decode Token Metadata given a Mint address. Only works for NFT Mints.
func (sc *SolanaClient) GetMetadata(ctx context.Context, mintPubKey solana.PublicKey) (*token_metadata.Metadata, error) {
	metaPubKey, _, err := solana.FindTokenMetadataAddress(mintPubKey)
	if err != nil {
		return nil, err
	}

	var meta token_metadata.Metadata
	err = sc.withRetry(
		func() error {
			err := sc.GetAccountDataBorshInto(ctx, metaPubKey, &meta)
			return err
		}, "GetMeta", map[string]string{"metadata": metaPubKey.String()},
		retry.Delay(sc.lowPrioRetryDelay()),
		retry.Attempts(sc.lowPrioRetryAttempts()),
	)

	if err != nil {
		return &meta, err
	}

	return &meta, nil
}

func (sc *SolanaClient) GetNFTTokenAccount(ctx context.Context, tokenMint solana.PublicKey) *token.Account {
	tokenAccountsResult, err := sc.GetTokenLargestAccounts(ctx, tokenMint, rpc.CommitmentFinalized)
	if err != nil {
		log.Fatal().Err(err).Str("mint", tokenMint.String()).Msg("failed getting associated token accounts")
	}

	address := tokenAccountsResult.Value[0].Address
	tokenAccount, err := sc.GetTokenAccount(ctx, address)
	if err != nil {
		log.Fatal().Err(err).Str("tokenAccount", address.String()).Msg("failed getting NFT token account")
	}

	return tokenAccount
}

func (sc *SolanaClient) GetTokenAccount(ctx context.Context, pubKey solana.PublicKey) (*token.Account, error) {
	var tokenAccount token.Account

	err := sc.withRetry(
		func() error {
			err := sc.GetAccountDataInto(ctx, pubKey, &tokenAccount)
			return err
		}, "GetTokenAccount", map[string]string{"tokenAccount": pubKey.String()},
		retry.Delay(sc.lowPrioRetryDelay()),
		retry.Attempts(sc.lowPrioRetryAttempts()),
	)

	return &tokenAccount, err
}

// Fetches and parses all SPL-Token Program instructions and inner instructions from all transaction data.
// The returned slice preserves the ordering of these Token Program instructions in the transaction.
func (sc *SolanaClient) GetTokenInstructions(txResult *rpc.GetTransactionResult, tx *solana.Transaction) []*token.Instruction {
	accountKeys := tx.Message.AccountKeys
	var (
		tokenProgramIdIdx uint16
		tokenInstructions []*token.Instruction
	)

	for i, publicKey := range accountKeys {
		if publicKey.String() == solana.TokenProgramID.String() {
			tokenProgramIdIdx = uint16(i)
			break
		}
	}

	// This map preserves the ordering of instructions and inner instructions within a transaction.
	// This is important since Solana uses that as basis for the order of their execution. The map's
	// key is the parent instruction's index within the transaction.
	innerInstructionsIndexMap := make(map[uint16]rpc.InnerInstruction)
	for _, innerInstruction := range txResult.Meta.InnerInstructions {
		innerInstructionsIndexMap[innerInstruction.Index] = innerInstruction
	}

	for i, instruction := range tx.Message.Instructions {
		// Check if the parent instruction is an SPL-Token Program instruction.
		// Add that to `tokenInstructions` slice, if it is.
		if instruction.ProgramIDIndex == tokenProgramIdIdx {
			tokenInstruction, _ := DecodeTokenInstruction(instruction, tx)
			tokenInstructions = append(tokenInstructions, tokenInstruction)
		}

		// If the inner instruction's index matches the current index, we check if any of these
		// inner instructions are SPL-Token Program instructions and add them to `tokenInstructions`.
		if innerInstruction, ok := innerInstructionsIndexMap[uint16(i)]; ok {
			for _, instruction := range innerInstruction.Instructions {
				if instruction.ProgramIDIndex == tokenProgramIdIdx {
					tokenInstruction, _ := DecodeTokenInstruction(instruction, tx)
					tokenInstructions = append(tokenInstructions, tokenInstruction)
				}
			}
		}
	}

	return tokenInstructions
}

// DecodeSystemInstruction decodes a `solana.CompiledInstruction` into a `*system.Instruction` which represents an Instruction
// from the System Program in Solana.
func DecodeSystemInstruction(instruction solana.CompiledInstruction, tx *solana.Transaction) (*system.Instruction, error) {
	accountMetas := instruction.ResolveInstructionAccounts(&tx.Message)
	systemInstruction, err := system.DecodeInstruction(accountMetas, instruction.Data)
	if err != nil {
		return nil, err
	}
	return systemInstruction, nil
}

// DecodeTokenInstruction decodes a `solana.CompiledInstruction` into a `*token.Instruction` which represents an Instruction
// from the SPL-Token Program in Solana.
func DecodeTokenInstruction(instruction solana.CompiledInstruction, tx *solana.Transaction) (*token.Instruction, error) {
	accountMetas := instruction.ResolveInstructionAccounts(&tx.Message)
	tokenInstruction, err := token.DecodeInstruction(accountMetas, instruction.Data)
	if err != nil {
		return nil, err
	}
	return tokenInstruction, nil
}

func (sc *SolanaClient) withRetry(retryFunc retry.RetryableFunc, funcName string, logs map[string]string, opts ...retry.Option) error {
	maxRetries := sc.Config.GetUint("solana.maxRetries")
	retryDelay := sc.Config.GetInt("solana.retryDelay")

	if len(opts) == 0 {
		opts = append(
			opts,
			retry.Delay(time.Duration(retryDelay)*time.Millisecond),
			retry.Attempts(maxRetries),
		)
	}

	opts = append(
		opts,
		retry.OnRetry(func(n uint, err error) {
			logEvent := log.Warn().Err(err).Uint("retries", n)
			for key, val := range logs {
				logEvent = logEvent.Str(key, val)
			}
			if n < maxRetries {
				logEvent.Msgf("%s Error", funcName)
			} else {
				logEvent.Msgf("%s Error Max Retry", funcName)
			}
		}),
	)

	err := retry.Do(retryFunc, opts...)

	return err
}

func (sc *SolanaClient) lowPrioRetryDelay() time.Duration {
	maxRetries := sc.Config.GetInt("solana.lowPrioRetryDelay")
	return time.Duration(maxRetries) * time.Millisecond
}

func (sc *SolanaClient) lowPrioRetryAttempts() uint {
	return sc.Config.GetUint("solana.lowPrioMaxRetries")
}
