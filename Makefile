ES_TAG := "7.6.2"

DOCKER_NAME := "elastictest"

.DEFAULT_GOAL := build

build/%: cmd/%/*.go pkg/**/*.go
	GOOS=linux go build -o $@ ./cmd/$*/

.PHONY: build
build: build/check build/in build/out

.PHONY: docker
docker:
	docker build -t concourse-elasticsearch -f Dockerfile .

.PHONY: test
test:
	docker ps --format '{{ .Names }}' | grep $(DOCKER_NAME) > /dev/null \
	|| (docker ps --format '{{ .Names }}' | grep $(DOCKER_NAME) > /dev/null && docker start $(DOCKER_NAME)) \
	|| (docker run -d --name $(DOCKER_NAME) -p 9200:9200 -p 9300:9300 -e "discovery.type=single-node" docker.elastic.co/elasticsearch/elasticsearch:$(ES_TAG) && sleep 5)
	go test -coverprofile coverage.out ./pkg/**/ ./cmd/**/
	go tool cover -html=coverage.out -o coverage.html

clean:
	docker stop $(DOCKER_NAME) 2> /dev/null || true
	docker rm $(DOCKER_NAME) 2> /dev/null || true
	rm build/* || true
