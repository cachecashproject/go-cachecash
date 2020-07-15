cd ~/go/src/github.com/cachecashproject/go-cachecash


echo "Starting Local Benchmark..."
echo "crc32..."
echo "Building Phase..."
docker-compose build

echo "Setting Containers Up..."
docker-compose up -d

echo "Building Cachecash-curl..."
make ccc

echo "Start Testing..."

fileSize=stat -c %s ./testdata/file0.bin
echo "Test data size is $fileSize"

for i in 'seq 1 2'
do
    echo -n "Trial $i ... "
    cat > crc32.txt
    echo -e "Trial $i: \n\
        'time ./bin/cachecash-curl -o output.bin -logLevel=debug -trace http://localhost:14268 cachecash://localhost:7070/file0.bin'" >> crc32.txt
    echo "\n"
done

    

