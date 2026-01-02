package services

import (
	"context"
	"time"

	"github.com/create-go-app/fiber-go-template/app/dto"
	models "github.com/create-go-app/fiber-go-template/app/entities"
	"github.com/create-go-app/fiber-go-template/app/interfaces/repositories"
	"github.com/create-go-app/fiber-go-template/app/interfaces/services"
	"github.com/create-go-app/fiber-go-template/pkg/core"
	"github.com/create-go-app/fiber-go-template/pkg/crypto"
	"github.com/google/uuid"
)

type WalletServiceImpl struct {
	walletRepo  repositories.WalletRepository
	addressRepo repositories.BlockchainAddressRepository
	cryptoSvc   crypto.Service
	txManager   repositories.TransactionManager
}

func NewWalletService(
	walletRepo repositories.WalletRepository,
	addressRepo repositories.BlockchainAddressRepository,
	cryptoSvc crypto.Service,
	txManager repositories.TransactionManager,
) services.WalletService {
	return &WalletServiceImpl{
		walletRepo:  walletRepo,
		addressRepo: addressRepo,
		cryptoSvc:   cryptoSvc,
		txManager:   txManager,
	}
}

func (s *WalletServiceImpl) CreateWallet(
	ctx context.Context,
	req *dto.CreateWalletReq,
) (*dto.CreateWalletRes, error) {

	var (
		walletId string
		address  string
	)
	var secretPhrase string

	err := s.txManager.Do(ctx, func(ctx context.Context) error {

		// 1️⃣ Wallet ID
		walletId = uuid.New().String()

		// 2️⃣ Generate mnemonic
		mnemonic, err := s.cryptoSvc.GenerateMnemonic()
		if err != nil {
			return err
		}
		secretPhrase = mnemonic
		// 3️⃣ Encrypt mnemonic
		encryptedMnemonic, err := s.cryptoSvc.EncryptMnemonic(
			mnemonic,
			req.Passphrase,
			walletId,
		)
		if err != nil {
			return err
		}

		// 4️⃣ Hash passphrase (optional)
		var passphraseHash string
		if req.Passphrase != "" {
			passphraseHash, err = s.cryptoSvc.HashPassphrase(req.Passphrase)
			if err != nil {
				return err
			}
		}

		now := time.Now()

		// 5️⃣ Create wallet
		wallet := &models.Wallet{
			WalletId:         walletId,
			WalletName:       req.WalletName,
			SecretPhraseHash: encryptedMnemonic,
			PassphraseHash:   passphraseHash,
			CreateDate:       now,
			UpdateDate:       now,
		}

		if err := s.walletRepo.Create(ctx, wallet); err != nil {
			return err
		}

		// 6️⃣ Generate first address
		address, err = s.cryptoSvc.GenerateAddress(mnemonic)
		if err != nil {
			return err
		}

		addr := &models.BlockchainAddress{
			AddressId:  uuid.New().String(),
			WalletId:   walletId,
			Address:    address,
			CreateDate: now,
			UpdateDate: now,
		}

		if err := s.addressRepo.Create(ctx, addr); err != nil {
			return err
		}

		// ✅ return nil → commit
		return nil
	})

	if err != nil {
		return nil, err
	}

	return &dto.CreateWalletRes{
		WalletId:     walletId,
		Address:      address,
		SecretPhrase: secretPhrase,
	}, nil
}

// RestoreWallet implements [services.WalletService].
func (s *WalletServiceImpl) RestoreWallet(
	ctx context.Context,
	req *dto.RestoreWalletReq,
) (*core.ApiResponse, error) {

	wallets, err := s.walletRepo.ListAll(ctx)
	if err != nil {
		return core.Error(500, "cannot load wallets", err.Error(), nil), nil
	}

	for _, wallet := range wallets {

		// 1️⃣ Verify passphrase (nếu wallet có passphrase)
		if wallet.PassphraseHash != "" {
			if req.Passphrase == "" {
				continue
			}

			if !s.cryptoSvc.VerifyPassphrase(
				wallet.PassphraseHash,
				req.Passphrase,
			) {
				continue
			}
		}

		// 2️⃣ Try decrypt mnemonic
		mnemonic, err := s.cryptoSvc.DecryptMnemonic(
			wallet.SecretPhraseHash,
			req.Passphrase,
			wallet.WalletId,
		)
		if err != nil {
			continue
		}

		// 3️⃣ Compare secret phrase
		if mnemonic != req.SecretPhrase {
			continue
		}

		// ✅ SUCCESS
		addresses := make([]string, 0)
		for _, addr := range wallet.BlockchainAddresses {
			addresses = append(addresses, addr.Address)
		}

		return core.Success(200, "wallet restored", dto.RestoreWalletRes{
			WalletId:  wallet.WalletId,
			Addresses: addresses,
		}, nil), nil
	}

	return core.Error(400, "restore failed", "invalid secret phrase or passphrase", nil), nil
}
