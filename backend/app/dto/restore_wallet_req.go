package dto

type RestoreWalletReq struct {
	WalletName   string `json:"wallet_name,omitempty"`
	SecretPhrase string `json:"secret_phrase" validate:"required"`
	Passphrase   string `json:"passphrase,omitempty"`
}
