package controllers

import (
	vaultapi "github.com/hashicorp/vault/api"
)

func GetClient(address, token string) (*vaultapi.Client, error) {
	vclient, err := vaultapi.NewClient(nil)
	if err != nil {
		return nil, err
	}
	vclient.SetAddress(address)
	vclient.SetToken(token)
	return vclient, nil
}
