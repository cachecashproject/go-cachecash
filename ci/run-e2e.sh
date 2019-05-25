#!/bin/bash
set -xe

for x in upstream{,-apache,-lighttpd,-caddy,-python}; do
	make clean

	echo "[*] Starting network (with $x as upstream)..."
	PUBLISHER_UPSTREAM="http://$x:80" docker-compose up -d

	echo "[*] Waiting until escrow is setup..."
	while ! curl -v 'http://127.0.0.1:7100/info' | jq -e '.Escrows|length==1'; do sleep 10; done

	echo "[*] Fetching from $x..."
	rm -f output.bin
	docker run --rm -v $PWD:/out --net=host cachecash/go-cachecash cachecash-curl -o /out/output.bin cachecash://localhost:8080/file1.bin
	diff -q output.bin testdata/content/file1.bin
	echo "[+] Success"

	docker-compose down
done

echo "[+] All tests finished successfully"
