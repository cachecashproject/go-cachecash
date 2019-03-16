#!/bin/bash
set -xe

for x in upstream{,-apache,-lighttpd,-caddy,-python}; do
	make clean

	echo "[*] Generate config for $x..."
	go run ./cmd/generate-config -outputPath cfg/ -upstream "http://$x:80" -publisherCacheServiceAddr publisher:8082

	echo "[*] Starting network..."
	docker-compose up -d

	echo "[*] Waiting for network to become healthy..."
	while ! psql 'host=127.0.0.1 port=5432 user=postgres dbname=publisher sslmode=disable' -c 'select 1;'; do sleep 10; done

	echo "[*] Fetching from $x..."
	rm -f output.bin
	go run ./cmd/cachecash-curl -o output.bin cachecash://localhost:8080/file1.bin
	diff -q output.bin testdata/content/file1.bin
	echo "[+] Success"

	docker-compose down
done

echo "[+] All tests finished successfully"
