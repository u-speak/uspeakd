install:
	go install
node1: install
	API_PORT=4001 \
	STATIC_DIR=false \
	IMAGE_PATH=/tmp/i1.db \
	KEY_PATH=/tmp/k1.db \
	POST_PATH=/tmp/p1.db \
	NODE_PORT=5001 \
	${GOPATH}/bin/uspeakd --repl

node2: install
	API_PORT=4002 \
	STATIC_DIR=false \
	IMAGE_PATH=/tmp/i2.db \
	KEY_PATH=/tmp/k2.db \
	POST_PATH=/tmp/p2.db \
	NODE_PORT=5002 \
	${GOPATH}/bin/uspeakd --repl

node3: install
	API_PORT=4003 \
	STATIC_DIR=false \
	IMAGE_PATH=/tmp/i3.db \
	KEY_PATH=/tmp/k3.db \
	POST_PATH=/tmp/p3.db \
	NODE_PORT=5003 \
	${GOPATH}/bin/uspeakd --repl

clean:
	-rm /tmp/i1.db
	-rm /tmp/i2.db
	-rm /tmp/i3.db
	-rm /tmp/k1.db
	-rm /tmp/k2.db
	-rm /tmp/k3.db
	-rm /tmp/p1.db
	-rm /tmp/p2.db
	-rm /tmp/p3.db
