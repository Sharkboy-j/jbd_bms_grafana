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
	@$(MAKE) start
	@$(MAKE) logs

.PHONY: build
build:
	@$(MAKE) pull
	go mod download
	GOARCH=arm GOARM=7 GOOS=linux go build -o bleTest .
	chmod +x bleTest
	sudo docker build -t bms --no-cache .
	@$(MAKE) run