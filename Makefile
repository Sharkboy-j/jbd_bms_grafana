.PHONY: pull
pull:
	git pull
.PHONY: logs
logs:
	sudo docker logs -f --tail 0 bms
	@$(MAKE) logs
.PHONY: build
build:
	git pull
	sudo docker build -t bms --no-cache .
	sudo docker kill bms
	sudo docker-compose up -d
	logs
.PHONY: run
run:
	sudo docker kill bms
	sudo docker-compose up -d
	logs
.PHONY: kill
kill:
	sudo docker kill bms
.PHONY: start
start:
	sudo docker-compose up -d