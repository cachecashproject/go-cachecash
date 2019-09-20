package main

import (
	"context"
	"fmt"
	"os"

	"github.com/cachecashproject/go-cachecash/ccmsg"
	"github.com/cachecashproject/go-cachecash/common"
	"github.com/cachecashproject/go-cachecash/keypair"
	"github.com/cachecashproject/go-cachecash/ledger"
	"github.com/cachecashproject/go-cachecash/wallet"
	_ "github.com/mattn/go-sqlite3"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"

	"github.com/urfave/cli"
)

func main() {
	common.Main(mainC)
}

func withWallet(f func(ctx context.Context, c *cli.Context, l *logrus.Logger, wallet *wallet.Wallet) error) func(c *cli.Context) error {
	return func(c *cli.Context) error {
		l := logrus.New()
		ctx := context.Background()

		keypairPath := c.GlobalString("keypair")
		ledgerAddr := c.GlobalString("ledger-addr")
		sync := c.GlobalBool("sync")
		insecure := c.GlobalBool("insecure")
		dbPath := c.GlobalString("wallet-db")

		kp, err := keypair.LoadOrGenerate(l, keypairPath)
		if err != nil {
			return errors.Wrap(err, "failed to get keypair")
		}
		w, err := wallet.NewWallet(l, kp, dbPath, ledgerAddr, insecure)
		if err != nil {
			return errors.Wrap(err, "failed to open wallet")
		}

		if sync {
			err = w.FetchBlocks(ctx)
			if err != nil {
				return errors.Wrap(err, "failed to sync wallet")
			}
		}

		defer w.Close()

		return f(ctx, c, l, w)
	}
}

func mainBalance(ctx context.Context, c *cli.Context, l *logrus.Logger, wallet *wallet.Wallet) error {
	balance, err := wallet.GetBalance(ctx)
	if err != nil {
		return errors.Wrap(err, "failed to get balance")
	}

	fmt.Println(balance)
	return nil
}

func mainAddress(ctx context.Context, c *cli.Context, l *logrus.Logger, wallet *wallet.Wallet) error {
	fmt.Println(wallet.Address())
	return nil
}

func mainSend(ctx context.Context, c *cli.Context, l *logrus.Logger, wallet *wallet.Wallet) error {
	to := c.String("to")
	// TODO: check before casting to uint32
	amount := uint32(c.Uint("amount"))

	address, err := ledger.ParseAddress(to)
	if err != nil {
		return errors.Wrap(err, "failed to parse address")
	}

	err = wallet.SendCoins(ctx, address, amount)
	if err != nil {
		return errors.Wrap(err, "failed to send coins")
	}

	fmt.Printf("transfer %v -> %v\n", to, amount)
	return nil
}

func mainFaucet(ctx context.Context, c *cli.Context, l *logrus.Logger, wallet *wallet.Wallet) error {
	faucetAddr := c.String("faucet-addr")
	insecure := c.GlobalBool("insecure")

	conn, err := common.GRPCDial(faucetAddr, insecure)
	if err != nil {
		return errors.Wrap(err, "failed to dial ledger service")
	}

	grpcClient := ccmsg.NewFaucetClient(conn)
	_, err = grpcClient.GetCoins(ctx, &ccmsg.GetCoinsRequest{
		Address: wallet.Address(),
	})
	if err != nil {
		return err
	}

	return nil
}

func mainC() error {
	app := cli.NewApp()
	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:  "keypair",
			Usage: "Path to keypair file",
			Value: "wallet.keypair.json",
		},
		cli.StringFlag{
			Name:  "ledger-addr",
			Usage: "Address of ledgerd instance",
			Value: "localhost:7778",
		},
		cli.StringFlag{
			Name:  "wallet-db",
			Usage: "Path to wallet db",
			Value: "wallet.db",
		},
		cli.BoolFlag{
			Name:  "sync",
			Usage: "Sync wallet with ledger on open",
		},
		cli.BoolFlag{
			Name:  "insecure",
			Usage: "Use insecure connection to ledger",
		},
	}
	app.Commands = []cli.Command{
		{
			Name:   "balance",
			Usage:  "get current balance",
			Action: withWallet(mainBalance),
		},
		{
			Name:   "address",
			Usage:  "get our own address",
			Action: withWallet(mainAddress),
		},
		{
			Name:  "send",
			Usage: "send coins to address",
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:     "to",
					Usage:    "Destination address",
					Required: true,
				},
				cli.UintFlag{
					Name:     "amount",
					Usage:    "transfer amount",
					Required: true,
				},
			},
			Action: withWallet(mainSend),
		},
		{
			Name:  "faucet",
			Usage: "request coins from a faucet",
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "faucet-addr",
					Usage: "Address of faucet instance",
					Value: "localhost:7781",
				},
			},
			Action: withWallet(mainFaucet),
		},
	}
	return app.Run(os.Args)
}
