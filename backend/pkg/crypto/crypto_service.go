package crypto

type Service interface {
	// 1. Sinh mnemonic (BIP39)
	GenerateMnemonic() (string, error)

	// 2. Mã hóa mnemonic bằng passphrase + walletId
	EncryptMnemonic(mnemonic, passphrase, walletId string) (string, error)

	// 3. Giải mã mnemonic
	DecryptMnemonic(cipher, passphrase, walletId string) (string, error)

	// 4. Hash passphrase để lưu DB
	HashPassphrase(pass string) (string, error)

	// 5. Verify passphrase khi unlock
	VerifyPassphrase(hash, pass string) bool

	// 6. Sinh address từ mnemonic (HD wallet)
	GenerateAddress(mnemonic string) (string, error)
}
