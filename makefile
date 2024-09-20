APP_NAME := server
APP_PATH := /home/roman/server
SERVICE_NAME := server.service
TMP_DIR := /home/roman/server/tmpData/inst

.PHONY: build install start stop restart status reload

# Сборка Go-приложения
build:
	@if [ ! -d "$(TMP_DIR)" ]; then \
		echo "Directory $(TMP_DIR) does not exist. Creating..."; \
		sudo mkdir -p $(TMP_DIR); \
	else \
		echo "Directory $(TMP_DIR) already exists."; \
	fi
	go version
	go build -o $(APP_NAME) .

# Установка скомпилированного бинарника в целевую директорию
install: build
	sudo cp $(APP_NAME) $(APP_PATH)
	sudo chmod +x $(APP_PATH)/$(APP_NAME)

reload:
	sudo cp $(SERVICE_NAME) /etc/systemd/system/
	sudo systemctl daemon-reload

# Запуск службы
start:
	sudo systemctl start $(SERVICE_NAME)

# Остановка службы
stop:
	sudo systemctl stop $(SERVICE_NAME)

# Перезапуск службы
restart:
	sudo systemctl restart $(SERVICE_NAME)

# Проверка статуса службы
status:
	sudo systemctl status $(SERVICE_NAME)
