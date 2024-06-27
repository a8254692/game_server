APP_ENV=
GOOS = linux
GOARCH = amd64

.PHONY: build
# go build
build:
	go mod tidy \
	&& cd OtherServer \
	&& go build \
	&& cd .. \
	&& cd DBServer \
	&& go build \
	&& cd .. \
	&& cd GameServer \
	&& go build \
	&& cd .. \
	&& cd GateWsServer \
	&& go build \
	&& cd .. \
	&& cd WebServer \
	&& go build

.PHONY: pack
# pack
pack:
	@echo "------------------->>>>  PLEASE ENTER THE APP_ENV=dev|prod  <<<<-------------------"
	@echo "----------------------------------------------------------------------------------------------------"
	@echo "--------------------------->>>>  THE APP_ENV IS : $(APP_ENV)  <<<<---------------------------"
	@echo "----------------------------------------------------------------------------------------------------"

	mkdir -p billiard/conf \
	&& cp Common/table/table.json billiard/conf  \
	&& cp Common/conf/$(APP_ENV)/server.ini billiard/conf  \
    && cp WebServer/conf/$(APP_ENV)/app.conf billiard/conf  \
	&& sed -i 's/\r//' start.sh  \
	&& cp start.sh billiard/  \
	&& sed -i 's/\r//' stop.sh  \
	&& cp stop.sh billiard/  \
	&& cd OtherServer \
	&& cp OtherServer ../billiard/  \
	&& cd .. \
	&& cd DBServer \
	&& cp DBServer ../billiard/  \
	&& cd .. \
	&& cd GameServer \
	&& cp GameServer ../billiard/  \
	&& cd .. \
	&& cd GateWsServer \
	&& cp GateWsServer ../billiard/  \
	&& cd .. \
	&& cd WebServer \
	&& cp WebServer ../billiard/  \

# show help
help:
	@echo ''
	@echo 'Usage:'
	@echo ' make [target]'
	@echo ''
	@echo 'Targets:'
	@awk '/^[a-zA-Z\-\_0-9]+:/ { \
	helpMessage = match(lastLine, /^# (.*)/); \
		if (helpMessage) { \
			helpCommand = substr($$1, 0, index($$1, ":")-1); \
			helpMessage = substr(lastLine, RSTART + 2, RLENGTH); \
			printf "\033[36m%-22s\033[0m %s\n", helpCommand,helpMessage; \
		} \
	} \
	{ lastLine = $$0 }' ${MAKEFILE_LIST}

.DEFAULT_GOAL := help
