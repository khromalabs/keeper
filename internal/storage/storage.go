package storage

import (
	"fmt"
	"khromalabs/keeper/internal/config"
)

type Storage interface {
	Create(fields map[string]string) (int64,error);
	Init(string, map[string]interface{}) error;
	Read(filter string, i int) ([]interface{},int,error);
	UpdateOrDelete(map[string]string,bool) error;
}

var conf *config.Config

func init() {
	conf = config.Get()
}

func Init(template string, templateData map[string]interface{}) (Storage,error) {
	var s Storage
	switch conf.Storage {
	case "sqlite":
		s = &StorageSqlite{}
	// case "csv":
	// storage := csv.Storage{uid}
	default:
		return nil, fmt.Errorf("Unknown storage: %v", template)
	}
	return s, s.Init(template,templateData)
}
