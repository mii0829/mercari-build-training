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
	CategoryInsert(ctx context.Context, categoryName string) (int, error)
	Insert(ctx context.Context, item *Item) error
	GetAll(ctx context.Context) ([]*Item, error)
	GetByID(ctx context.Context, id int) (*Item, error)
	SearchByKeyword(ctx context.Context, keyword string) ([]*Item, error)
}

// itemRepository is an implementation of ItemRepository
type itemRepository struct {
	// fileName is the path to the JSON file storing items.
	fileName string
	db       *sql.DB
}

// NewItemRepository creates a new itemRepository.
func NewItemRepository() ItemRepository {
	return &itemRepository{
		fileName: "items.json",
		db:       getDB(),
	}
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

func (i *itemRepository) CategoryInsert(ctx context.Context, categoryName string) (int, error) {
	db := i.db

	//既存のカテゴリIDを探す
	var catID int
	err := db.QueryRowContext(ctx,
		`SELECT id FROM categories WHERE name =?`,
		categoryName,
	).Scan(catID)

	if err == sql.ErrNoRows {
		//既存IDが見つからなければINSERTする
		res, err := db.ExecContext(ctx, `INSERT INTO categories (name) VALUES (?)`, categoryName)
		if err != nil {
			return 0, err
		}
		lastID, err := res.LastInsertId()
		if err != nil {
			return 0, err
		}
		catID = int(lastID)
	} else if err != nil {
		//それ以外のエラーはそのまま返す
		return 0, err
	}
	return catID, nil
}

// Insert inserts an item into the repository.
func (i *itemRepository) Insert(ctx context.Context, item *Item) error {
	db := i.db

	catID, err := i.CategoryInsert(ctx, item.Category)
	if err != nil {
		slog.Error("failed to CategoryInsert", "error", err)
		return err
	}

	query := `INSERT INTO items (name, category_id, image_name) VALUES (?, ?, ?)`
	_, err = db.ExecContext(ctx, query, item.Name, catID, item.Image)

	if err != nil {
		slog.Error("failed to insert item", "error", err)
		return err
	}

	return nil
}

// GetAll：items.jsonから全商品を取得
func (i *itemRepository) GetAll(ctx context.Context) ([]*Item, error) {
	db := i.db

	//JOINでcategoryとitemテーブルをつなげて取得する
	rows, err := db.QueryContext(ctx, `
        SELECT i.id, i.name, c.name AS category, i.image_name
          FROM items i
          JOIN categories c ON i.category_id = c.id
    `)
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
	db := r.db

	row := db.QueryRowContext(ctx, `
        SELECT i.id, i.name, c.name AS category, i.image_name
          FROM items i
          JOIN categories c ON i.category_id = c.id
         WHERE i.id = ?
    `, id)

	var item Item
	err := row.Scan(&item.ID, &item.Name, &item.Category, &item.Image)
	if err != nil {
		return nil, err
	}

	return &item, nil
}

// StoreImage stores an image and returns an error if any.
// This package doesn't have a related interface for simplicity.
func StoreImage(fileName string, image []byte) error {
	// STEP 4-4: add an implementation to store an image

	return nil
}

func (r *itemRepository) SearchByKeyword(ctx context.Context, keyword string) ([]*Item, error) {
	db := r.db

	// ここではキーワードが無いなら エラーメッセージを返す
	if keyword == "" {
		return nil, errors.New("keyword is required")
	}

	// LIKE で検索機能を実装('%' || ? || '%' で部分一致もできる)
	rows, err := db.QueryContext(ctx, `
        SELECT i.id, i.name, c.name AS category, i.image_name
          FROM items i
          JOIN categories c ON i.category_id = c.id
         WHERE i.name LIKE '%' || ? || '%'
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

func NewItemRepositoryWithDB(db *sql.DB) ItemRepository {
	return &itemRepository{
		fileName: "items.json",
		db:       db,
	}
}
