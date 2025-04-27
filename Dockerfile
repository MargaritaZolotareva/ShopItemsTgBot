# Используем официальный образ Go
FROM golang:1.22

# Устанавливаем рабочую директорию внутри контейнера
WORKDIR /app

# Копируем все файлы проекта внутрь контейнера
COPY . .

RUN go get github.com/go-delve/delve/cmd/dlv
RUN go install github.com/go-delve/delve/cmd/dlv@latest
# Скачиваем зависимости и собираем бинарник
RUN go mod tidy && go build -o VkScraper cmd/server/main.go

RUN ls -l /app/VkScraper
