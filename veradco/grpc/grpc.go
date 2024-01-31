package grpc

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	// "flag"
	"fmt"
	"log"
	"net"
	"os"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"

	// "google.golang.org/grpc/credentials/alts"

	// pb "google.golang.org/grpc/examples/helloworld/helloworld"
	pb "github.com/smart-duck/veradco/veradco/protoc"

	"github.com/smart-duck/veradco/veradco/kres"
	"github.com/smart-duck/veradco/veradco/plugin"
	"github.com/smart-duck/veradco/veradco/admissioncontroller"

	// "gopkg.in/yaml.v2"
	// v1 "k8s.io/api/core/v1"

	// "k8s.io/apimachinery/pkg/runtime/serializer/json"
	"encoding/json"

	admission "k8s.io/api/admission/v1"
	// meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
)

/*
var (
	port = flag.Int("port", 50051, "The server port")
	// security = flag.String("security", "none", "mTLS if set to false") // tls or mtls or something else
	serverCertFile   = flag.String("server-cert", "none", "server cert in case TLS is used")                    // cert/server-cert.pem
	serverKeyFile    = flag.String("server-key", "none", "server key in case TLS is used")                      // cert/server-key.pem
	clientCaCertFile = flag.String("client-ca-cert", "none", "client cert in case mTLS is used on client side") // cert/ca-client-cert.pem
)
*/

type GrpcPlugin struct {
	pb.UnimplementedPluginServer
	initialized bool
	Port int
	ServerCertFile *string
	ServerKeyFile *string
	ClientCaCertFile *string
	VeradcoPlugin plugin.VeradcoPlugin
}

func NewGrpcPlugin() GrpcPlugin {
	return GrpcPlugin{Port: 50051}
}

func (gp *GrpcPlugin) StartServer() error {

	if gp.VeradcoPlugin == nil {
		return fmt.Errorf("No plugin passed")
	}

	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", gp.Port))
	if err != nil {
		return err
	}

	var s *grpc.Server

	transportCreds, err := gp.loadTLSCredentials()
	if err != nil {
		return err
	}

	if transportCreds == nil {
		s = grpc.NewServer()
	} else {
		s = grpc.NewServer(
			grpc.Creds(transportCreds),
		)
	}

	pb.RegisterPluginServer(s, gp)
	log.Printf("server listening at %v", lis.Addr())
	if err := s.Serve(lis); err != nil {
		return err
	}
	return nil
}

func (gp *GrpcPlugin) Discover(ctx context.Context, in *pb.Empty) (*pb.ConfigurationResponse, error) {
	conf, err := gp.VeradcoPlugin.Discover()
	if err != nil {
		return nil, err
	}
	return &pb.ConfigurationResponse{Configuration: conf}, nil
}

func (gp *GrpcPlugin) Execute(ctx context.Context, in *pb.AdmissionReview) (*pb.AdmissionResponse, error) {
	// log.Printf("Received: %v", in.GetReview())

	// First call
	if ! gp.initialized {
		err := gp.VeradcoPlugin.Init(in.GetConfiguration())
		if err != nil {
			return nil, err
		}
		gp.initialized = true
	}

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
		var result *admissioncontroller.Result
		// Should be a *meta.PartialObjectMetadata
		kobj, err := kres.ParseOther(review.Request)
		if err != nil {
			result = &admissioncontroller.Result{Msg: err.Error()}
		} else {

			// Execute(kobj runtime.Object, operation string, dryRun bool, r *admission.AdmissionRequest) (*admissioncontroller.Result, error)
			result, err = gp.VeradcoPlugin.Execute(kobj, string(review.Request.Operation), in.GetDryRun(), review.Request)
		}

		if result != nil {
			b, err := json.Marshal(result)
			if err != nil {
				respErr = fmt.Sprintf("%v", err)
			} else {
				response = b
			}
		} else {
			respErr = fmt.Sprintf("%v", err)
		}

		// Apply the plugins
		// return veradcoCfg.ProceedPlugins(body, other, r, scope, endpoint)

	} else {
		respErr = "request is undefined"
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

func (gp *GrpcPlugin) loadTLSCredentials() (credentials.TransportCredentials, error) {
	// No security case
	if gp.ServerCertFile == nil || gp.ServerKeyFile == nil {
		return nil, nil
	}

	// TLS case
	serverCert, err := tls.LoadX509KeyPair(*gp.ServerCertFile, *gp.ServerKeyFile)
	if err != nil {
		return nil, err
	}
	if gp.ClientCaCertFile == nil {
		config := &tls.Config{
			Certificates: []tls.Certificate{serverCert},
		}

		return credentials.NewTLS(config), nil
	}

	// mTLS case
	pemClientCA, err := os.ReadFile(*gp.ClientCaCertFile)
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