package main

import (
	"context"
	"flag"

	"github.com/cachecashproject/go-cachecash/ccmsg"
	"github.com/cachecashproject/go-cachecash/common"
	"github.com/cachecashproject/go-cachecash/keypair"
	"github.com/cachecashproject/go-cachecash/ledger"
	"github.com/cachecashproject/go-cachecash/ledger/txscript"
	_ "github.com/mattn/go-sqlite3"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"golang.org/x/crypto/ed25519"
)

var (
	ledgerAddr  = flag.String("ledgerAddr", "localhost:7778", "Address of ledgerd instance")
	keypairPath = flag.String("keypair", "ledger.keypair.json", "Path to keypair file")
)

// sudo chmod 0666 data/ledger/ledger.keypair.json && go run ./cmd/ledger-cli -keypair data/ledger/ledger.keypair.json

func main() {
	common.Main(mainC)
}

func getFirstGenesisTransaction(ctx context.Context, l *logrus.Logger, grpcClient ccmsg.LedgerClient) (*ledger.TXID, []ledger.TransactionOutput, error) {
	resp, err := grpcClient.GetBlocks(ctx, &ccmsg.GetBlocksRequest{
		StartDepth: 0,
		Limit:      5,
	})
	if err != nil {
		return nil, nil, errors.Wrap(err, "failed to get blocks")
	}

	blocks := resp.Blocks
	if len(blocks) != 1 {
		return nil, nil, errors.New("TODO: chains with more than the genesis block are currently unsupported")
	}

	block := ledger.Block{}
	err = block.Unmarshal(blocks[0])
	if err != nil {
		return nil, nil, errors.Wrap(err, "failed to unmarshal block")
	}

	if len(block.Transactions) == 0 {
		return nil, nil, errors.New("missing transactions in genesis block")
	}

	txid, err := block.Transactions[0].TXID()
	if err != nil {
		return nil, nil, errors.Wrap(err, "failed to get txid")
	}

	return &txid, block.Transactions[0].Outputs(), nil
}

func makeOutputScript(pubkey ed25519.PublicKey) ([]byte, error) {
	pubKeyHash := txscript.Hash160Sum(pubkey)
	scriptPubKey, err := txscript.MakeP2WPKHOutputScript(pubKeyHash)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create scriptPubKey")
	}

	scriptBytes, err := scriptPubKey.Marshal()
	if err != nil {
		return nil, errors.Wrap(err, "failed to marshal output script")
	}

	return scriptBytes, nil
}

func makeInputScript(pubkey ed25519.PublicKey) ([]byte, error) {
	pubKeyHash := txscript.Hash160Sum(pubkey)
	scriptPubKey, err := txscript.MakeP2WPKHInputScript(pubKeyHash)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create scriptPubKey")
	}

	scriptBytes, err := scriptPubKey.Marshal()
	if err != nil {
		return nil, errors.Wrap(err, "failed to marshal input script")
	}

	return scriptBytes, nil
}

func moveCoins(ctx context.Context, l *logrus.Logger, grpcClient ccmsg.LedgerClient, prevtx ledger.TXID, prevOutputs []ledger.TransactionOutput, kp *keypair.KeyPair, target *keypair.KeyPair) (*ledger.TXID, []ledger.TransactionOutput, error) {
	inputScriptBytes, err := makeOutputScript(kp.PublicKey)
	if err != nil {
		return nil, nil, errors.Wrap(err, "failed to create input script")
	}

	outputScriptBytes, err := makeInputScript(target.PublicKey)
	if err != nil {
		return nil, nil, errors.Wrap(err, "failed to create output script")
	}

	tx := ledger.Transaction{
		Version: 1,
		Flags:   0,
		Body: &ledger.TransferTransaction{
			Inputs: []ledger.TransactionInput{
				{
					Outpoint: ledger.Outpoint{
						PreviousTx: prevtx,
						Index:      0,
					},
					ScriptSig:  inputScriptBytes,
					SequenceNo: 0xFFFFFFFF,
				},
			},
			Outputs: []ledger.TransactionOutput{
				{
					Value:        420000000,
					ScriptPubKey: outputScriptBytes,
				},
			},
			LockTime: 0,
		},
	}
	err = tx.GenerateWitnesses(kp, prevOutputs)
	if err != nil {
		return nil, nil, errors.Wrap(err, "failed to generate witnesses")
	}

	l.Info("sending transaction to ledgerd...")
	_, err = grpcClient.PostTransaction(ctx, &ccmsg.PostTransactionRequest{Tx: tx})
	if err != nil {
		return nil, nil, errors.Wrap(err, "failed to post transaction")
	}

	resp, err := grpcClient.GetBlocks(ctx, &ccmsg.GetBlocksRequest{
		StartDepth: 0,
		Limit:      5,
	})
	if err != nil {
		return nil, nil, errors.Wrap(err, "failed to get blocks")
	}
	l.Info("got blocks: ", len(resp.Blocks))

	txid, err := tx.TXID()
	if err != nil {
		return nil, nil, err
	}

	return &txid, tx.Outputs(), nil
}

func mainC() error {
	l := logrus.New()
	p, err := common.NewConfigParser(l, "ledger-cli")
	if err != nil {
		return err
	}
	insecure := p.GetInsecure()
	flag.Parse()
	ctx := context.Background()

	kp, err := keypair.LoadOrGenerate(l, *keypairPath)
	if err != nil {
		return errors.Wrap(err, "failed to get keypair")
	}

	conn, err := common.GRPCDial(*ledgerAddr, insecure)
	if err != nil {
		return errors.Wrap(err, "failed to dial ledger service")
	}

	grpcClient := ccmsg.NewLedgerClient(conn)

	txid, prevOutputs, err := getFirstGenesisTransaction(ctx, l, grpcClient)
	if err != nil {
		return err
	}
	l.Info("1st tx: ", txid)

	target, err := keypair.Generate()
	if err != nil {
		return errors.Wrap(err, "failed to generate next target address")
	}
	txid, prevOutputs, err = moveCoins(ctx, l, grpcClient, *txid, prevOutputs, kp, target)
	if err != nil {
		return err
	}
	l.Info("2nd tx: ", txid)
	kp = target

	target, err = keypair.Generate()
	if err != nil {
		return errors.Wrap(err, "failed to generate next target address")
	}
	txid, prevOutputs, err = moveCoins(ctx, l, grpcClient, *txid, prevOutputs, kp, target)
	if err != nil {
		return err
	}
	l.Info("3rd tx: ", txid)
	// kp = target

	l.Info("fin")
	return nil
}
