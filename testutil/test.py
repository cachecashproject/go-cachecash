
f = open("sha384.txt", "r")
data = []
for i in range(100):
    data.append(float(f.readline()))

average = 0
for j in range(len(data)):
    average += data[j]
average = average / 100
print("sha384", average)