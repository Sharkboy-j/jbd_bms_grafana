.PHONY: pull
pull:
	git pull

.PHONY: build
build:
	git pull
	sudo docker build -t bms --no-cache .
	sudo docker kill bms
	sudo docker-compose up -d
	sudo docker logs -f bms

.PHONY: run
run:
	sudo docker kill bms
	sudo docker-compose up -d
	sudo docker logs -f --tail 10 bms

.PHONY: logs
logs:
	sudo docker logs -f --tail 10 bms

.PHONY: kill
kill:
	sudo docker kill bms

.PHONY: start
start:
	sudo docker-compose up -d

.PHONY: pull
pull:
	git pull