package vault

import (
	vaultapi "github.com/hashicorp/vault/api"
)

func GetClient(address, token string) *vaultapi.Client {
	vclient, _ := vaultapi.NewClient(nil)
	vclient.SetAddress(address)
	vclient.SetToken(token)
	return vclient
}
