package elasticsearch

import (
	v7 "github.com/olivere/elastic/v7"
	"github.com/pkg/errors"
)

type Config struct {
	URL         string `mapstructure:"url"`
	Sniff       bool   `mapstructure:"sniff"`
	Gzip        bool   `mapstructure:"gzip"`
	Explain     bool   `mapstructure:"explain"`
	FetchSource bool   `mapstructure:"fetchSource"`
	Version     bool   `mapstructure:"version"`
	Pretty      bool   `mapstructure:"pretty"`
}

func NewElasticClient(config Config) (*v7.Client, error) {
	client, err := v7.NewClient(
		v7.SetURL(config.URL),
		v7.SetSniff(config.Sniff),
		v7.SetGzip(config.Gzip),
	)
	if err != nil {
		return nil, errors.Wrap(err, "v7.NewClient")
	}

	return client, nil
}
