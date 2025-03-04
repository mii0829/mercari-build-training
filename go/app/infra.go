package app

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	// STEP 5-1: uncomment this line
	// _ "github.com/mattn/go-sqlite3"
)

var errImageNotFound = errors.New("image not found")

type Item struct {
	ID       int    `db:"id" json:"-"`
	Name     string `db:"name" json:"name"`
	Category string `db:"category" json:"category"`
	Image    string `db:"image" json:"image"`
}

// Please run `go generate ./...` to generate the mock implementation
// ItemRepository is an interface to manage items.
//
//go:generate go run go.uber.org/mock/mockgen -source=$GOFILE -package=${GOPACKAGE} -destination=./mock_$GOFILE
type ItemRepository interface {
	Insert(ctx context.Context, item *Item) error
	GetAll(ctx context.Context) ([]Item, error)
}

// itemRepository is an implementation of ItemRepository
type itemRepository struct {
	// fileName is the path to the JSON file storing items.
	fileName string
}

// NewItemRepository creates a new itemRepository.
func NewItemRepository() ItemRepository {
	return &itemRepository{fileName: "items.json"}
}

// Insert inserts an item into the repository.
func (i *itemRepository) Insert(ctx context.Context, item *Item) error {
	// STEP 4-1: add an implementation to store an item

	items, err := i.loadItems()
	if len(items) > 0 {
		item.ID = items[len(items)-1].ID + 1
	} else {
		item.ID = 1
	}

	if err != nil {
		return err
	}

	items = append(items, *item)

	return i.saveItems(items)
}

// GetAll：items.jsonから全商品を取得
func (i *itemRepository) GetAll(ctx context.Context) ([]Item, error) {
	return i.loadItems()
}

// items.jsonを読み込み
func (i *itemRepository) loadItems() ([]Item, error) {
	file, err := os.OpenFile(i.fileName, os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var items []Item
	err = json.NewDecoder(file).Decode(&items)
	if err != nil && err.Error() != "EOF" {
		return nil, err
	}

	return items, nil
}

// items.jsonに保存
func (i *itemRepository) saveItems(items []Item) error {
	file, err := os.Create(i.fileName)
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	return encoder.Encode(items)
}

// StoreImage stores an image and returns an error if any.
// This package doesn't have a related interface for simplicity.
func StoreImage(fileName string, image []byte) error {
	// STEP 4-4: add an implementation to store an image

	return nil
}
