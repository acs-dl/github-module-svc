package config

import (
	"context"
	"os"

	vault "github.com/hashicorp/vault/api"
	"gitlab.com/distributed_lab/logan/v3/errors"
)

type GithubCfg struct {
	SuperToken string `json:"super_token"`
	UsualToken string `json:"usual_token"`
}

func (c *config) Github() *GithubCfg {
	return c.github.Do(func() interface{} {
		var cfg GithubCfg

		vaultCfg := vault.DefaultConfig()
		vaultCfg.Address = os.Getenv("VAULT_ADDR")

		client, err := vault.NewClient(vaultCfg)
		if err != nil {
			panic(errors.Wrap(err, "failed to initialize a Vault client"))
		}

		client.SetToken(os.Getenv("VAULT_TOKEN"))

		secret, err := client.KVv2("secret").Get(context.Background(), "github")
		if err != nil {
			panic(errors.Wrap(err, "failed to read from the vault"))
		}

		value, ok := secret.Data["super_token"].(string)
		if !ok {
			panic(errors.New("super token has wrong type"))
		}
		cfg.SuperToken = value

		value, ok = secret.Data["usual_token"].(string)
		if !ok {
			panic(errors.New("usual token has wrong type"))
		}
		cfg.UsualToken = value

		return &cfg
	}).(*GithubCfg)
}
