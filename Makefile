MAKEFLAGS += --no-print-directory

.PHONY: pull
pull:
	git pull

.PHONY: logs
logs:
	@./logs.sh
.PHONY: kill
kill:
	sudo docker kill bms

.PHONY: start
start:
	sudo docker-compose up -d

.PHONY: run
run:
	@$(MAKE) kill
	@sleep 3
	@$(MAKE) start
	@$(MAKE) logs

.PHONY: build
build:
	@$(MAKE) pull
	go mod download
	go build -o bleTest .
	sudo docker build -t bms --no-cache .
	@$(MAKE) run