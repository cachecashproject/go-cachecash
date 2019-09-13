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

func openWallet(ctx context.Context, l *logrus.Logger, c *cli.Context) (*wallet.Wallet, error) {
	keypairPath := c.String("keypair")
	ledgerAddr := c.String("ledger-addr")
	sync := c.Bool("sync")
	insecure := c.Bool("insecure")
	dbPath := c.String("wallet-db")

	kp, err := keypair.LoadOrGenerate(l, keypairPath)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get keypair")
	}
	w, err := wallet.NewWallet(l, kp, dbPath, ledgerAddr, insecure)
	if err != nil {
		return nil, errors.Wrap(err, "failed to open wallet")
	}

	if sync {
		err = w.FetchBlocks(ctx)
		if err != nil {
			return nil, errors.Wrap(err, "failed to sync wallet")
		}
	}

	return w, nil
}

func mainC() error {
	l := logrus.New()

	app := cli.NewApp()
	// Global flags are broken
	// https://github.com/urfave/cli/issues/795
	globalFlags := []cli.Flag{
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
			Name:  "balance",
			Usage: "get current balance",
			Flags: globalFlags,
			Action: func(c *cli.Context) error {
				ctx := context.Background()
				wallet, err := openWallet(ctx, l, c)
				if err != nil {
					return errors.Wrap(err, "failed to open wallet")
				}
				defer wallet.Close()

				balance, err := wallet.GetBalance(ctx)
				if err != nil {
					return errors.Wrap(err, "failed to get balance")
				}

				fmt.Println(balance)
				return nil
			},
		},
		{
			Name:  "address",
			Usage: "get our own address",
			Flags: globalFlags,
			Action: func(c *cli.Context) error {
				ctx := context.Background()
				wallet, err := openWallet(ctx, l, c)
				if err != nil {
					return errors.Wrap(err, "failed to open wallet")
				}
				defer wallet.Close()

				fmt.Println(wallet.Address())
				return nil
			},
		},
		{
			Name:  "send",
			Usage: "send coins to address",
			Flags: append([]cli.Flag{
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
			}, globalFlags...),
			Action: func(c *cli.Context) error {
				ctx := context.Background()
				wallet, err := openWallet(ctx, l, c)
				if err != nil {
					return errors.Wrap(err, "failed to open wallet")
				}
				defer wallet.Close()

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
			},
		},
		{
			Name:  "faucet",
			Usage: "request coins from a faucet",
			Flags: append([]cli.Flag{
				cli.StringFlag{
					Name:  "faucet-addr",
					Usage: "Address of faucet instance",
					Value: "localhost:7781",
				},
			}, globalFlags...),
			Action: func(c *cli.Context) error {
				ctx := context.Background()
				wallet, err := openWallet(ctx, l, c)
				if err != nil {
					return errors.Wrap(err, "failed to open wallet")
				}
				defer wallet.Close()

				faucetAddr := c.String("faucet-addr")
				insecure := c.Bool("insecure")

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
			},
		},
	}
	return app.Run(os.Args)
}
