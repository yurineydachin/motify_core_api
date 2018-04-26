package etcd

func SetCerts() {
	certFile = "certs/client.crt"
	keyFile = "certs/client.key.pem"
	caFile = "certs/ca.cert.pem"
}

func UnsetCerts() {
	certFile = ""
	keyFile = ""
	caFile = ""
}
