# Используем официальный образ Golang для сборки приложения
FROM golang:1.22 AS builder

WORKDIR /app

# Копируем go.mod и go.sum для установки зависимостей
COPY go.mod go.sum ./

# Устанавливаем зависимости
RUN go mod download

# Копируем исходный код
COPY . .

# Сборка приложения
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /todo_scheduler

# Используем образ с нужной версией GLIBC для запуска приложения
FROM ubuntu:latest

# Устанавливаем рабочую директорию внутри контейнера
WORKDIR /app

# Копируем собранное приложение и необходимые файлы
COPY --from=builder /todo_scheduler .
COPY --from=builder /app/web ./web

# Указываем порт, который будет использоваться приложением
EXPOSE 7540

# Устанавливаем переменные окружения
ENV TODO_PORT=7540
ENV TODO_DBFILE=/app/scheduler.db
ENV TODO_PASSWORD=your_password

# Указываем команду для запуска приложения
CMD ["/app/todo_scheduler"]