#!/bin/bash
set -xe

for x in upstream{,-apache,-lighttpd,-caddy,-python}; do
	make clean

	echo "[*] Generate config for $x..."
	docker run --rm -v $PWD/cfg:/cfg cachecash/go-cachecash generate-config -outputPath /cfg/ -upstream "http://$x:80" -publisherCacheServiceAddr publisher:8082

	echo "[*] Starting network..."
	docker-compose up -d

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
