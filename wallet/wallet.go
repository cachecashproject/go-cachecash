package wallet

import (
	"context"
	"crypto/sha256"
	"database/sql"

	"github.com/cachecashproject/go-cachecash/ccmsg"
	"github.com/cachecashproject/go-cachecash/common"
	"github.com/cachecashproject/go-cachecash/keypair"
	"github.com/cachecashproject/go-cachecash/ledger"
	"github.com/cachecashproject/go-cachecash/ledger/txscript"
	"github.com/cachecashproject/go-cachecash/wallet/migrations"
	"github.com/cachecashproject/go-cachecash/wallet/models"
	_ "github.com/mattn/go-sqlite3"
	"github.com/pkg/errors"
	migrate "github.com/rubenv/sql-migrate"
	"github.com/sirupsen/logrus"
	"github.com/volatiletech/sqlboiler/boil"
	"github.com/volatiletech/sqlboiler/queries/qm"
	"golang.org/x/crypto/ed25519"
)

type Account struct {
	PublicKey  ed25519.PublicKey
	PrivateKey ed25519.PrivateKey // May be nil
}

func GenerateAccount() (*Account, error) {
	pub, priv, err := ed25519.GenerateKey(nil)
	if err != nil {
		return nil, errors.Wrap(err, "failed to generate keypair")
	}

	return &Account{
		PublicKey:  pub,
		PrivateKey: priv,
	}, nil
}

func (ac *Account) P2WPKHAddress(v ledger.AddressVersion) *ledger.P2WPKHAddress {
	pkh := sha256.Sum256(ac.PublicKey)

	return &ledger.P2WPKHAddress{
		AddressVersion:        v,
		WitnessProgramVersion: 0,
		PublicKeyHash:         pkh[:ledger.AddressHashSize],
	}
}

type Wallet struct {
	l    *logrus.Logger
	kp   *keypair.KeyPair
	db   *sql.DB
	grpc ccmsg.LedgerClient
}

func NewWallet(l *logrus.Logger, kp *keypair.KeyPair, dbPath string, ledgerAddr string, insecure bool) (*Wallet, error) {
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, errors.Wrap(err, "failed to connect to database")
	}
	l.Info("opened database")

	l.Info("applying migrations")
	n, err := migrate.Exec(db, "sqlite3", migrations.Migrations, migrate.Up)
	if err != nil {
		return nil, errors.Wrap(err, "failed to apply migrations")
	}
	l.Infof("applied %d migrations", n)

	conn, err := common.GRPCDial(ledgerAddr, insecure)
	if err != nil {
		return nil, errors.Wrap(err, "failed to dial ledger service")
	}

	grpc := ccmsg.NewLedgerClient(conn)

	return &Wallet{
		l:    l,
		kp:   kp,
		db:   db,
		grpc: grpc,
	}, nil
}

func (w *Wallet) PublicKey() ed25519.PublicKey {
	return w.kp.PublicKey
}

func (w *Wallet) BlockHeight(ctx context.Context) (int64, error) {
	count, err := models.Blocks().Count(ctx, w.db)
	return int64(count), err
}

func (w *Wallet) FetchBlocks(ctx context.Context) error {
	height, err := w.BlockHeight(ctx)
	if err != nil {
		return err
	}

	w.l.WithFields(logrus.Fields{
		"height": height,
	}).Info("Fetching blocks")
	resp, err := w.grpc.GetBlocks(ctx, &ccmsg.GetBlocksRequest{
		StartDepth: height,
		Limit:      5,
	})
	if err != nil {
		return errors.Wrap(err, "failed to fetch blocks")
	}

	if len(resp.Blocks) == 0 {
		w.l.Info("No new blocks")
	}

	for _, block := range resp.Blocks {
		w.l.Info("Adding block")
		err = w.AddBlock(ctx, *block)
		if err != nil {
			return err
		}
		bytes, err := block.Marshal()
		if err != nil {
			return err
		}

		blockModel := &models.Block{
			Height: int64(height),
			Bytes:  string(bytes),
		}
		err = blockModel.Insert(ctx, w.db, boil.Infer())
		if err != nil {
			return nil
		}
	}

	return nil
}

func (w *Wallet) MatchesOurWallet(output ledger.TransactionOutput) (bool, error) {
	pubKeyHash := txscript.Hash160Sum(w.kp.PublicKey)
	script, err := txscript.MakeP2WPKHInputScript(pubKeyHash)
	if err != nil {
		return false, errors.Wrap(err, "failed to crate p2wpkh input script")
	}

	scriptBytes, err := script.Marshal()
	if err != nil {
		return false, errors.Wrap(err, "failed to marshal input script")
	}

	matches := string(output.ScriptPubKey) == string(scriptBytes)
	return matches, nil
}

func (w *Wallet) AddBlock(ctx context.Context, block ledger.Block) error {
	for _, tx := range block.Transactions.Transactions {
		txid, err := tx.TXID()
		if err != nil {
			return err
		}

		err = w.markTXInputsSpent(ctx, tx)
		if err != nil {
			return err
		}

		for idx, output := range tx.Outputs() {
			if matches, err := w.MatchesOurWallet(output); err != nil || !matches {
				continue
			}

			w.l.Info("discovered spendable transaction output")
			err = w.AddUTXO(ctx, &models.Utxo{
				Txid:         string(txid[:]),
				Idx:          int64(idx),
				Amount:       int64(output.Value),
				ScriptPubkey: string(output.ScriptPubKey),
			})
			if err != nil {
				return errors.Wrap(err, "failed to add utxo to db")
			}
		}
	}

	return nil
}

func (w *Wallet) AddUTXO(ctx context.Context, utxo *models.Utxo) error {
	return utxo.Insert(ctx, w.db, boil.Infer())
}

func (w *Wallet) GetUTXOs(ctx context.Context) ([]*models.Utxo, error) {
	return models.Utxos().All(ctx, w.db)
}

func (w *Wallet) DeleteUTXO(ctx context.Context, utxo ledger.Outpoint) error {
	_, err := models.Utxos(qm.Where("txid = ? and idx = ?", string(utxo.PreviousTx[:]), utxo.Index)).DeleteAll(ctx, w.db)
	return err
}

func (w *Wallet) generateTX(ctx context.Context, target ledger.Address, amount uint32) (*ledger.Transaction, error) {
	utxos, err := w.GetUTXOs(ctx)
	if err != nil {
		return nil, err
	}

	inputs := []ledger.TransactionInput{}
	prevOutputs := []ledger.TransactionOutput{}
	outputs := []ledger.TransactionOutput{}

	spendingSum := uint32(0)
	for _, utxo := range utxos {
		spendingSum += uint32(utxo.Amount)

		txid := ledger.TXID{}
		copy(txid[:], utxo.Txid)

		pubKeyHash := txscript.Hash160Sum(w.kp.PublicKey)
		scriptSig, err := txscript.MakeP2WPKHOutputScript(pubKeyHash)
		if err != nil {
			return nil, errors.Wrap(err, "todo")
		}
		scriptSigBytes, err := scriptSig.Marshal()
		if err != nil {
			return nil, errors.Wrap(err, "failed to marshal input script")
		}

		inputs = append(inputs, ledger.TransactionInput{
			Outpoint: ledger.Outpoint{
				PreviousTx: txid,
				Index:      uint8(utxo.Idx),
			},
			ScriptSig:  scriptSigBytes,
			SequenceNo: 0xFFFFFFFF,
		})

		prevOutputs = append(prevOutputs, ledger.TransactionOutput{
			Value:        uint32(utxo.Amount),
			ScriptPubKey: []byte(utxo.ScriptPubkey),
		})

		if spendingSum >= amount {
			break
		}
	}

	if spendingSum < amount {
		return nil, errors.New("insufficient funds")
	}

	scriptPubkey, err := txscript.MakeP2WPKHInputScript(target.PubKeyHash())
	if err != nil {
		return nil, errors.Wrap(err, "failed to create p2wpkh input script")
	}
	scriptPubkeyBytes, err := scriptPubkey.Marshal()
	if err != nil {
		return nil, errors.Wrap(err, "failed to marshal p2wpkh input script")
	}

	outputs = append(outputs, ledger.TransactionOutput{
		Value:        amount,
		ScriptPubKey: scriptPubkeyBytes,
	})

	change := spendingSum - amount
	if change > 0 {
		w.l.Info("adding change to tx: ", change)

		pubKeyHash := txscript.Hash160Sum(w.kp.PublicKey)
		scriptPubkey, err := txscript.MakeP2WPKHInputScript(pubKeyHash)
		if err != nil {
			return nil, errors.Wrap(err, "todo")
		}
		scriptPubkeyBytes, err := scriptPubkey.Marshal()
		if err != nil {
			return nil, errors.Wrap(err, "todo")
		}

		outputs = append(outputs, ledger.TransactionOutput{
			Value:        change,
			ScriptPubKey: scriptPubkeyBytes,
		})
	}

	tx := ledger.Transaction{
		Version: 1,
		Flags:   0,
		Body: &ledger.TransferTransaction{
			Inputs:   inputs,
			Outputs:  outputs,
			LockTime: 0,
		},
	}

	err = tx.GenerateWitnesses(w.kp, prevOutputs)
	if err != nil {
		return nil, errors.Wrap(err, "failed to generate witnesses")
	}

	return &tx, nil
}

func (w *Wallet) markTXInputsSpent(ctx context.Context, tx *ledger.Transaction) error {
	// mark outputs as spent
	for _, txo := range tx.Inputs() {
		err := w.DeleteUTXO(ctx, txo.Outpoint)
		if err != nil {
			return errors.Wrap(err, "failed to mark utxo as spent")
		}
	}
	return nil
}

func (w *Wallet) Address() string {
	address := ledger.MakeP2WPKHAddress(w.kp.PublicKey)
	return address.Base58Check()
}

func (w *Wallet) GetBalance(ctx context.Context) (uint64, error) {
	utxos, err := w.GetUTXOs(ctx)
	if err != nil {
		return 0, err
	}

	sum := uint64(0)
	for _, utxo := range utxos {
		sum += uint64(utxo.Amount)
	}
	return sum, nil
}

func (w *Wallet) SendCoins(ctx context.Context, target ledger.Address, amount uint32) error {
	w.l.Info("generating transaction")
	tx, err := w.generateTX(ctx, target, amount)
	if err != nil {
		return errors.Wrap(err, "failed to generate tx")
	}

	w.l.Info("sending transaction to ledgerd...")
	_, err = w.grpc.PostTransaction(ctx, &ccmsg.PostTransactionRequest{Tx: *tx})
	if err != nil {
		return errors.Wrap(err, "failed to post transaction")
	}
	w.l.Info("tx got accepted")

	err = w.markTXInputsSpent(ctx, tx)
	if err != nil {
		return errors.Wrap(err, "failed to mark tx inputs as spent")
	}

	return nil
}

func (w *Wallet) Close() error {
	return w.db.Close()
}
