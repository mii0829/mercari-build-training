FROM node:20-alpine

# 作業ディレクトリの設定
WORKDIR /app

# パッケージのインストール
COPY package*.json ./
RUN npm install

# ソースコードのコピー
COPY . .

# ビルド
RUN npm run build

# 実行ユーザーの設定
RUN addgroup -S mercari && adduser -S trainee -G mercari
RUN chown -R trainee:mercari /app
USER trainee

# ポートの公開
EXPOSE 3000

# 実行コマンド
# --host 0.0.0.0：Docker 内の Vite サーバーが外部アクセスを許可するため。
CMD ["npm", "start", "--", "--host", "0.0.0.0", "--port", "3000"]
