package headscale

import (
	"errors"

	"github.com/jinzhu/gorm"
	// this handles postgreSQL
	_ "github.com/jinzhu/gorm/dialects/postgres"
)

const dbVersion = "1"

// KV struct is a key:value store
// Every DB row is a key value pair
type KV struct {
	Key   string
	Value string
}

func (h *Headscale) initDB() error {
	db, err := gorm.Open("postgres", h.dbString)
	if err != nil {
		return err
	}
	db.Exec("create extension if not exists \"uuid-ossp\";")
	db.AutoMigrate(&Machine{})
	db.AutoMigrate(&KV{})
	db.Close()

	h.setValue("db_version", dbVersion)
	return nil
}

func (h *Headscale) db() (*gorm.DB, error) {
	db, err := gorm.Open("postgres", h.dbString)
	if err != nil {
		return nil, err
	}
	return db, nil
}

func (h *Headscale) getValue(key string) (string, error) {
	db, err := h.db()
	if err != nil {
		return "", err
	}
	defer db.Close()
	var row KV
	if db.First(&row, "key = ?", key).RecordNotFound() {
		return "", errors.New("not found")
	}
	return row.Value, nil
}

func (h *Headscale) setValue(key string, value string) error {
	kv := KV{
		Key:   key,
		Value: value,
	}
	db, err := h.db()
	if err != nil {
		return err
	}
	defer db.Close()
	_, err = h.getValue(key)
	if err == nil {
		db.Model(&kv).Where("key = ?", key).Update("value", value)
		return nil
	}

	db.Create(kv)
	return nil
}
