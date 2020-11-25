#ARCH
ARCH="`uname -s`"
LINUX="Linux"
Darwin="Darwin"
env="local"
repo=""

gomod:
    export GO111MODULE=on

mkdir:
	rm -rf $(GOPATH)/src/devops
	mkdir -p $(GOPATH)/src/devops

clone: mkdir
	cd $(GOPATH)/src/devops
	git clone git@github.com:woodliu/PrometheusMiddleware.git

swag: gomod
	go env -w GOPROXY=https://goproxy.cn,direct
	go get github.com/swaggo/swag/cmd/swag

doc: swag
	rm -rf docs && swag init

build:
	@if [ $(ARCH) = $(LINUX) ]; \
    	then \
    		go build -o prometheusservice -tags 'netgo osusergo' -ldflags '-extldflags "-static"' main.go; \
    	elif [ $(ARCH) = $(Darwin) ]; \
    	then \
    		GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o prometheusservice -ldflags '-s -extldflags "-sectcreate __TEXT __info_plist Info.plist"' main.go; \
    	else \
    		echo "ARCH unknow"; \
    	fi

docker: build
	docker build -t $(repo)-$(env)/prometheusservice:$(tag) .

push:
	docker push $(repo)-$(env)/prometheusservice:$(tag)

rmImage:
	docker rmi -f $(repo)-$(env)/prometheusservice:$(tag)
