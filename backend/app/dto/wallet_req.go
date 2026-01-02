package dto

type CreateWalletReq struct {
	WalletName string `json:"wallet_name" validate:"required,min=3,max=50"`
	Passphrase string `json:"passphrase,omitempty"`
}
