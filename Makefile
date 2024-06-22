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
	sudo docker logs -f bms