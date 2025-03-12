FROM golang:1.24

WORKDIR /app

# Go modules ファイルをコピー
COPY go.mod go.sum ./
RUN go mod download

# ソースコードをコピー
COPY . .

# 依存ライブラリのインストールとビルド
#SQlite 実行のためにCGOは1である必要がある
RUN CGO_ENABLED=1 GOOS=linux go build -o server ./cmd/api/main.go

# 実行ユーザーの設定
RUN addgroup --system mercari && adduser --system --ingroup mercari trainee
RUN chown -R trainee:mercari /app

USER trainee

# ポートの公開
EXPOSE 9000

# 実行コマンド
CMD ["./server"]
