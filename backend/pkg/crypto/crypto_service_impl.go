package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"io"

	"github.com/btcsuite/btcd/btcutil/hdkeychain"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/tyler-smith/go-bip39"
	"golang.org/x/crypto/bcrypt"
	"golang.org/x/crypto/scrypt"
)

type CryptoServiceImpl struct{}

// =======================
// MNEMONIC
// =======================

func (c *CryptoServiceImpl) GenerateMnemonic() (string, error) {
	entropy, err := bip39.NewEntropy(128) // 12 words
	if err != nil {
		return "", err
	}
	return bip39.NewMnemonic(entropy)
}

// =======================
// ENCRYPT / DECRYPT
// =======================

func (c *CryptoServiceImpl) EncryptMnemonic(
	mnemonic,
	passphrase,
	walletId string,
) (string, error) {

	key, err := deriveKey(passphrase, walletId)
	if err != nil {
		return "", err
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}

	cipherData := gcm.Seal(nil, nonce, []byte(mnemonic), nil)
	payload := append(nonce, cipherData...)

	return base64.StdEncoding.EncodeToString(payload), nil
}

func (c *CryptoServiceImpl) DecryptMnemonic(
	cipherText,
	passphrase,
	walletId string,
) (string, error) {

	raw, err := base64.StdEncoding.DecodeString(cipherText)
	if err != nil {
		return "", err
	}

	key, err := deriveKey(passphrase, walletId)
	if err != nil {
		return "", err
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	if len(raw) < gcm.NonceSize() {
		return "", errors.New("ciphertext too short")
	}

	nonce := raw[:gcm.NonceSize()]
	cipherData := raw[gcm.NonceSize():]

	plain, err := gcm.Open(nil, nonce, cipherData, nil)
	if err != nil {
		return "", err
	}

	return string(plain), nil
}

// =======================
// PASSPHRASE
// =======================

func (c *CryptoServiceImpl) HashPassphrase(pass string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(pass), bcrypt.DefaultCost)
	return string(hash), err
}

func (c *CryptoServiceImpl) VerifyPassphrase(hash, pass string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(pass)) == nil
}

// =======================
// ADDRESS (BIP44 ETH)
// =======================

func (c *CryptoServiceImpl) GenerateAddress(mnemonic string) (string, error) {
	seed := bip39.NewSeed(mnemonic, "")

	masterKey, err := hdkeychain.NewMaster(seed, &chaincfg.MainNetParams)
	if err != nil {
		return "", err
	}

	// m/44'
	purpose, err := masterKey.Derive(hdkeychain.HardenedKeyStart + 44)
	if err != nil {
		return "", err
	}

	// m/44'/60'
	coinType, err := purpose.Derive(hdkeychain.HardenedKeyStart + 60)
	if err != nil {
		return "", err
	}

	// m/44'/60'/0'
	account, err := coinType.Derive(hdkeychain.HardenedKeyStart + 0)
	if err != nil {
		return "", err
	}

	// m/44'/60'/0'/0
	change, err := account.Derive(0)
	if err != nil {
		return "", err
	}

	// m/44'/60'/0'/0/0
	addressKey, err := change.Derive(0)
	if err != nil {
		return "", err
	}

	privKey, err := addressKey.ECPrivKey()
	if err != nil {
		return "", err
	}

	ecdsaKey := privKey.ToECDSA()
	address := crypto.PubkeyToAddress(ecdsaKey.PublicKey)

	return address.Hex(), nil
}

// =======================
// INTERNAL
// =======================

func deriveKey(passphrase, walletId string) ([]byte, error) {
	return scrypt.Key(
		[]byte(passphrase),
		[]byte(walletId),
		32768, // N
		8,     // r
		1,     // p
		32,    // 256-bit
	)
}

// =======================
// CONSTRUCTOR
// =======================

func NewCryptoService() Service {
	return &CryptoServiceImpl{}
}
