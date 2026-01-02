// app/dto/restore_wallet_res.go
package dto

type RestoreWalletRes struct {
	WalletId  string   `json:"wallet_id"`
	Addresses []string `json:"addresses"`
}
