.PHONY: pull
pull:
	git pull

.PHONY: logs
logs:
	sudo docker logs -f --tail 0 bms
	@$(MAKE) logs

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
	git pull
	sudo docker build -t bms --no-cache .
	@$(MAKE) run