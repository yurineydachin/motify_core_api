LAST_GOPATH_DIR:=$(lastword $(subst :, ,$(GOPATH)))
GOBIN:=$(LAST_GOPATH_DIR)/bin
GLIDE_PATH:=$(LAST_GOPATH_DIR)/src/github.com/Masterminds/glide
GLIDE_VERSION:=v0.12.3
GLIDE_BIN:=$(GOBIN)/glide-$(GLIDE_VERSION)
BIN:=./bin/echo
ROLLOUT_SETTER_BIN=./bin/rollout_setter
DATASYNC_APP_BIN:=./bin/datasync_app
DATASYNC_LISTENER_BIN:=./bin/datasync_listener

ETCD_VERSION=v3.2.4
TOKEN=etcd-discovery-token
CLUSTER_STATE=new
ETCD_STANDALONE_NAME=etcd
ETCD_NETWORK_NAME=etcd
ETCD_NETWORK_SUBNET=172.18.0.0/16
ETCD_NAME0=etcd0
ETCD_NAME1=etcd1
ETCD_NAME2=etcd2
CLUSTER=${ETCD_NAME0}=http://${ETCD_NAME0}:2380,${ETCD_NAME1}=http://${ETCD_NAME1}:2380,${ETCD_NAME2}=http://${ETCD_NAME2}:2380


etcd-run-standalone: etcd-rm-standalone
	docker run -d -p 2379:2379 \
		--name $(ETCD_STANDALONE_NAME) quay.io/coreos/etcd:$(ETCD_VERSION) \
		/usr/local/bin/etcd \
			-advertise-client-urls http://0.0.0.0:2379 \
			-listen-client-urls http://0.0.0.0:2379 \
			-auto-compaction-retention=2

etcd-create-network:
	if [ "`docker network ls --filter name=${ETCD_NETWORK_NAME} -q`" == "" ]; then \
		docker network create --subnet=${ETCD_NETWORK_SUBNET} ${ETCD_NETWORK_NAME}; \
	fi

etcd-run-cluster: etcd-create-network etcd-rm-cluster
	# For node 1
	docker run -d \
		-p 4000:2379 \
		--network=etcd \
		--name ${ETCD_NAME0} quay.io/coreos/etcd:${ETCD_VERSION} \
		/usr/local/bin/etcd \
			--name ${ETCD_NAME0} \
			--initial-advertise-peer-urls http://${ETCD_NAME0}:2380 --listen-peer-urls http://0.0.0.0:2380 \
			--advertise-client-urls http://0.0.0.0:4000 --listen-client-urls http://0.0.0.0:2379 \
			--initial-cluster ${CLUSTER} \
			--initial-cluster-state ${CLUSTER_STATE} --initial-cluster-token ${TOKEN}

	# For node 2
	docker run -d \
		-p 4001:2379 \
		--network=etcd \
		--name ${ETCD_NAME1} quay.io/coreos/etcd:${ETCD_VERSION} \
		/usr/local/bin/etcd \
			--name ${ETCD_NAME1} \
			--initial-advertise-peer-urls http://${ETCD_NAME1}:2380 --listen-peer-urls http://0.0.0.0:2380 \
			--advertise-client-urls http://0.0.0.0:4001 --listen-client-urls http://0.0.0.0:2379 \
			--initial-cluster ${CLUSTER} \
			--initial-cluster-state ${CLUSTER_STATE} --initial-cluster-token ${TOKEN}

	# For node 3
	docker run -d \
		-p 4002:2379 \
		--network=etcd \
		--name ${ETCD_NAME2} quay.io/coreos/etcd:${ETCD_VERSION} \
		/usr/local/bin/etcd \
			--name ${ETCD_NAME2} \
			--initial-advertise-peer-urls http://${ETCD_NAME2}:2380 --listen-peer-urls http://0.0.0.0:2380 \
			--advertise-client-urls http://0.0.0.0:4002 --listen-client-urls http://0.0.0.0:2379 \
			--initial-cluster ${CLUSTER} \
			--initial-cluster-state ${CLUSTER_STATE} --initial-cluster-token ${TOKEN}

etcd-start-standalone:
	docker start $(ETCD_STANDALONE_NAME)

etcd-stop-standalone:
	docker stop $(ETCD_STANDALONE_NAME)

etcd-rm-standalone:
	# ignore error if no container exists
	-docker rm -f $(ETCD_STANDALONE_NAME)

etcd-start-cluster:
	docker start $(ETCD_NAME0) $(ETCD_NAME1) $(ETCD_NAME2)

etcd-stop-cluster:
	docker stop $(ETCD_NAME0) $(ETCD_NAME1) $(ETCD_NAME2)

etcd-rm-cluster:
	# ignore error if no container exists
	-docker rm -f $(ETCD_NAME0) $(ETCD_NAME1) $(ETCD_NAME2)

core-deps:
	cd .. && $(MAKE) deps

deps: core-deps
	$(info #Install example dependencies...)
	$(GLIDE_BIN) install --force

build: deps fast-build

fast-build:
	$(info #Building binary to $(BIN)...)
	go build -o $(BIN) *.go
	go build -o $(ROLLOUT_SETTER_BIN) ./rollout_setter/*.go
	go build -o $(DATASYNC_APP_BIN) ./datasync/main_app.go
	go build -o $(DATASYNC_LISTENER_BIN) ./datasync/main_listener.go

set-rollout:
	$(ROLLOUT_SETTER_BIN) -rollout=unstable1 -from=1 -to=500

# run the test binary
run: set-rollout kill
	$(BIN) -host=localhost -port=8000 -service_name=alice -try_service=bob -etcd_endpoints=http://localhost:4000 > ./logs/alice_1.log 2>&1 &
	$(BIN) -host=localhost -port=8001 -service_name=alice -try_service=bob -etcd_endpoints=http://localhost:4001 > ./logs/alice_2.log 2>&1 &
	$(BIN) -host=localhost -port=8002 -service_name=alice -try_service=bob -etcd_endpoints=http://localhost:4002 > ./logs/alice_3.log 2>&1 &

	$(BIN) -host=localhost -port=9000 -service_name=bob -try_service=alice -etcd_endpoints=http://localhost:4000 > ./logs/bob_1.log 2>&1 &
	$(BIN) -host=localhost -port=9001 -service_name=bob -try_service=alice -etcd_endpoints=http://localhost:4001 > ./logs/bob_2.log 2>&1 &
	$(BIN) -host=localhost -port=9002 -service_name=bob -try_service=alice -etcd_endpoints=http://localhost:4002 > ./logs/bob_3.log 2>&1 &

	$(BIN) -host=localhost -port=10000 -service_name=bob -rollout_type=unstable1 -etcd_endpoints=http://localhost:4002 > ./logs/bob_unstable.log 2>&1 &

kill:
	ps aux | grep $(BIN) | egrep -v "grep|tail" | tr -s ' ' | cut -d ' ' -f 2 | xargs kill

run-datasync: kill-datasync
	$(DATASYNC_APP_BIN) -host=localhost -port=5000 -etcd_endpoints=http://localhost:4000 -exported_entities=foo,bar,buz > ./logs/datasync_app_1.log 2>&1 &
	$(DATASYNC_APP_BIN) -host=localhost -port=5001 -etcd_endpoints=http://localhost:4001 -exported_entities=one,two,three > ./logs/datasync_app_2.log 2>&1 &
	$(DATASYNC_APP_BIN) -host=localhost -port=5002 -etcd_endpoints=http://localhost:4002 -exported_entities=1,2,3 > ./logs/datasync_app_3.log 2>&1 &

	$(DATASYNC_LISTENER_BIN) -port=5500 -etcd_endpoints=http://localhost:4001 > ./logs/datasync_listener.log 2>&1 &

kill-datasync:
	ps aux | grep $(DATASYNC_APP_BIN) | egrep -v "grep|tail" | tr -s ' ' | cut -d ' ' -f 2 | xargs kill
	ps aux | grep $(DATASYNC_LISTENER_BIN) | egrep -v "grep|tail" | tr -s ' ' | cut -d ' ' -f 2 | xargs kill

# start etcd, build binary and run
run-build: etcd-run-cluster build run
