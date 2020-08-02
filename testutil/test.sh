cd ~/go/src/github.com/cachecashproject/go-cachecash


echo -e "\n\n\n Starting Local Benchmark..."
echo -e "\n\n\n crc32..."
docker-compose down

echo -e "\n\n\n Building Phase..."
docker-compose build

echo -e "\n\n\n Setting Containers Up..."
docker-compose up -d

echo -e "\n\n\n Building Cachecash-curl..."
make ccc

echo -e "\n\n\n Start Testing..."

fileSize=$(stat -c %s ./testdata/content/file0.bin)
echo -e "\n\n\n Test data size is $fileSize"
touch ./testutil/crc32.txt
echo -e "\n\n\n Result Generated is in ./testutil/crc32.txt"
for i in {1..100}
do
    echo -n "Trial $i ... "
    START=$(date +%s.%N)
    CACHECASH_INSECURE=true ./bin/cachecash-curl -o output.bin -logLevel=debug -trace http://localhost:14268 cachecash://localhost:7070/file0.bin
    END=$(date +%s.%N)
    DIFF=$(echo "$END - $START" | bc)
    echo -e "$DIFF" >> ./testutil/crc32.txt
done

echo -e "\n\n\n CRC32 benchmark is done"

echo -e "\n\n\n Shutting down containers..." 
docker-compose down

echo -e "\n\n\n md5..."
sed -i 's/dCrypto = "crc32"/dCrypto = "md5"/g' ./colocationpuzzle/puzzle.go

echo -e "\n\n\n Building Phase..."
docker-compose build

echo -e "\n\n\n Setting Containers Up..."
docker-compose up -d

echo -e "\n\n\n Building Cachecash-curl..."
make ccc

echo -e "\n\n\n Start Testing..."

fileSize=$(stat -c %s ./testdata/content/file0.bin)
echo -e "\n\n\n Test data size is $fileSize"
touch ./testutil/md5.txt
echo -e "\n\n\n Result Generated is in ./testutil/md5.txt"
for i in {1..100}
do
    echo -n "Trial $i ... "
    START=$(date +%s.%N)
    CACHECASH_INSECURE=true ./bin/cachecash-curl -o output.bin -logLevel=debug -trace http://localhost:14268 cachecash://localhost:7070/file0.bin
    END=$(date +%s.%N)
    DIFF=$(echo "$END - $START" | bc)
    echo -e "$DIFF" >> ./testutil/md5.txt
done

echo -e "\n\n\n MD5 benchmark is done"

echo -e "\n\n\n Shutting down containers..." 
docker-compose down

echo "SHA384..."
sed -i 's/dCrypto = "md5"/dCrypto = "sha384"/g' ./colocationpuzzle/puzzle.go

echo -e "\n\n\n Building Phase..."
docker-compose build

echo -e "\n\n\n Setting Containers Up..."
docker-compose up -d

echo -e "\n\n\n Building Cachecash-curl..."
make ccc

echo -e "\n\n\n Start Testing..."

fileSize=$(stat -c %s ./testdata/content/file0.bin)
echo "Test data size is $fileSize"
touch ./testutil/sha384.txt
echo "Result Generated is in ./testutil/sha384.txt"
for i in {1..100}
do
    echo -n "Trial $i ... "
    START=$(date +%s.%N)
    CACHECASH_INSECURE=true ./bin/cachecash-curl -o output.bin -logLevel=debug -trace http://localhost:14268 cachecash://localhost:7070/file0.bin
    END=$(date +%s.%N)
    DIFF=$(echo "$END - $START" | bc)
    echo -e "$DIFF" >> ./testutil/sha384.txt
done

echo -e "\n\n\n\n ALL BENCHMARK FINISHED"
echo "\n\n\n\n Shutting down containers..."

docker-compose down

