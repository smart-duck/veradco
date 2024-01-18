/*
 *
 * Copyright 2015 gRPC authors.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 */

// Package main implements a server for Greeter service.
package main

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"flag"
	"fmt"
	"log"
	"net"
	"os"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"

	// "google.golang.org/grpc/credentials/alts"

	// pb "google.golang.org/grpc/examples/helloworld/helloworld"
	pb "github.com/smart-duck/veradco/protoc"

	"github.com/smart-duck/veradco/kres"
	"github.com/smart-duck/veradco/admissioncontroller"

	// "gopkg.in/yaml.v2"
	// v1 "k8s.io/api/core/v1"

	// "k8s.io/apimachinery/pkg/runtime/serializer/json"
	"encoding/json"

	admission "k8s.io/api/admission/v1"
	// meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
)

var (
	port = flag.Int("port", 50051, "The server port")
	// security = flag.String("security", "none", "mTLS if set to false") // tls or mtls or something else
	serverCertFile   = flag.String("server-cert", "none", "server cert in case TLS is used")                    // cert/server-cert.pem
	serverKeyFile    = flag.String("server-key", "none", "server key in case TLS is used")                      // cert/server-key.pem
	clientCaCertFile = flag.String("client-ca-cert", "none", "client cert in case mTLS is used on client side") // cert/ca-client-cert.pem
)

// Usage:
// mTLS: go run plugin/main.go -server-cert=cert/server-cert.pem -server-key=cert/server-key.pem -client-ca-cert=cert/ca-client-cert.pem
// TLS: go run plugin/main.go -server-cert=cert/server-cert.pem -server-key=cert/server-key.pem
// none: go run plugin/main.go

func loadTLSCredentials() (credentials.TransportCredentials, error) {
	// No security case
	if *serverCertFile == "none" || *serverKeyFile == "none" {
		return nil, nil
	}

	// TLS case
	serverCert, err := tls.LoadX509KeyPair(*serverCertFile, *serverKeyFile)
	if err != nil {
		return nil, err
	}
	if *clientCaCertFile == "none" {
		config := &tls.Config{
			Certificates: []tls.Certificate{serverCert},
		}

		return credentials.NewTLS(config), nil
	}

	// mTLS case
	pemClientCA, err := os.ReadFile(*clientCaCertFile)
	if err != nil {
		return nil, err
	}

	certPool := x509.NewCertPool()
	if !certPool.AppendCertsFromPEM(pemClientCA) {
		return nil, fmt.Errorf("failed to add client CA's certificate")
	}
	config := &tls.Config{
		Certificates: []tls.Certificate{serverCert},
		ClientAuth:   tls.RequireAndVerifyClientCert,
		ClientCAs:    certPool,
	}

	return credentials.NewTLS(config), nil
}

// server is used to implement helloworld.GreeterServer.
type server struct {
	pb.UnimplementedPluginServer
}

// SayHello implements helloworld.GreeterServer
func (s *server) Execute(ctx context.Context, in *pb.AdmissionReview) (*pb.AdmissionResponse, error) {
	// log.Printf("Received: %v", in.GetReview())

	response := []byte(nil)
	respErr := ""

	// Should be an admission review
	var review admission.AdmissionReview
	decoder := serializer.NewCodecFactory(runtime.NewScheme()).UniversalDeserializer()
	if _, _, err := decoder.Decode(in.GetReview(), nil, &review); err != nil {
		respErr = fmt.Sprintf("%v", err)
		return &pb.AdmissionResponse{Response: response, Error: respErr}, nil
	}

	if review.Request != nil {
		// strResponse := fmt.Sprintf("UID: %s, Kind: %s, Resource: %s, Name: %s, Namespace: %s, Operation: %s", string(review.Request.UID), review.Request.Kind.String(), review.Request.Resource.String(), review.Request.Name, review.Request.Namespace, string(review.Request.Operation))
		// response = []byte(strResponse)
		var result admissioncontroller.Result
		// Should be a *meta.PartialObjectMetadata
		kobj, err := kres.ParseOther(review.Request)
		if err != nil {
			result = admissioncontroller.Result{Msg: err.Error()}
		} else {

			if kobj != nil {
				result = admissioncontroller.Result{Allowed:  true, Msg: kobj.Kind}
			} else {
				result = admissioncontroller.Result{Allowed:  true}
			}
		}

		b, err := json.Marshal(result)
		if err != nil {
			respErr = fmt.Sprintf("%v", err)
		} else {
			response = b
		}

		// Apply the plugins
		// return veradcoCfg.ProceedPlugins(body, other, r, scope, endpoint)

	}
	

	// Should be a pod for test
	// var pod v1.Pod
	// err := yaml.Unmarshal([]byte(in.GetReview()), &pod)
	// if err != nil {
	// 	respErr = fmt.Sprintf("%v", err)
	// } else {
	// 	if pod.Annotations == nil {
	// 		pod.Annotations = make(map[string]string)
	// 	}
	// 	pod.Annotations["your-annotation-key"] = "your-annotation-value"
	// 	// serializer := json.NewYAMLSerializer(json.DefaultMetaFactory, nil, nil)
	// 	data, err := json.Marshal(pod)
	// 	// err := serializer.Encode(pod, data)
	// 	if err != nil {
	// 		panic(err)
	// 	}
	// 	// response = pod.String()
	// 	response = data
	// }

	return &pb.AdmissionResponse{Response: response, Error: respErr}, nil
}

func main() {
	flag.Parse()
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", *port))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	// altsTC := alts.NewServerCreds(alts.DefaultServerOptions())
	// log.Printf("altsTC: %v", altsTC)
	// s := grpc.NewServer(grpc.Creds(altsTC))

	var s *grpc.Server

	transportCreds, err := loadTLSCredentials()
	if err != nil {
		log.Fatal("cannot load TLS credentials: ", err)
	}

	if transportCreds == nil {
		s = grpc.NewServer()
	} else {
		s = grpc.NewServer(
			grpc.Creds(transportCreds),
		)
	}

	pb.RegisterPluginServer(s, &server{})
	log.Printf("server listening at %v", lis.Addr())
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
