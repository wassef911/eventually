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

pprof_heap:
	go tool pprof -http :8006 http://localhost:6060/debug/pprof/heap?seconds=10

pprof_cpu:
	go tool pprof -http :8006 http://localhost:6060/debug/pprof/profile?seconds=10

pprof_allocs:
	go tool pprof -http :8006 http://localhost:6060/debug/pprof/allocs?seconds=10

swagger:
	swag init --parseDependency --parseInternal -g **/**/*.go

