SERVER_CN="sehyoung"

# commond name : certificate's unique name (usually use domain name)

# Generate CA's private key in 2048bit without password
# [help] This key is used to issue certificate.
openssl genrsa -out ca.key 2048

# Create self-signed certificate.
# [help] This certificate is used by client.
#        x509 : PKI(public key infrastructure)'s standard by ITU-T
#        -> issue certificate that has each distinct public key
#        subj : certificate's user
openssl req -new -x509 -days 365 -key ca.key -out ca.crt -subj "/CN=${SERVER_CN}"

# create server private key
# [help] DES, AES : encryption algorithm
#       -> The AES was introduced to overcome the des. (selected creterion: safety, cost, efficiency)
#       aes256 : 256bit aes
openssl genrsa -aes256 -out server.key 2048

# create certificate-sigend-request.
# [help] request certificate using this to CA
openssl req -new -key server.key -out server.csr -subj "/CN=${SERVER_CN}"

# The CA takes the csr and create the server's certificate using own private key.
# [help] set_serial : Each certificate is uniquely identified by a serial number.
openssl x509 -req -days 365 -in server.csr -CA ca.crt -CAkey ca.key -set_serial 01 -out server.crt

# Convert the server's private key format to the format used by grpc
openssl pkcs8 -topk8 -nocrypt -in server.key -out server.pem

# ******
# server.key : 6572