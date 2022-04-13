.PHONY: run

ADDRESS ?= localhost:5701
MAP_NAME ?= map1
USE_PUBLIC_IP ?= false

run:
	go run -tags hazelcastinternal main.go -ad=${ADDRESS} -n=${MAP_NAME} -p=${USE_PUBLIC_IP}