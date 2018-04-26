package utils

import "strings"

//GetGrpcEndpoints return endpoint without 'http://' in url
//because gRPC protocol
func GetGrpcEndpoints(origin []string) []string {
	r := make([]string, len(origin))
	for i, link := range origin {
		r[i] = strings.TrimPrefix(link, "http://")
	}
	return r
}
