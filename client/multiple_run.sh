
# chat-server : multiple client test
for var in in {1..10}
do
    go run client.go ${var}
done