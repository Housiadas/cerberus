package vault

import (
	"context"
	"fmt"

	"github.com/hashicorp/vault-client-go"
)

type Config struct {
	Address string
	Token   string
}

type Client struct {
	client *vault.Client
}

func New(cfg Config) (*Client, error) {
	client, err := vault.New(
		vault.WithAddress(cfg.Address),
	)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrCreateClient, err)
	}

	err = client.SetToken(cfg.Token)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrSetToken, err)
	}

	return &Client{client: client}, nil
}

func (c *Client) GetKVSecret(
	ctx context.Context,
	mountPath, secretPath string,
) (map[string]any, error) {
	resp, err := c.client.Secrets.KvV2Read(ctx, secretPath, vault.WithMountPath(mountPath))
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrReadSecret, err)
	}

	return resp.Data.Data, nil
}

func (c *Client) GetJWTSecret(ctx context.Context) ([]byte, error) {
	secrets, err := c.GetKVSecret(ctx, "secret", "cerberus/jwt")
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrGetJWTSecret, err)
	}

	secretValue, ok := secrets["access_token_secret"].(string)
	if !ok {
		return nil, ErrInvalidSecretValue
	}

	return []byte(secretValue), nil
}
