package config

import (
	"gitlab.com/distributed_lab/figure"
	"gitlab.com/distributed_lab/kit/kv"
	"gitlab.com/distributed_lab/logan/v3/errors"
)

type GithubCfg struct {
	Token string `figure:"token"`
}

func (c *config) Github() *GithubCfg {
	return c.github.Do(func() interface{} {
		var config GithubCfg
		err := figure.
			Out(&config).
			From(kv.MustGetStringMap(c.getter, "github")).
			Please()
		if err != nil {
			panic(errors.Wrap(err, "failed to figure out github params from config"))
		}

		return &config
	}).(*GithubCfg)
}
