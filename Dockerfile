# Используем официальный образ Golang для сборки приложения
FROM golang:1.22 AS builder

WORKDIR /app

# Устанавливаем переменные окружения для сборки
ENV CGO_ENABLED=0
ENV GOOS=linux
ENV GOARCH=amd64

# Копируем go.mod и go.sum для установки зависимостей
COPY go.mod go.sum ./

# Устанавливаем зависимости
RUN go mod download

# Копируем исходный код
COPY . .

# Сборка приложения
RUN go build -o /app/todo_scheduler

# Используем легкий образ Alpine для запуска приложения
FROM alpine:latest

# Устанавливаем рабочую директорию внутри контейнера
WORKDIR /app

# Копируем собранное приложение и необходимые файлы
COPY --from=builder /app/todo_scheduler .
COPY --from=builder /app/web ./web

# Указываем порт, который будет использоваться приложением
EXPOSE 7540

# Устанавливаем переменные окружения
ENV TODO_PORT=7540
ENV TODO_DBFILE=/app/scheduler.db
ENV TODO_PASSWORD=your_password

# Указываем команду для запуска приложения
CMD ["/app/todo_scheduler"]