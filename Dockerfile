# -------- STAGE 1: BUILD --------
FROM golang:1.23-alpine AS builder

# Создаем рабочую директорию
WORKDIR /app

# Копируем go.mod и go.sum отдельно для кеширования зависимостей
COPY go.mod go.sum ./
RUN go mod download

# Копируем весь проект
COPY . ./

# Собираем бинарник из cmd/main.go
RUN go build -o server ./cmd/main.go

# -------- STAGE 2: RUN --------
FROM debian:bullseye-slim

# Устанавливаем Chromium для Chromedp
RUN apt-get update && apt-get install -y --no-install-recommends \
    chromium \
    ca-certificates \
    && rm -rf /var/lib/apt/lists/*

# Создаем рабочую директорию
WORKDIR /app

# Копируем скомпилированный бинарник
COPY --from=builder /app/server /app/server

# Копируем шаблоны (templates), если нужны в рантайме
COPY --from=builder /app/templates /app/templates

# При необходимости скопируйте и другие нужные файлы (static, etc.)

# Открываем порт (пример: 8080)
EXPOSE 8080

# Точка входа
ENTRYPOINT ["/app/server"]