.PHONY: build-app build-image build-all
build-app:
	docker run --rm -it \
	-v ${CURDIR}:/go/src/github.com/chihkaiyu/dcard-homework \
	-v ${CURDIR}/build:/app \
	-v ${GOPATH}/pkg/mod:/go/pkg/mod \
	"golang:1.13.5-alpine" sh -c "cd /go/src/github.com/chihkaiyu/dcard-homework/app && go build -i -o /app/app"

build-image:
	docker build -t "dcard-homework" .

build-all: build-app build-image

up-cluster:
	docker-compose -p dcard-homework up -d

down-cluster:
	docker-compose -p dcard-homework down

start-cluster:
	docker-compose -p dcard-homework start

stop-cluster:
	docker-compose -p dcard-homework stop

restart-cluster:
	docker-compose -p dcard-homework restart