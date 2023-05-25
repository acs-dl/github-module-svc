package config

import (
	knox "gitlab.com/distributed_lab/knox/knox-fork/client/external_kms"

	"gitlab.com/distributed_lab/logan/v3/errors"
)

type GithubCfg struct {
	SuperToken string `json:"super_token"`
	UsualToken string `json:"usual_token"`
}

func (c *config) Github() *GithubCfg {
	return c.github.Do(func() interface{} {
		var cfg GithubCfg
		client := knox.NewKeyManagementClient(c.getter)

		key, err := client.GetKey("super_token", "5165714923704681000")
		if err != nil {
			panic(errors.Wrap(err, "failed to get super token key"))
		}

		cfg.SuperToken = string(key[:])

		key, err = client.GetKey("usual_token", "3128228019338087400")
		if err != nil {
			panic(errors.Wrap(err, "failed to get usual token key"))
		}
		cfg.UsualToken = string(key[:])

		return &cfg
	}).(*GithubCfg)
}
