package app

import (
	"context"
	"database/sql"
	"errors"
	"log/slog"
	"sync"

	_ "github.com/mattn/go-sqlite3"
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
	GetAll(ctx context.Context) ([]*Item, error)
	GetByID(ctx context.Context, id int) (*Item, error)
	SearchByKeyword(ctx context.Context, keyword string) ([]*Item, error)
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

var (
	db   *sql.DB
	once sync.Once
)

// DBを読み込む
func getDB() *sql.DB {
	once.Do(func() { //１回だけDBを開く
		var err error
		db, err = sql.Open("sqlite3", "db/mercari.sqlite3")
		if err != nil {
			slog.Error("failed to connect to database", "error", err)
		} else {
			slog.Info("successfully connected to database")
		}
	})
	return db
}

// Insert inserts an item into the repository.
func (i *itemRepository) Insert(ctx context.Context, item *Item) error {
	db := getDB()

	query := "INSERT INTO items (name, category, image_name) VALUES (?, ?, ?)"
	_, err := db.ExecContext(ctx, query, item.Name, item.Category, item.Image)

	if err != nil {
		slog.Error("failed to insert item", "error", err)
		return err
	}

	return nil
}

// GetAll：items.jsonから全商品を取得
func (i *itemRepository) GetAll(ctx context.Context) ([]*Item, error) {
	db := getDB()

	rows, err := db.QueryContext(ctx, "SELECT id, name, category, image_name FROM items")

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var items []*Item
	for rows.Next() {
		var item Item
		err := rows.Scan(&item.ID, &item.Name, &item.Category, &item.Image)
		if err != nil {
			return nil, err
		}
		items = append(items, &item)
	}

	return items, nil
}

// IDから特定の商品を取得
func (r *itemRepository) GetByID(ctx context.Context, id int) (*Item, error) {
	db := getDB()

	row := db.QueryRowContext(ctx, "SELECT id, name, category, image_name FROM items WHERE id = ?", id)

	var item Item
	err := row.Scan(&item.ID, &item.Name, &item.Category, &item.Image)
	if err != nil {
		return nil, err
	}

	return &item, nil
}

// items.jsonを読み込み
// func (i *itemRepository) loadItems() ([]Item, error) {
// 	file, err := os.OpenFile(i.fileName, os.O_RDWR|os.O_CREATE, 0644)
// 	if err != nil {
// 		return nil, err
// 	}
// 	defer file.Close()

// 	var items []Item
// 	err = json.NewDecoder(file).Decode(&items)
// 	if err != nil && err.Error() != "EOF" {
// 		return nil, err
// 	}

// 	return items, nil
// }

// StoreImage stores an image and returns an error if any.
// This package doesn't have a related interface for simplicity.
func StoreImage(fileName string, image []byte) error {
	// STEP 4-4: add an implementation to store an image

	return nil
}

func (r *itemRepository) SearchByKeyword(ctx context.Context, keyword string) ([]*Item, error) {
	db := getDB()

	// ここではキーワードが無いなら エラーメッセージを返す
	if keyword == "" {
		return nil, errors.New("keyword is required")
	}

	// LIKE で検索機能を実装('%' || ? || '%' で部分一致もできる)
	rows, err := db.QueryContext(ctx, `
        SELECT id, name, category, image_name
          FROM items
         WHERE name LIKE '%' || ? || '%'
    `, keyword)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []*Item
	for rows.Next() {
		var item Item
		if err := rows.Scan(&item.ID, &item.Name, &item.Category, &item.Image); err != nil {
			return nil, err
		}
		items = append(items, &item)
	}

	// 見つからない場合はエラーを返す
	if len(items) == 0 {
		return nil, errors.New("no items found matching the keyword")
	}
	return items, nil
}
