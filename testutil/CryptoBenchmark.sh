cd ~/go/src/github.com/cachecashproject/go-cachecash


echo "Starting Local Benchmark..."
echo "crc32..."
docker-compose down

echo "Building Phase..."
docker-compose build

echo "Setting Containers Up..."
docker-compose up -d

echo "Building Cachecash-curl..."
make ccc

echo "Start Testing..."

fileSize=$(stat -c %s ./testdata/content/file0.bin)
echo "Test data size is $fileSize"
cat > ./testutil/crc32.txt
echo "Result Generated is in ./testutil/crc32.txt"
for i in {1...100}
do
    echo -n "Trial $i ... "
    START=$(date +&s.%N)
    CACHECASH_INSECURE=true ./bin/cachecash-curl -o output.bin -logLevel=debug -trace http://localhost:14268 cachecash://localhost:7070/file0.bin
    END=$(date +&s.%N)
    DIFF=(END - START)
    echo -e "$DIFF \n" >> crc32.txt
    echo "\n"
done

echo "CRC32 benchmark is done"

echo "Shutting down containers..." 
docker-compose down

echo "md5..."
sed -i 's/dCrypto = "crc32"/dCrypto = "md5"/g' ./colocationpuzzle/puzzle.go

echo "Building Phase..."
docker-compose build

echo "Setting Containers Up..."
docker-compose up -d

echo "Building Cachecash-curl..."
make ccc

echo "Start Testing..."

fileSize=$(stat -c %s ./testdata/content/file0.bin)
echo "Test data size is $fileSize"
cat > ./testutil/md5.txt
echo "Result Generated is in ./testutil/md5.txt"
for i in {1...100}
do
    echo -n "Trial $i ... "
    START=$(date +&s.%N)
    CACHECASH_INSECURE=true ./bin/cachecash-curl -o output.bin -logLevel=debug -trace http://localhost:14268 cachecash://localhost:7070/file0.bin
    END=$(date +&s.%N)
    DIFF=$(echo "$END - $START" | bc)
    echo -e "$DIFF \n" >> md5.txt
    echo "\n"
done

echo "MD5 benchmark is done"

echo "Shutting down containers..." 
docker-compose down

echo "SHA384..."
sed -i 's/dCrypto = "md5"/dCrypto = "sha384"/g' ./colocationpuzzle/puzzle.go

echo "Building Phase..."
docker-compose build

echo "Setting Containers Up..."
docker-compose up -d

echo "Building Cachecash-curl..."
make ccc

echo "Start Testing..."

fileSize=$(stat -c %s ./testdata/content/file0.bin)
echo "Test data size is $fileSize"
cat > ./testutil/sha384.txt
echo "Result Generated is in ./testutil/md5.txt"
for i in {1...100}
do
    echo -n "Trial $i ... "
    START=$(date +&s.%N)
    CACHECASH_INSECURE=true ./bin/cachecash-curl -o output.bin -logLevel=debug -trace http://localhost:14268 cachecash://localhost:7070/file0.bin
    END=$(date +&s.%N)
    DIFF=$(echo "$END - $START" | bc)
    echo -e "$DIFF \n" >> sha384.txt
    echo "\n"
done

echo "ALL BENCHMARK FINISHED"
echo "Shutting down containers..."

docker-compose down

