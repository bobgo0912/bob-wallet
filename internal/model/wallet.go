package model

import (
	"github.com/bobgo0912/b0b-common/pkg/sql"
	"github.com/jmoiron/sqlx"
	"time"
)

const WalletTableName = "wallet"

// Wallet
type Wallet struct {
	Id       uint64    `db:"id" json:"id"`              //id(PRI)
	PlayerId uint64    `db:"player_id" json:"playerId"` //playerId
	Balance  uint64    `db:"balance" json:"balance"`    //balance
	Freeze   uint64    `db:"freeze" json:"freeze"`      //freeze
	CreateAt time.Time `db:"create_at" json:"createAt"` //createAt
	Version  int       `db:"version" json:"version"`    //version
}

type WalletStore struct {
	*sql.BaseStore[Wallet]
}

func GetConnection() (*sqlx.DB, error) {
	if WalletDb != nil {
		return WalletDb, nil
	}
	var err error
	WalletDb, err = sql.Db("wallet", nil)
	if err != nil {
		return nil, err
	}
	return WalletDb, nil
}

func GetWalletStore() (*WalletStore, error) {
	connection, err := GetConnection()
	if err != nil {
		return nil, err
	}
	return &WalletStore{&sql.BaseStore[Wallet]{Db: connection, TableName: WalletTableName}}, nil
}
