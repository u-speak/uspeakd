install:
	go install
node1: install
	API_PORT=4001 \
	STATIC_DIR=false \
	DATA_PATH=/tmp/n1.data.db \
	TANGLE_PATH=/tmp/n1.tangle.db \
	NODE_PORT=5001 \
	DIAG_PORT=1337 \
	${GOPATH}/bin/uspeakd --repl

node2: install
	API_PORT=4002 \
	STATIC_DIR=false \
	DATA_PATH=/tmp/n2.data.db \
	TANGLE_PATH=/tmp/n2.tangle.db \
	NODE_PORT=5002 \
	DIAG_PORT=1338 \
	${GOPATH}/bin/uspeakd --repl

node3: install
	API_PORT=4003 \
	STATIC_DIR=false \
	DATA_PATH=/tmp/n2.data.db \
	TANGLE_PATH=/tmp/n2.tangle.db \
	NODE_PORT=5003 \
	${GOPATH}/bin/uspeakd --repl

clean:
	-rm /tmp/n1.data.db
	-rm /tmp/n1.tangle.db
	-rm /tmp/n2.data.db
	-rm /tmp/n2.tangle.db
	-rm /tmp/n3.data.db
	-rm /tmp/n3.tangle.db
