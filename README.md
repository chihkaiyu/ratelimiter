# Rate Limiter
This repo implements DCard homework. Implement a server that:
- Accept 60 requests per minute
- Response with current number of request
- Response error if exceed the limit
- You may use any database or design your own in-memory data structure
- Please implement it with tests
- Do not use any rate limit library

# Prerequisite
To run this repo, please prepare your environment with:
- Docker
- Docker compose
- Golang (if you'd like to build golang source)

# Build And Run
To build this repo, you may build the golang binary or docker image.
## Build Golang Binary
Simply run the command below at root folder of this repo:
```
go build -i -o ./build/app ./app
```
or (the binary would be located at `./build/app`)
```
make build-app
```

## Build Docker Image
Simply run the command below at root folder of this repo and please make sure you've already build the binary:
```
docker build -t "ratelimiter" .
```
or
```
make build-image
```

You could also simply build them at once:
```
make build-all
```

# Run
## From Source
If you'd like to run from golang binary, you have to prepare redis service by yourself.
Use docker would be a eaiser:
```
docker run -d redis:6.0.10-alpine -p 6379:6379 --name redis
./build/app
```

## From Docker Compose
Use docker compose would is easier since it start everything at once:
```
docker-compose -p ratelimiter up -d
```
or
```
make up-cluster
```

You may want to alter some config in `command` in `docker-compose.yml`. See Flags section for more detail.

# Test
If you'd like to test all:
```
go test ./...
```
or
```
make test
```

Change directory to the pakcage could test package separately:
```
cd ./service/redis
go test ./...
```

# Flags
There are some flags you could set for different purpose:

| Flag | Default value | Purpose |
| ---- | ------------- | ------- |
| env | dev | environment flag |
| port | 9000 | the port for API server listening to |
| redis_addr | localhost:6379 | the host and port of redis |
| ratelimiter_strategy | fixedwindow | rate limiter strategy, you could set: fixedwindow, slidingwindow, tokenbucket |
| fixed_window_size | 60 | window length, in second |
| fixed_window_limit | 60 | the number of requests could be accepted in a window |
| sliding_window_size | 60 | window length, in second |
| sliding_window_limit | 60 | the number of requests could be accepted in a window |
| bucketsize | 60 | the size of token bucket |
| refill_per_second | 1 | how many tokens to be refilled in one second |

# Strategy
## Fixed Window

## Sliding Window

## Token Bucket