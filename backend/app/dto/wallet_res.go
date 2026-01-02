package dto

type CreateWalletRes struct {
	WalletId     string `json:"wallet_id"`
	Address      string `json:"address"`
	SecretPhrase string
}
