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
	@$(MAKE) start
	@$(MAKE) logs

.PHONY: build
build:
	@$(MAKE) pull
	sudo docker build -t bms --no-cache .
	@$(MAKE) run