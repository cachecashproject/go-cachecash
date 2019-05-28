package keypair

import (
	"encoding/json"
	"io/ioutil"
	"os"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"golang.org/x/crypto/ed25519"
)

type KeyPair struct {
	PublicKey  ed25519.PublicKey  `json:"public_key"`
	PrivateKey ed25519.PrivateKey `json:"private_key"`
}

func LoadOrGenerate(l *logrus.Logger, path string) (*KeyPair, error) {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		l.Info("keypair doesn't exist, generating")
		if err := GenerateFile(path); err != nil {
			return nil, err
		}
	}
	return Load(path)
}

func Generate() (*KeyPair, error) {
	pub, priv, err := ed25519.GenerateKey(nil)
	if err != nil {
		return nil, errors.Wrap(err, "failed to generate keypair")
	}
	return &KeyPair{
		PublicKey:  pub,
		PrivateKey: priv,
	}, nil
}

func GenerateFile(path string) error {
	kp, err := Generate()
	if err != nil {
		return err
	}

	buf, err := json.MarshalIndent(kp, "", "  ")
	if err != nil {
		return errors.Wrap(err, "failed to marshal keypair")
	}

	err = ioutil.WriteFile(path, buf, 0600)
	if err != nil {
		return errors.Wrap(err, "failed to write keypair")
	}

	return nil
}

func Load(path string) (*KeyPair, error) {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var kp KeyPair
	if err := json.Unmarshal(data, &kp); err != nil {
		return nil, err
	}

	return &kp, nil
}
