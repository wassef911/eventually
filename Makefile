.PHONY:

dev:
	@echo Starting dev docker compose
	docker-compose -f docker-compose.yaml up -d --build
	docker-compose -f docker-compose.yaml logs -f app

test:
	@echo Starting test 
	go test -v ./... 

clean:
	docker system prune -f

deps-reset:
	git checkout -- go.mod
	go mod tidy

deps-upgrade:
	go get -u -t -d -v ./...
	go mod tidy

deps-cleancache:
	go clean -modcache

lint:
	@echo Starting linters
	golangci-lint run ./...

swagger:
	swag init --parseDependency --parseInternal -g **/**/*.go

