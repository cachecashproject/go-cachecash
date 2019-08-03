package main

import (
	"context"
	"flag"

	"github.com/cachecashproject/go-cachecash/ccmsg"
	"github.com/cachecashproject/go-cachecash/common"
	"github.com/cachecashproject/go-cachecash/keypair"
	"github.com/cachecashproject/go-cachecash/ledger"
	_ "github.com/mattn/go-sqlite3"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

var (
	ledgerAddr  = flag.String("ledgerAddr", "localhost:7778", "Address of ledgerd instance")
	keypairPath = flag.String("keypair", "ledger.keypair.json", "Path to keypair file")
)

func main() {
	common.Main(mainC)
}

func getFirstGenesisTransaction(ctx context.Context, l *logrus.Logger, grpcClient ccmsg.LedgerClient) (*ledger.TXID, error) {
	resp, err := grpcClient.GetBlocks(ctx, &ccmsg.GetBlocksRequest{
		StartDepth: 0,
		Limit:      5,
	})
	if err != nil {
		return nil, errors.Wrap(err, "failed to get blocks")
	}

	blocks := resp.Blocks
	if len(blocks) != 1 {
		return nil, errors.New("TODO: chains with more than the genesis block are currently unsupported")
	}

	block := ledger.Block{}
	err = block.Unmarshal(blocks[0])
	if err != nil {
		return nil, errors.Wrap(err, "failed to unmarshal block")
	}

	if len(block.Transactions) == 0 {
		return nil, errors.New("missing transactions in genesis block")
	}

	txid, err := block.Transactions[0].TXID()
	if err != nil {
		return nil, errors.Wrap(err, "failed to get txid")
	}

	return &txid, nil
}

func moveCoins(ctx context.Context, l *logrus.Logger, grpcClient ccmsg.LedgerClient, prevtx ledger.TXID, kp *keypair.KeyPair) (*ledger.TXID, error) {
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
					ScriptSig:  []byte{},
					SequenceNo: 0xFFFFFFFF,
				},
			},
			Outputs: []ledger.TransactionOutput{
				{
					Value:        420000000,
					ScriptPubKey: kp.PublicKey,
				},
			},
			Witnesses: []ledger.TransactionWitness{
				{
					Data: [][]byte{},
				},
			},
			LockTime: 0,
		},
	}

	l.Info("sending transaction to ledgerd...")
	_, err := grpcClient.PostTransaction(ctx, &ccmsg.PostTransactionRequest{Tx: tx})
	if err != nil {
		return nil, errors.Wrap(err, "failed to post transaction")
	}

	resp, err := grpcClient.GetBlocks(ctx, &ccmsg.GetBlocksRequest{
		StartDepth: 0,
		Limit:      5,
	})
	if err != nil {
		return nil, errors.Wrap(err, "failed to get blocks")
	}
	l.Info("got blocks: ", len(resp.Blocks))

	txid, err := tx.TXID()
	if err != nil {
		return nil, err
	}

	return &txid, nil
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

	txid, err := getFirstGenesisTransaction(ctx, l, grpcClient)
	if err != nil {
		return err
	}
	l.Info("1st tx: ", txid)

	txid, err = moveCoins(ctx, l, grpcClient, *txid, kp)
	if err != nil {
		return err
	}
	l.Info("2nd tx: ", txid)

	txid, err = moveCoins(ctx, l, grpcClient, *txid, kp)
	if err != nil {
		return err
	}
	l.Info("3rd tx: ", txid)

	l.Info("fin")
	return nil
}
