package main

import (
	"context"
	"flag"

	"github.com/cachecashproject/go-cachecash/ccmsg"
	"github.com/cachecashproject/go-cachecash/common"
	"github.com/cachecashproject/go-cachecash/keypair"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

var (
	faucetAddr  = flag.String("faucetAddr", "localhost:7781", "Address of faucet instance")
	keypairPath = flag.String("keypair", "wallet.keypair.json", "Path to keypair file")
)

func main() {
	common.Main(mainC)
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

	conn, err := common.GRPCDial(*faucetAddr, insecure)
	if err != nil {
		return errors.Wrap(err, "failed to dial ledger service")
	}

	grpcClient := ccmsg.NewFaucetClient(conn)

	_, err = grpcClient.GetCoins(ctx, &ccmsg.GetCoinsRequest{
		PublicKey: &ccmsg.PublicKey{
			PublicKey: kp.PublicKey,
		},
	})
	if err != nil {
		return err
	}

	l.Info("success")

	return nil
}
