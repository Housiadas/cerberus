#!/bin/sh

export VAULT_ADDR='http://127.0.0.1:8200'
export VAULT_TOKEN='dev-root-token'

# Wait for Vault to be ready
until vault status > /dev/null 2>&1; do
  echo "Waiting for Vault to be ready..."
  sleep 1
done

echo "Vault is ready, initializing secrets..."

# Enable KV secrets engine v2 at path 'secret'
vault secrets enable -path=secret -version=2 kv 2>/dev/null || true

# Store the JWT access token secret
vault kv put secret/cerberus/jwt \
  access_token_secret="your-256-bit-access-secret-from-vault"

echo "Secrets initialized successfully!"

# Keep container running
tail -f /dev/null
