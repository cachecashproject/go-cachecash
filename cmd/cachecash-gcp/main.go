package main

import (
	"context"
	"database/sql"
	"flag"
	"io/ioutil"
	"os"
	"path"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli"
	"gopkg.in/yaml.v2"

	"github.com/cachecashproject/go-cachecash/common"
	"github.com/cachecashproject/go-cachecash/dbtx"
	"github.com/cachecashproject/go-cachecash/keypair"
	"github.com/cachecashproject/go-cachecash/ledger"
	"github.com/cachecashproject/go-cachecash/ledgerclient"
	"github.com/cachecashproject/go-cachecash/log"
)

var (
	configPath  = flag.String("config", "gcp.config.json", "Path to configuration file")
	keypairPath = flag.String("keypair", "gcp.keypair.json", "Path to keypair file")
)

type ConfigFile struct {
	Database     string        `json:"database"`
	Insecure     bool          `json:"insecure"`
	LedgerAddr   string        `json:"ledger_address"`
	SyncInterval time.Duration `json:"sync_interval"`
}

func loadConfigFile(l *logrus.Logger, path string) (*ConfigFile, error) {
	conf := &ConfigFile{}
	p, err := common.NewConfigParser(l, "gcp")
	if err != nil {
		return nil, err
	}
	err = p.ReadFile(path)
	if err != nil {
		return nil, err
	}

	conf.Insecure = p.GetInsecure()
	conf.Database = p.GetString("database", "gcp.db")
	conf.LedgerAddr = p.GetString("ledger_addr", "ledger:7778")
	conf.SyncInterval = p.GetSeconds("sync-interval", ledgerclient.DEFAULT_SYNC_INTERVAL)

	return conf, nil
}

func mainC() error {
	app := cli.NewApp()

	app.HideVersion = true
	app.Usage = "Manipulate Cachecash global configuration parameters"

	app.Commands = []cli.Command{
		{
			Name:      "gen",
			ArgsUsage: "[schemafile]",
			Description: `
				Generate go code from a schema file. Exits non-zero and shows the error if there is one, otherwise exits 0
				with no output after writing out the code in the current directory.
				`,
			Usage:  "Generate go code from a schema file.",
			Action: gen,
		},
		{
			Name:      "merge",
			ArgsUsage: "[patchfile]",
			Usage:     "P ",
			Description: `
				Reads the current parameters from the block chain, applies a patch to them, displaying the merged result and then optionally submits that patch to the miner.
			`,
			Flags: []cli.Flag{
				cli.BoolFlag{
					Name:  "submit, s",
					Usage: "submit the result to the chain",
				},
			},
			Action: merge,
		},
	}

	return app.Run(os.Args)
}

func gen(ctx *cli.Context) error {
	if len(ctx.Args()) != 1 {
		return errors.New("must supply a schema file")
	}

	schema, err := ledger.NewGlobalConfigSchemaFromFile(ctx.Args()[0])
	if err != nil {
		return errors.Wrap(err, "failed to parse schema")
	}
	dir, err := os.Getwd()
	if err != nil {
		return errors.Wrap(err, "failed to get working dir")
	}
	content, err := schema.Generate()
	if err != nil {
		return errors.Wrap(err, "failed to get generate code")
	}
	name := path.Join(dir, "globalconfigschema_generated.go")
	return ioutil.WriteFile(name, content, 0644)
}

func merge(c *cli.Context) error {
	l := log.NewCLILogger("cachecash-gcp", log.CLIOpt{})
	if len(c.Args()) != 1 {
		return errors.New("must supply a patch file")
	}

	patch, err := ledger.NewGlobalConfigPatchFromFile(c.Args()[1])
	if err != nil {
		return errors.Wrap(err, "failed to parse patch")
	}

	cf, err := loadConfigFile(&l.Logger, *configPath)
	if err != nil {
		return errors.Wrap(err, "failed to load config file")
	}

	// We need to refactor/create a similar thing to the wallet, for this API
	ctx := dbtx.ContextWithExecutor(context.Background(), nil)
	db, err := sql.Open("sqlite3", cf.Database)
	if err != nil {
		return errors.Wrap(err, "failed to open gcp database")
	}

	persistence := ledger.NewChainStorageSQL(&l.Logger, ledger.NewChainStorageSqlite())
	if err := persistence.RunMigrations(db); err != nil {
		return err
	}
	storage := ledger.NewDatabase(persistence)
	r, err := ledgerclient.NewReplicator(&l.Logger, storage, cf.LedgerAddr, cf.Insecure)
	if err != nil {
		return errors.Wrap(err, "failed to create replicator")
	}
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	go r.SyncChain(ctx, cf.SyncInterval)
	// read the central min
	for {
		synced, err := r.IsSynced(ctx)
		if err != nil {
			return err
		}
		if synced {
			break
		}
		// Sleep with a slight offset so that we don't end up lock-step always
		// checking out of sync. This would be better event driven, but that can
		// be added in future.
		l.Info("Chain not synced, sleeping")
		time.Sleep(cf.SyncInterval + time.Second)
	}
	tx, err := patch.ToTransaction(ctx, storage)
	if err != nil {
		return errors.Wrap(err, "failed to convert patch to Transaction")
	}

	// Need Markus' gcp code here?
	// Should be blockID based?
	// state, err := storage.getGCP(storage.Height())
	state := ledger.NewGlobalConfigState()
	state, err = state.Apply(tx)
	if err != nil {
		return errors.Wrap(err, "transaction application failed")
	}
	l.Info("Current global config parameters:")
	doc, err := state.ToYAML()
	if err != nil {
		return err
	}
	if err := yaml.NewEncoder(os.Stdout).Encode(doc); err != nil {
		return err
	}

	if !c.Bool("submit") {
		return nil
	}
	// Everything from here on in has to do with submitting the transaction to
	// the miner
	kp, err := keypair.Load(*keypairPath)
	if err != nil {
		return errors.Wrap(err, "failed to get keypair")
	}
	tx.SigPublicKey = kp.PublicKey
	// TODO: sign tx with kp.private
	// tx.Signature :=
	// TODO: submit tx to the chain
	// TODO? wait for acceptance on the chain
	return nil
}

func main() {
	common.Main(mainC)
}
