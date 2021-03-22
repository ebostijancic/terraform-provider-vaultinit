package vault

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestInitVault(t *testing.T) {
	client, err := NewVaultClient("http://localhost:8200")

	assert.Nil(t, err)
	assert.NotNil(t, client)

	keys, err := client.Init(2, 2)
	assert.Nil(t, err)
	assert.NotEmpty(t, keys.KeysBase64)
}

func TestUnsealVault(t *testing.T) {
	client, err := NewVaultClient("http://localhost:8200")

	assert.Nil(t, err)
	assert.NotNil(t, client)
	keysBase64 := []string{
		"",
		"",
	}

	err = client.Unseal(keysBase64)
	assert.Nil(t, err)

}
