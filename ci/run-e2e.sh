#!/bin/bash
set -xe

wait_escrow() {
	while ! curl -v 'http://127.0.0.1:7100/info' | jq -e '.Escrows|length==1'; do sleep 10; done
}

for x in upstream{,-apache,-lighttpd,-caddy,-python}; do
	time make clean

	echo "[*] Starting network (with $x as upstream)..."
	PUBLISHER_UPSTREAM="http://$x:80" time docker-compose up -d

	echo "[*] Waiting until escrow is setup..."
	time wait_escrow

	echo "[*] Fetching from $x..."
	rm -f output.bin
	docker run --rm -v $PWD:/out --net=host cachecash/go-cachecash cachecash-curl -o /out/output.bin cachecash://localhost:7070/file1.bin
	diff -q output.bin testdata/content/file1.bin
	echo "[+] Success"

	time docker-compose down
done

echo "[+] All tests finished successfully"
