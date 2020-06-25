echo "Make sure https://github.com/smallstep/cli is installed on your system and the CA server is run using the step-ca utility: step-ca $(step path)/config/ca.json"

# Request server certificate
step ca certificate "127.0.0.1" server.crt server.key
step ca root ca.crt

# Request client certificate
step ca certificate "client1" client1.crt client1.key

# Prolong the individual certificate expiration date
step ca renew server.crt server.key --expires-in=720h --force