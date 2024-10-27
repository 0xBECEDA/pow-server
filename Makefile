build:
	docker build -f ./docker/Dockerfile -t app .

run: build
	docker compose -f docker/docker-compose.yml up

stop:
	docker compose -f docker/docker-compose.yml down
