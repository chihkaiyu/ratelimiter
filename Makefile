.PHONY: build-app build-image build-all
build-app:
	docker run --rm -it \
	-v ${CURDIR}:/go/src/github.com/chihkaiyu/ratelimiter \
	-v ${CURDIR}/build:/app \
	-v ${GOPATH}/pkg/mod:/go/pkg/mod \
	"golang:1.13.5-alpine" sh -c "cd /go/src/github.com/chihkaiyu/ratelimiter/app && go build -i -o /app/app"

build-image:
	docker build -t "ratelimiter" .

build-all: build-app build-image

up-cluster:
	docker-compose -p ratelimiter up -d

down-cluster:
	docker-compose -p ratelimiter down

start-cluster:
	docker-compose -p ratelimiter start

stop-cluster:
	docker-compose -p ratelimiter stop

restart-cluster:
	docker-compose -p ratelimiter restart

test:
	go test ./...