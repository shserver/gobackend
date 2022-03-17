sudo security add-trusted-cert -d -r trustRoot -k /Library/Keychains/System.keychain ca.crt
sudo security delete-certificate -c "ca.crt"