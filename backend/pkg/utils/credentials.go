package utils

import (
	"fmt"

	"github.com/create-go-app/fiber-go-template/pkg/repository"
)

// GetCredentialsByRole func for getting credentials from a role name.
func GetCredentialsByRole(role string) ([]string, error) {
	// Define credentials variable.
	var credentials []string

	// Switch given role.
	switch role {
	case repository.AdminRoleName:
		// Admin credentials (all access).
		credentials = []string{
			repository.TaskCreateCredential,
			repository.TaskUpdateCredential,
			repository.TaskDeleteCredential,
			repository.TaskViewCredential,
			repository.HistoryCreateCredential,
			repository.HistoryViewCredential,
		}
	case repository.ModeratorRoleName:
		credentials = []string{
			repository.TaskCreateCredential,
			repository.TaskUpdateCredential,
			repository.TaskViewCredential,
		}
	case repository.UserRoleName:
		credentials = []string{
			repository.TaskCreateCredential,
		}
	default:
		// Return error message.
		return nil, fmt.Errorf("role '%v' does not exist", role)
	}

	return credentials, nil
}
