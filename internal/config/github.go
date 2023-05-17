package config

import (
	"encoding/json"
	"os"

	"gitlab.com/distributed_lab/logan/v3/errors"
)

type GithubCfg struct {
	SuperToken string `json:"super_token"`
	UsualToken string `json:"usual_token"`
}

func (c *config) Github() *GithubCfg {
	return c.github.Do(func() interface{} {
		var cfg GithubCfg
		value, ok := os.LookupEnv("github")
		if !ok {
			panic(errors.New("no github env variable"))
		}
		err := json.Unmarshal([]byte(value), &cfg)
		if err != nil {
			panic(errors.Wrap(err, "failed to figure out github params from env variable"))
		}

		return &cfg
	}).(*GithubCfg)
}
