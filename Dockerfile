# Этап сборки: используем официальный образ golang для сборки приложения.
FROM golang:1.23 AS builder
WORKDIR /app

# Копируем файлы модулей и скачиваем зависимости.
COPY go.mod go.sum ./
RUN go mod download

# Копируем исходный код и собираем бинарник.
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64  go build -o cv-generator main.go

# Финальный образ: Alpine с установленным Chromium.
FROM alpine:latest
# Обновляем репозитории и устанавливаем Chromium и необходимые библиотеки.
RUN apk update && apk add --no-cache chromium nss freetype harfbuzz ca-certificates ttf-freefont

# Указываем путь к Chromium для chromedp.
ENV CHROME_BIN=/usr/bin/chromium-browser

WORKDIR /app
# Копируем бинарный файл из билдера.
COPY --from=builder /app/ .

# Открываем порт (по умолчанию наш сервер слушает 8080)
EXPOSE 8080

ENTRYPOINT ["./cv-generator"]