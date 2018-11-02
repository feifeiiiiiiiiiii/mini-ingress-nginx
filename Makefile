all: push

VERSION = edge
TAG = $(VERSION)
PREFIX = nginx/nginx-ingress

DOCKER_RUN = docker run --rm -v $(shell pwd):/go/src/github.com/feifeiiiiiiiiiii/mini-ingress-nginx
DOCKER_BUILD_RUN = docker run --rm -v $(shell pwd):/go/src/github.com/feifeiiiiiiiiiii/mini-ingress-nginx -w /go/src/github.com/feifeiiiiiiiiiii/mini-ingress-nginx/cmd/nginx-ingress/
GOLANG_CONTAINER = golang:1.10
DOCKERFILEPATH = build
DOCKERFILE = Dockerfile # note, this can be overwritten e.g. can be DOCKERFILE=DockerFileForPlus

BUILD_IN_CONTAINER = 1
PUSH_TO_GCR =
GENERATE_DEFAULT_CERT_AND_KEY =
DOCKER_BUILD_OPTIONS =

GIT_COMMIT=$(shell git rev-parse --short HEAD)

nginx-ingress:
ifeq ($(BUILD_IN_CONTAINER),1)
	$(DOCKER_BUILD_RUN) -e CGO_ENABLED=0 $(GOLANG_CONTAINER) go build -installsuffix cgo -ldflags "-w -X main.version=${VERSION} -X main.gitCommit=${GIT_COMMIT}" -o /go/src/github.com/feifeiiiiiiiiiii/mini-ingress-nginx/nginx-ingress
else
	CGO_ENABLED=0 GOOS=linux go build -installsuffix cgo -ldflags "-w -X main.version=${VERSION} -X main.gitCommit=${GIT_COMMIT}" -o nginx-ingress github.com/feifeiiiiiiiiiii/mini-ingress-nginx/cmd/nginx-ingress
endif

container: nginx-ingress
	cp $(DOCKERFILEPATH)/$(DOCKERFILE) ./Dockerfile
	docker build $(DOCKER_BUILD_OPTIONS) -f Dockerfile -t $(PREFIX):$(TAG) .

push: container
	docker push $(PREFIX):$(TAG)

clean:
	rm -f nginx-ingress
	rm -f Dockerfile
