# --- Этап 1: Сборка приложения ---
FROM golang:1.21-alpine AS builder

# Устанавливаем рабочую директорию внутри контейнера
WORKDIR /app

# Копируем файлы зависимостей и загружаем их
COPY go.mod go.sum ./
RUN go mod download

# Копируем остальной исходный код
COPY . .

# Собираем приложение. Флаги убирают отладочную информацию и уменьшают размер бинарника
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o main ./cmd/Wi-Fi-router-bandwidth-backend/main.go


# --- Этап 2: Создание финального легковесного образа ---
FROM alpine:latest

WORKDIR /app

# Копируем скомпилированный бинарник из этапа сборки
COPY --from=builder /app/main .

# Копируем статические файлы и шаблоны, необходимые для работы приложения
COPY templates/ ./templates/
COPY resources/ ./resources/

# Указываем, что наше приложение будет слушать порт 8080
EXPOSE 8080

# Команда для запуска нашего приложения при старте контейнера
CMD ["./main"]