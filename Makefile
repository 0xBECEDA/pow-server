build:
	docker build -f ./docker/Dockerfile -t app .

run: build
	docker compose -f docker/docker-compose.yml up

stop:
	docker compose -f docker/docker-compose.yml down

clean-test-cache:
	go clean -testcache

unit-tests: clean-test-cache
	go test ./internal/...

tests: clean-test-cache
	go test ./... -v