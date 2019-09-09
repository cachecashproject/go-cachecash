package main

import (
	"context"
	"flag"
	"math/rand"
	"time"

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
	mixRounds   = flag.Int("mix-rounds", 3, "Run the mixer n times")
	mixWallets  = flag.Int("mix-wallets", 12, "Run the mixer with n wallets")
)

// sudo chmod 0666 data/ledger/ledger.keypair.json && go run ./cmd/mixer-sim -keypair data/ledger/ledger.keypair.json

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

type wallet struct {
	l     *logrus.Logger
	kp    *keypair.KeyPair
	utxos []*utxo
}

func (w *wallet) balance() uint64 {
	sum := uint64(0)
	for _, utxo := range w.utxos {
		sum += uint64(utxo.value)
	}
	return sum
}

func (w *wallet) spendUtxos() ([]ledger.TransactionInput, []ledger.TransactionOutput, error) {
	inputs := []ledger.TransactionInput{}
	prevOutputs := []ledger.TransactionOutput{}

	for _, utxo := range w.utxos {
		w.l.Info("spending utxo: ", utxo.value)

		inputScriptBytes, err := txscript.MakeOutputScript(w.kp.PublicKey)
		if err != nil {
			return nil, nil, errors.Wrap(err, "failed to create output script")
		}

		inputs = append(inputs, ledger.TransactionInput{
			Outpoint: ledger.Outpoint{
				PreviousTx: utxo.txid,
				Index:      utxo.idx,
			},
			ScriptSig:  inputScriptBytes,
			SequenceNo: 0xFFFFFFFF,
		})
		prevOutputs = append(prevOutputs, utxo.txo)
	}

	return inputs, prevOutputs, nil
}

func (w *wallet) addUTXO(utxo *utxo) {
	w.utxos = append(w.utxos, utxo)
}

type utxo struct {
	txid ledger.TXID
	idx  uint8

	value uint32
	txo   ledger.TransactionOutput
}

type simulator struct {
	l          *logrus.Logger
	grpcClient ccmsg.LedgerClient
	wallets    map[string]*wallet
}

func (s *simulator) addGenesisWallet(kp *keypair.KeyPair, txid ledger.TXID, prevOutputs []ledger.TransactionOutput) {
	s.wallets[string(kp.PublicKey)] = &wallet{
		l:  s.l,
		kp: kp,
		utxos: []*utxo{
			{
				txid:  txid,
				idx:   0,
				value: prevOutputs[0].Value,
				txo:   prevOutputs[0],
			},
		},
	}
}

func (s *simulator) genWallet() error {
	kp, err := keypair.Generate()
	if err != nil {
		return err
	}
	s.wallets[string(kp.PublicKey)] = &wallet{
		l:  s.l,
		kp: kp,
	}
	return nil
}

func (s *simulator) randomWallets() []ed25519.PublicKey {
	num := 0
	for num == 0 {
		num = rand.Intn(len(s.wallets))
	}

	keys := make([]ed25519.PublicKey, 0, len(s.wallets))
	for k := range s.wallets {
		keys = append(keys, ed25519.PublicKey(k))
	}

	rand.Shuffle(len(keys), func(i, j int) { keys[i], keys[j] = keys[j], keys[i] })

	return keys[:num]
}

func (s *simulator) genOutputs(wallet *wallet) ([]ledger.TransactionOutput, map[string]*utxo, error) {
	outputWallets := s.randomWallets()
	numWallets := uint64(len(outputWallets))

	balance := wallet.balance()
	remaining := balance % numWallets
	perWalletAmount := (balance - remaining) / numWallets

	outputs := []ledger.TransactionOutput{}
	utxos := map[string]*utxo{}

	for i, wallet := range outputWallets {
		amount := perWalletAmount
		if i == 0 {
			amount += remaining
		}

		s.l.Infof("sending %d coins to %v", amount, wallet)

		outputScriptBytes, err := txscript.MakeInputScript(wallet)
		if err != nil {
			return nil, nil, errors.Wrap(err, "failed to create input script")
		}

		txo := ledger.TransactionOutput{
			Value:        uint32(amount),
			ScriptPubKey: outputScriptBytes,
		}
		outputs = append(outputs, txo)

		utxos[string(wallet)] = &utxo{
			// txid is set at the end
			idx:   uint8(i),
			value: txo.Value,
			txo:   txo,
		}
	}

	return outputs, utxos, nil
}

func (s *simulator) submit(tx ledger.Transaction, prevOutputs []ledger.TransactionOutput, wallet *wallet) error {
	ctx := context.Background()

	s.l.Info("sending transaction to ledgerd...")
	_, err := s.grpcClient.PostTransaction(ctx, &ccmsg.PostTransactionRequest{Tx: tx})
	if err != nil {
		return errors.Wrap(err, "failed to post transaction")
	}
	s.l.Info("block got accepted")

	return nil
}

func (s *simulator) pollBlock(height uint64) (uint64, error) {
	ctx := context.Background()

	for {
		s.l.Info("polling for new block")
		resp, err := s.grpcClient.GetBlocks(ctx, &ccmsg.GetBlocksRequest{
			StartDepth: height,
			Limit:      50,
		})
		if err != nil {
			return height, errors.Wrap(err, "failed to get blocks")
		}

		blocks := uint64(len(resp.Blocks))
		height = height + blocks
		s.l.Info("current blocks height: ", height)

		if blocks > 0 {
			return height, nil
		}

		time.Sleep(2 * time.Second)
	}
}

func (s *simulator) run(rounds int) error {
	height := uint64(0)

	for i := 0; i <= rounds; i++ {
		s.l.Infof("starting round: %d/%d", i, rounds)
		for _, wallet := range s.wallets {
			s.l.Info("mixing coins from ", wallet.kp.PublicKey)
			if wallet.balance() == 0 {
				s.l.Info("wallet is empty, skipping")
				continue
			}

			inputs, prevOutputs, err := wallet.spendUtxos()
			if err != nil {
				return errors.Wrap(err, "failed to collect utxos")
			}
			outputs, utxos, err := s.genOutputs(wallet)
			if err != nil {
				return errors.Wrap(err, "failed to generate outputs")
			}
			wallet.utxos = []*utxo{} // everything is spent

			tx := ledger.Transaction{
				Version: 1,
				Flags:   0,
				Body: &ledger.TransferTransaction{
					Inputs:   inputs,
					Outputs:  outputs,
					LockTime: 0,
				},
			}
			err = tx.GenerateWitnesses(wallet.kp, prevOutputs)
			if err != nil {
				return errors.Wrap(err, "failed to generate witnesses")
			}

			for key, utxo := range utxos {
				txid, err := tx.TXID()
				if err != nil {
					return err
				}
				utxo.txid = txid
				s.wallets[key].addUTXO(utxo)
			}

			err = s.submit(tx, prevOutputs, wallet)
			if err != nil {
				return err
			}
		}

		var err error
		height, err = s.pollBlock(height)
		if err != nil {
			return err
		}

		s.l.Infof("ending round: %d/%d", i, rounds)
	}

	return nil
}

func mainC() error {
	l := logrus.New()
	p, err := common.NewConfigParser(l, "mixer-sim")
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
	l.Info("genesis tx: ", txid)

	s := &simulator{
		l:          l,
		grpcClient: grpcClient,
		wallets:    map[string]*wallet{},
	}

	s.addGenesisWallet(kp, *txid, prevOutputs)
	for i := 0; i < *mixWallets; i++ {
		l.Info("adding wallet: ", i)
		err = s.genWallet()
		if err != nil {
			return errors.Wrap(err, "failed to generate wallet")
		}
	}

	err = s.run(*mixRounds)
	if err != nil {
		return err
	}

	l.Info("fin")
	return nil
}
