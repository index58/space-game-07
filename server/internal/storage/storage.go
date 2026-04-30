package storage

import (
	"path/filepath"

	"space-game-07-server/internal/data"
)

const accountsFileName = "Accounts.json"

// ServerData объединяет данные сервера, загружаемые из JSON-файлов при старте.
type ServerData struct {
	Accounts *data.Accounts
}

// LoadServerData загружает все JSON-файлы данных сервера из указанного рабочего каталога.
func LoadServerData(workingDirectory string) (*ServerData, error) {
	accounts := data.NewAccounts()
	if err := accounts.LoadFromFile(filepath.Join(workingDirectory, "data", accountsFileName)); err != nil {
		return nil, err
	}

	return &ServerData{
		Accounts: accounts,
	}, nil
}
