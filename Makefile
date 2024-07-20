AUTHENTICATION_BINARY=authenticationApp
BROKER_BINARY=brokerApp
CONFIG_BINARY=configApp
FILE_CONFIG_BINARY=fileConfigApp
MONITORING_BINARY=monitoringApp

up_build:
	sudo docker-compose down
	sudo docker-compose up --build

up:
	sudo docker-compose up

down:
	sudo docker-compose down

build_authentication:
	cd ./authentication-service && env GOOS=linux CGO_ENABLED=0 go build -o ${AUTHENTICATION_BINARY} .

build_broker:
	cd ./broker-service && env GOOS=linux CGO_ENABLED=0 go build -o ${BROKER_BINARY} .

build_config:
	cd ./config-service && env GOOS=linux CGO_ENABLED=0 go build -o ${CONFIG_BINARY} .

build_file_config:
	cd ./file-config-service && env GOOS=linux CGO_ENABLED=0 go build -o ${FILE_CONFIG_BINARY} .

build_monitoring:
	cd ./monitoring-service && env GOOS=linux CGO_ENABLED=0 go build -o ${MONITORING_BINARY} .