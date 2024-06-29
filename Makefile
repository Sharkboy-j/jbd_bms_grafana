.PHONY: pull
pull:
	git pull
.PHONY: logs
logs:
	@while true; do \
        sudo docker logs -f --tail 10 bms \
    done
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
