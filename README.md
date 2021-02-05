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
```shell
go build -i -o ./build/app ./app
```
or (the binary would be located at `./build/app`)
```shell
make build-app
```

## Build Docker Image
Simply run the command below at root folder of this repo and please make sure you've already build the binary:
```shell
docker build -t "ratelimiter" .
```
or
```shell
make build-image
```

You could also simply build them at once:
```shell
make build-all
```

# Run
## From Source
If you'd like to run from golang binary, you have to prepare redis service by yourself.
Use docker would be a eaiser:
```shell
docker run -d redis:6.0.10-alpine -p 6379:6379 --name redis
./build/app
```

## From Docker Compose
Use docker compose would is easier since it start everything at once:
```shell
docker-compose -p ratelimiter up -d
```
or
```shell
make up-cluster
```

You may want to alter some config in `command` in `docker-compose.yml`. See Flags section for more detail.

# Test
## Unit Test
If you'd like to test source code:
```shell
go test ./...
```
or
```shell
make test
```

Change directory to the pakcage could test package separately:
```shell
cd ./service/redis
go test ./...
```

## Rate Limiter Test
Send HTTP request to API server to test rate limiter.  
### Request
```
GET /api/v1/ping
```

### Response
Status code: 200
```json
{
    "current_request_count": 1
}
```
Status code: 429
```json
{
    "error": "too many request"
}
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

# Strategy Analysis
I've implemented 3 strategies for rate limiting: fixed window, sliding window and token bucket.  
Annotation
- N: the number of different IP
- L: the limit of request (only in fixed window and sliding window)
- S: window size or bucket sizewindow)
- R: token refill rate (only in token bucket)
## Fixed Window
We split the time series into multiple non-overlap window. At every window, server only accepts at most N requests which N is the limit can be set by flag.  
### Burst
Burst: 2L, L is the limit of request.  
In this strategy, the burst is double of limit. For example, window size is 60 seconds and limit is 60 requests. We send 60 requests concurrently at 00:00:59 within one seconds and server should accept them all. At next second, we could send another 60 requests concurrently again. Within 2 seconds, we have sent 120 requests to server.  

### Space Complexity
O(N), which N is the number of different IP.  
We have to track the number of request of every IP at current window. We could expire the key-value pair immediately after the window passed.

## Sliding Window
When a request comes in, we count the number of request between now-S and now which S the window size. Deny when the count is greater than L or accept the request.
### Burst
Burst: L, there is no burst in sliding window.  
At any time, the number of requests in the window won't exceed L.

### Space Complexity
O(N*L), which N is the number of different IP and L the limit.  
We have to track the number of requests of every IP and sort them by the timestamp. The sorted set contains at most L records.

## Token Bucket
We accept the request when it could take one token from the bucket. We also put R tokens into bucket which R is the token refill rate and we stop refill when the number of token exceed S which S is the bucket size.  

### Burst
Burst: S+R*window size, which S is the bucket size.  
For example, the limit is L requests in X seconds and the maximum number of concurrent requests would be S+X*R. The burst occurs or not depends on the settings.

### Space Complexity
O(N), which N is the number of different IP.  
We have to track the number of token in the bucket and the timestamp of last request for every different IP.