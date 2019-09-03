package main

import (
	"encoding/base64"
	"fmt"
	"os"

	"github.com/cachecashproject/go-cachecash/common"
	"github.com/cachecashproject/go-cachecash/ledger/txscript"
	"github.com/pkg/errors"
)

func main() {
	common.Main(mainC)
}

func mainC() error {
	if len(os.Args) < 2 {
		return errors.New("Usage: scriptdump <base64 script>")
	}

	data, err := base64.StdEncoding.DecodeString(os.Args[1])
	if err != nil {
		return errors.Wrap(err, "Failed to decode base64")
	}

	script, err := txscript.ParseScript(data)
	if err != nil {
		return errors.Wrap(err, "Failed to parse script")
	}

	pretty, err := script.PrettyPrint()
	if err != nil {
		return errors.Wrap(err, "Failed to pretty print")
	}

	fmt.Println(pretty)
	return nil
}
