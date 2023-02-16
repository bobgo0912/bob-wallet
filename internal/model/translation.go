package model

import (
	"github.com/bobgo0912/b0b-common/pkg/sql"
)

const TranslationTableName = "translation"

// Translation translation
type Translation struct {
	Id            uint64 `db:"id" json:"id"`                        //id(PRI)
	Type          int    `db:"type" json:"type"`                    //type
	Amount        uint64 `db:"amount" json:"amount"`                //amount
	TranslationId string `db:"translation_id" json:"translationId"` //tl_id
	Describe      string `db:"describe" json:"describe"`            //describe
	PlayerId      uint64 `db:"player_id" json:"playerId"`           //playerId
	Status        int    `db:"status" json:"status"`                //status 0 1
	Version       int    `db:"version" json:"version"`              //version
}

type TranslationStore struct {
	*sql.BaseStore[Translation]
}

func GetTranslationStore() (*TranslationStore, error) {
	connection, err := GetConnection()
	if err != nil {
		return nil, err
	}
	return &TranslationStore{&sql.BaseStore[Translation]{Db: connection, TableName: TranslationTableName}}, nil
}
