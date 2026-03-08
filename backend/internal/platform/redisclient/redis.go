package redisclient

import "errors"

type Config struct {
	URL string
}

type Client struct {
	URL string
}

func New(cfg Config) (*Client, error) {
	if cfg.URL == "" {
		return nil, errors.New("redis url is required")
	}

	return &Client{URL: cfg.URL}, nil
}
