package app

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"go.uber.org/mock/gomock"
)

func TestParseAddItemRequest(t *testing.T) {
	t.Parallel()

	type wants struct {
		req *AddItemRequest
		err bool
	}

	// STEP 6-1: define test cases
	cases := map[string]struct {
		args map[string]string
		wants
	}{
		"ok: valid request": {
			args: map[string]string{
				"name":     "jacket",  // fill here
				"category": "fashion", // fill here
			},
			wants: wants{
				req: &AddItemRequest{
					Name:     "jacket",  // fill here
					Category: "fashion", // fill here
				},
				err: false,
			},
		},
		"ng: empty request": {
			args: map[string]string{},
			wants: wants{
				req: nil,
				err: true,
			},
		},
	}

	for name, tt := range cases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			// prepare request body
			values := url.Values{}
			for k, v := range tt.args {
				values.Set(k, v)
			}

			// prepare HTTP request
			req, err := http.NewRequest("POST", "http://localhost:9000/items", strings.NewReader(values.Encode()))
			if err != nil {
				t.Fatalf("failed to create request: %v", err)
			}
			req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

			// execute test target
			got, err := parseAddItemRequest(req)

			// confirm the result
			if err != nil {
				if !tt.err {
					t.Errorf("unexpected error: %v", err)
				}
				return
			}
			if diff := cmp.Diff(tt.wants.req, got); diff != "" {
				t.Errorf("unexpected request (-want +got):\n%s", diff)
			}
		})
	}
}

func TestHelloHandler(t *testing.T) {
	t.Parallel()

	// Please comment out for STEP 6-2
	//predefine what we want
	type wants struct {
		code int               // desired HTTP status code
		body map[string]string // desired body
	}
	want := wants{
		code: http.StatusOK,
		body: map[string]string{"message": "Hello, world!"},
	}

	// set up test
	req := httptest.NewRequest("GET", "/hello", nil)
	res := httptest.NewRecorder()

	h := &Handlers{}
	h.Hello(res, req)

	// STEP 6-2: confirm the status code
	if res.Code != want.code {
		t.Errorf("unexpected status code. want=%d, got=%d", want.code, res.Code)
	}
	// STEP 6-2: confirm response body
	var got map[string]string
	if err := json.Unmarshal(res.Body.Bytes(), &got); err != nil {
		t.Fatalf("failed to unmarshal response body: %v", err)
	}
	if diff := cmp.Diff(want.body, got); diff != "" {
		t.Errorf("unexpected response body (-want +got):\n%s", diff)
	}
}

func TestAddItem(t *testing.T) {
	t.Parallel()

	type wants struct {
		code int
		body map[string]string
	}
	cases := map[string]struct {
		args     map[string]string
		injector func(m *MockItemRepository)
		wants
	}{
		"ok: correctly inserted": {
			args: map[string]string{
				"name":     "used iPhone 16e",
				"category": "phone",
			},
			injector: func(m *MockItemRepository) {
				m.EXPECT().Insert(gomock.Any(), gomock.Any()).Return(nil)
			},
			wants: wants{
				code: http.StatusOK,
				body: map[string]string{"message": "item received: used iPhone 16e"},
			},
		},
		"ng: failed to insert": {
			args: map[string]string{
				"name":     "used iPhone 16e",
				"category": "phone",
			},
			injector: func(m *MockItemRepository) {
				item := &Item{
					Name:     "used iPhone 16e",
					Category: "phone",
					Image:    "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855.jpg",
				}
				m.EXPECT().Insert(gomock.Any(), item).Return(errors.New("failed to insert"))
			},
			wants: wants{
				code: http.StatusInternalServerError,
			},
		},
	}

	for name, tt := range cases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockIR := NewMockItemRepository(ctrl)
			tt.injector(mockIR)
			h := &Handlers{itemRepo: mockIR}

			values := url.Values{}
			for k, v := range tt.args {
				values.Set(k, v)
			}
			req := httptest.NewRequest("POST", "/items", strings.NewReader(values.Encode()))
			req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

			rr := httptest.NewRecorder()
			h.AddItem(rr, req)

			if tt.wants.code != rr.Code {
				t.Errorf("expected status code %d, got %d", tt.wants.code, rr.Code)
			}
			if tt.wants.code >= 400 {
				return
			}

			var got map[string]string
			if err := json.Unmarshal(rr.Body.Bytes(), &got); err != nil {
				t.Fatalf("failed to unmarshal response body: %v", err)
			}
			if diff := cmp.Diff(tt.wants.body, got); diff != "" {
				t.Errorf("unexpected response body (-want +got):\n%s", diff)
			}
		})
	}
}

// STEP 6-4: uncomment this test
func TestAddItemE2e(t *testing.T) {
	// 環境によっては E2E テストをスキップできるようにする
	if testing.Short() {
		t.Skip("skipping e2e test")
	}

	// テスト用のDBをセットアップ
	db, closers, err := setupDB(t)
	if err != nil {
		t.Fatalf("failed to set up database: %v", err)
	}
	// テスト終了後にクリーンアップ
	t.Cleanup(func() {
		for _, c := range closers {
			c()
		}
	})

	type wants struct {
		code int
		body map[string]string
	}
	cases := map[string]struct {
		args map[string]string
		wants
	}{
		"ok: correctly inserted": {
			args: map[string]string{
				"name":     "used iPhone 16e",
				"category": "phone",
			},
			wants: wants{
				code: http.StatusOK,
				body: map[string]string{"message": "item received: used iPhone 16e"},
			},
		},
		"ng: failed to insert": {
			args: map[string]string{
				"name":     "",
				"category": "phone",
			},
			wants: wants{
				code: http.StatusBadRequest,
			},
		},
	}

	for name, tt := range cases {
		t.Run(name, func(t *testing.T) {
			h := &Handlers{
				itemRepo: NewItemRepositoryWithDB(db),
			}

			values := url.Values{}
			for k, v := range tt.args {
				values.Set(k, v)
			}
			req := httptest.NewRequest("POST", "/items", strings.NewReader(values.Encode()))
			req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

			rr := httptest.NewRecorder()
			h.AddItem(rr, req)

			if tt.wants.code != rr.Code {
				t.Errorf("expected status code %d, got %d", tt.wants.code, rr.Code)
			}
			if tt.wants.code >= 400 {
				return
			}

			var got map[string]string
			if err := json.Unmarshal(rr.Body.Bytes(), &got); err != nil {
				t.Fatalf("failed to unmarshal response body: %v", err)
			}
			if diff := cmp.Diff(tt.wants.body, got); diff != "" {
				t.Errorf("unexpected response body (-want +got):\n%s", diff)
			}
		})
	}
}

func setupDB(t *testing.T) (db *sql.DB, closers []func(), e error) {
	t.Helper()

	defer func() {
		if e != nil {
			for _, c := range closers {
				c()
			}
		}
	}()

	// create a temporary file for e2e testing
	f, err := os.CreateTemp(".", "*.sqlite3")
	if err != nil {
		return nil, nil, err
	}
	closers = append(closers, func() {
		f.Close()
		os.Remove(f.Name())
	})

	// set up tables
	db, err = sql.Open("sqlite3", f.Name())
	if err != nil {
		return nil, nil, err
	}
	closers = append(closers, func() {
		db.Close()
	})

	// TODO: replace it with real SQL statements.
	cmd := `
		CREATE TABLE IF NOT EXISTS "items" (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL,
			category_id INTEGER,
			image_name TEXT,
			FOREIGN KEY (category_id) REFERENCES categories(id)
		);

		CREATE TABLE categories (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL
		);

		`

	_, err = db.Exec(cmd)
	if err != nil {
		return nil, nil, err
	}

	return db, closers, nil
}
