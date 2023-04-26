package config

import (
	"gitlab.com/distributed_lab/figure"
	"gitlab.com/distributed_lab/kit/kv"
	"gitlab.com/distributed_lab/logan/v3/errors"
)

type GithubCfg struct {
	SuperToken string `figure:"super_token"`
	UsualToken string `figure:"usual_token"`
}

func (c *config) Github() *GithubCfg {
	return c.github.Do(func() interface{} {
		var cfg GithubCfg
		err := figure.
			Out(&cfg).
			With(figure.BaseHooks).
			From(kv.MustGetStringMap(c.getter, "github")).
			Please()

		if err != nil {
			panic(errors.Wrap(err, "failed to figure out github params from config"))
		}

		return &cfg
	}).(*GithubCfg)
}
