package conf

import (
	"os"
	"crypto/tls"
	"crypto/x509"
	"context"
	"io/ioutil"
	log "k8s.io/klog/v2"
	"gopkg.in/yaml.v3"
	"encoding/json"
	veradcoplugin "github.com/smart-duck/veradco/plugin"
	goplugin "plugin"
	"fmt"
	"regexp"
	admission "k8s.io/api/admission/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"github.com/smart-duck/veradco/admissioncontroller"
	"github.com/smart-duck/veradco/kres"
	"github.com/smart-duck/veradco/monitoring"
	"time"
	"strings"
  // "math/rand"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
	pb "github.com/smart-duck/veradco/protoc"
)

const (
	timeout = time.Millisecond * 500
)

type Plugin struct {
	Name string `yaml:"name"`
	Path string `yaml:"path"` // can be a grpc address: localhost:50051
	Code string `yaml:"code,omitempty"`
	Kinds string `yaml:"kinds"`
	Operations string `yaml:"operations"`
	Namespaces string `yaml:"namespaces"`
	Labels []map[string]string `yaml:"labels,omitempty"`
	Annotations []map[string]string `yaml:"annotations,omitempty"`
	DryRun bool `yaml:"dryRun"`
	Configuration string `yaml:"configuration"`
	Scope string `yaml:"scope"`
	Endpoints string `yaml:"endpoints,omitempty"`
	GrpcClientCertFile string `yaml:"grpcClientCertFile,omitempty"` // cert/client-cert.pem
	GrpcClientKeyFile string `yaml:"grpcClientKeyFile,omitempty"` // cert/client-key.pem
	GrpcServerCaCertFile string `yaml:"grpcServerCaCertFile,omitempty"` // cert/ca-cert.pem
	GrpcAutoAccept bool `yaml:"grpcAutoAccept,omitempty"`
	GrpcUnallowOnFailure bool `yaml:"grpcUnallowOnFailure,omitempty"`
	GrpcConn *grpc.ClientConn `yaml:"-"`
	VeradcoPlugin veradcoplugin.VeradcoPlugin `yaml:"-"`
	VeradcoPluginLoaded bool `yaml:"-"`
}

type VeradcoCfg struct {
	FailOnPluginLoadingFails bool `yaml:"failOnPluginLoadingFails"`
	// RejectOnPluginError bool `yaml:"rejectOnPluginError"`Managed at k8s level failurePolicy
	Plugins []*Plugin `yaml:"plugins"`
}

func (veradcoCfg *VeradcoCfg) ReadConf(cfgFile string) error {

	yfile, err := ioutil.ReadFile(cfgFile)

	if err != nil {

		log.Errorf("Failed to read configuration %s: %v", cfgFile, err)

		return err
	}

	err = yaml.Unmarshal(yfile, &veradcoCfg)

	if err != nil {
		log.Errorf("Failed to load configuration %s: %v", cfgFile, err)

		return err
	}

	return nil
}

func (plugin *Plugin) queryGrpcPlugin(review []byte) (*pb.AdmissionResponse, error) {
	c := pb.NewPluginClient(plugin.GrpcConn)

	// Contact the server and print out its response.
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	var result *pb.AdmissionResponse
	var err error

	result, err = c.Execute(ctx, &pb.AdmissionReview{Review: review})

	return result, err
}

func (plugin *Plugin) loadTLSCredentials() (credentials.TransportCredentials, error) {
	// No security case
	if plugin.GrpcServerCaCertFile == "" {
		return nil, nil
	}
	pemServerCA, err := os.ReadFile(plugin.GrpcServerCaCertFile)
	if err != nil {
		return nil, err
	}

	certPool := x509.NewCertPool()
	if !certPool.AppendCertsFromPEM(pemServerCA) {
		return nil, fmt.Errorf("failed to add server CA's certificate")
	}

	// TLS case
	if plugin.GrpcClientCertFile == "" || plugin.GrpcClientKeyFile == "" {
		config := &tls.Config{
			RootCAs: certPool,
		}
		return credentials.NewTLS(config), nil
	}

	// mTLS case
	clientCert, err := tls.LoadX509KeyPair(plugin.GrpcClientCertFile, plugin.GrpcClientKeyFile)
	if err != nil {
		return nil, err
	}

	config := &tls.Config{
		Certificates: []tls.Certificate{clientCert},
		RootCAs:      certPool,
	}

	return credentials.NewTLS(config), nil
}

func (plugin *Plugin) newGrpcConn(addr string) (*grpc.ClientConn, error) {
	transportCreds, err := plugin.loadTLSCredentials()
	if err != nil {
		return nil, err
	}

	if transportCreds == nil {
		return grpc.Dial(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	} else {
		return grpc.Dial(addr, grpc.WithTransportCredentials(transportCreds))
	}
}

func (veradcoCfg *VeradcoCfg) LoadPlugins() (int, error) {
	numberOfPluginsLoaded := 0
	log.Infof(">>>> Loading plugins")

	reGrpc := regexp.MustCompile(`^[^:]+:[0-9]+$`)

	for _, plugin := range veradcoCfg.Plugins {
		log.Infof(">> Loading plugin  %s", plugin.Name)

		// Test if it is a GRPC plugin or a legacy one
		if reGrpc.MatchString(plugin.Path) {
			// GRPC plugin
			conn, err := plugin.newGrpcConn(plugin.Path)

			if err != nil {
				log.Errorf("Unable to load GRPC plugin %s: %v", plugin.Name, err)
				if ! plugin.DryRun && veradcoCfg.FailOnPluginLoadingFails {
					return numberOfPluginsLoaded, err			
				}
			} else {
				// defer conn.Close(): conn never closed TODO?
				plugin.GrpcConn = conn
				plugin.VeradcoPluginLoaded = true
				numberOfPluginsLoaded++
			}
		} else {
			// Legacy plugin: go plugin

			// Try to execute plugins
			plug, err := goplugin.Open(plugin.Path)
			if err != nil {
				log.Errorf("Unable to load plugin %s: %v", plugin.Name, err)
				if ! plugin.DryRun && veradcoCfg.FailOnPluginLoadingFails {
					return numberOfPluginsLoaded, err			
				}
			} else {
				pluginHandler, err := plug.Lookup("VeradcoPlugin")
				if err != nil {
					log.Errorf("Unable to find handler for plugin %s: %v", plugin.Name, err)
					if ! plugin.DryRun && veradcoCfg.FailOnPluginLoadingFails {
						return numberOfPluginsLoaded, err			
					}
				} else {
					var veradcoPlugin veradcoplugin.VeradcoPlugin

					veradcoPlugin, ok := pluginHandler.(veradcoplugin.VeradcoPlugin)
					if !ok {
						log.Errorf("Plugin %s does not implement awaited interface", plugin.Name)
						if ! plugin.DryRun && veradcoCfg.FailOnPluginLoadingFails {
							return numberOfPluginsLoaded, fmt.Errorf("Plugin %s does not implement awaited interface\n", plugin.Name)			
						}
					} else {

						log.Infof(">> Init plugin %s", plugin.Name)
						err := veradcoPlugin.Init(plugin.Configuration)

						if err != nil {
							log.Errorf("Unable to init plugin %s (skipped): %v", plugin.Name, err)
							if ! plugin.DryRun && veradcoCfg.FailOnPluginLoadingFails {
								return numberOfPluginsLoaded, err			
							}
						} else {
							plugin.VeradcoPlugin = veradcoPlugin
							plugin.VeradcoPluginLoaded = true
							numberOfPluginsLoaded++
						}
						// log.Infof("Plugin: %v\n", plugin)
					}
				}
			}
		}
	}

	if numberOfPluginsLoaded > 0 {
		log.Infof(">> %d plugins loaded over %d", numberOfPluginsLoaded, len(veradcoCfg.Plugins))
		return numberOfPluginsLoaded, nil
	}
	return numberOfPluginsLoaded, fmt.Errorf("No plugin loaded")
}

// func (veradcoCfg *VeradcoCfg) GetPlugins(r *admission.AdmissionRequest, kind string, operation string, namespace string, labels map[string]string, annotations map[string]string, scope string) (*[]*Plugin, error) {
func (veradcoCfg *VeradcoCfg) GetPlugins(r *admission.AdmissionRequest, scope string, endpoint string) (*[]*Plugin, error) {
	// log.Infof("Plugins: %v\n", veradcoCfg.Plugins)

	log.V(2).Infof(">>>> GetPlugins called")

	result := []*Plugin{}

	// Browse all plugins to filter the relevant ones
	for _, plugin := range veradcoCfg.Plugins {

		if ! plugin.VeradcoPluginLoaded {
			// Plugin has not been loaded properly
			continue
		}

		// Check endpoint
		if len(plugin.Endpoints) > 0 {
			match, err := matchRegex(plugin.Endpoints, endpoint)
			if err != nil {
				log.Errorf("Failed to evaluate regex %s for %s: %s", plugin.Endpoints, r.Name, err)
				continue
			} else {
				if ! *match {
					continue
				}
			}
		}

		// check scope
		match, err := matchRegex(plugin.Scope, scope)
		if err != nil {
			log.Errorf("Failed to evaluate regex %s for %s: %s", plugin.Scope, r.Name, err)
			continue
		} else {
			if ! *match {
				continue
			}
		}

		// Check Resource kind
		match, err = matchRegex(plugin.Kinds, r.Kind.Kind)
		if err != nil {
			log.Errorf("Failed to evaluate regex %s for %s: %s", plugin.Kinds, r.Name, err)
			continue
		} else {
			if ! *match {
				continue
			}
		}

		// Check Operation
		match, err = matchRegex(plugin.Operations, string(r.Operation))
		if err != nil {
			log.Errorf("Failed to evaluate regex %s for %s: %s", plugin.Operations, r.Name, err)
			continue
		} else {
			if ! *match {
				continue
			}
		}

		// Check Namespace
		match, err = matchRegex(plugin.Namespaces, string(r.Namespace))
		if err != nil {
			log.Errorf("Failed to evaluate regex %s for %s: %s", plugin.Namespaces, r.Name, err)
			continue
		} else {
			if ! *match {
				continue
			}
		}

		// Check labels
		if len(plugin.Labels) > 0 {

			var check = false

			log.V(3).Infof("Check labels to filter plugin %s", plugin.Name)
			
			obj, err := kres.ParseOther(r)

			if err == nil {
				for _, label := range plugin.Labels {
					log.V(3).Infof("Inspect label %s for plugin %s", label["key"], plugin.Name)
					tmp, exists := obj.ObjectMeta.Labels[label["key"]]
					if ! exists {
						log.V(3).Infof("Inspect label %s for plugin %s: does NOT exist", label["key"], plugin.Name)
						check = true
						break
					} else {
						matched, err := matchRegex(label["value"], tmp)
						if err != nil {
							log.Errorf("Failed to evaluate label regex %s for %s: %s", label["value"], r.Name, err)
							check = true
							break
						}
						if ! *matched {
							log.V(3).Infof("Inspect label %s for plugin %s: does NOT match / value: %v / regex: %s", label["key"], plugin.Name, tmp, label["value"])
							check = true
							break
						}
					}
				}
			}
			if check {
				// skip plugin
				log.V(3).Infof("skip plugin %s: does NOT match labels", plugin.Name)
				continue
			}
		}

		// Check annotations
		if len(plugin.Annotations) > 0 {

			var check = false

			log.V(3).Infof("Check annotations to filter plugin %s", plugin.Name)

			obj, err := kres.ParseOther(r)

			if err == nil {
				for _, annotation := range plugin.Annotations {
					tmp, exists := obj.ObjectMeta.Annotations[annotation["key"]]
					if ! exists {
						log.V(3).Infof("Inspect annotation %s for plugin %s: does NOT exist", annotation["key"], plugin.Name)
						check = true
						break
					} else {
						matched, err := matchRegex(annotation["value"], tmp)
						if err != nil {
							log.Errorf("Failed to evaluate annotation regex %s for %s: %s", annotation["value"], r.Name, err)
							check = true
							break
						}
						if ! *matched {
							log.V(3).Infof("Inspect annotation %s for plugin %s: does NOT match / %v", annotation["key"], plugin.Name, tmp)
							check = true
							break
						}
					}
				}
			}
			if check {
				// skip plugin
				log.V(3).Infof("skip plugin %s: does NOT match labels", plugin.Name)
				continue
			}
		}

		// Add the plugin
		result = append(result, plugin)
	}

	log.Infof(">> Number of plugins selected: %d", len(result))

	return &result, nil
}

func (veradcoCfg *VeradcoCfg) ProceedPlugins(body *[]byte, kobj runtime.Object, r *admission.AdmissionRequest, scope string, endpoint string) (*admissioncontroller.Result, error) {

	plugins, err := veradcoCfg.GetPlugins(r, scope, endpoint)

	if err != nil {
		log.Errorf("Failed to load plugins: %v", err)
		return &admissioncontroller.Result{Allowed: true}, nil
	}

	globalResult := admissioncontroller.Result{
		Allowed:  true,
		Msg: "",
		PatchOps: make([]admissioncontroller.PatchOperation, 0),
	}

	for _, plug := range *plugins {
		log.V(1).Infof(">> Execute plugin %s\n", plug.Name)
		// Execute(meta meta.TypeMeta, kobj interface{}, r *admission.AdmissionRequest) (*admissioncontroller.Result, error)
		// veradcoPlugin.Execute(meta.TypeMeta{}, pod, r)
		startTime := time.Now()
		var result *admissioncontroller.Result
		var err error
		// Grpc plugin
		if plug.GrpcConn != nil {
			if plug.GrpcAutoAccept {
				go func() {
					_, err := plug.queryGrpcPlugin(*body)
					if err != nil {
						log.Errorf("Error while executing async GRPC plugin %s: %v", plug.Name, err)
					}
				}()
				result = &admissioncontroller.Result{Allowed: true}
			} else {
				var admResponse *pb.AdmissionResponse
				var errGrpc error
				admResponse, errGrpc = plug.queryGrpcPlugin(*body)
				if errGrpc != nil {
					log.Errorf("Error while executing sync GRPC plugin %s: %v", plug.Name, errGrpc)
					errorMsg := fmt.Sprintf("Error while executing GRPC plugin %s: %v", plug.Name, errGrpc)
					result = &admissioncontroller.Result{Allowed: !plug.GrpcUnallowOnFailure, Msg: errorMsg}
					// log.V(4).Infof("plug.GrpcUnallowOnFailure %s: %t", plug.Name, plug.GrpcUnallowOnFailure)
				} else {
					if admResponse.GetError() != "" {
						log.V(4).Infof(">> admResponse.GetError() %s: %s", plug.Name, admResponse.GetError())
					}
					if len(admResponse.GetResponse()) > 0 {
						log.V(4).Infof(">> admResponse.GetResponse() %s: %c", plug.Name, admResponse.GetResponse())
						// Initialize to Allowed
						hookResult := admissioncontroller.Result{Allowed: true}
						// var hookResult admissioncontroller.Result
						// grpcResp := string(admResponse.GetResponse())
						// log.V(4).Infof("grpcResp %s: %s", plug.Name, grpcResp)

						// IF NEED TO CHHECK FORMAT...
						// decoder := json.NewDecoder(strings.NewReader(string(jsonBlob)))
						// decoder.DisallowUnknownFields()
						// err = decoder.Decode(&animals)
						// if err != nil {
						// 	fmt.Println("error:", err)
						// }
						// fmt.Printf("%+v\n", animals)

						// errUnmarshal := yaml.Unmarshal([]byte(grpcResp), &hookResult)
						errUnmarshal := json.Unmarshal(admResponse.GetResponse(), &hookResult)
						if errUnmarshal != nil {
							log.Errorf("Failed to unmarshal response %s (unexpected): %v", plug.Name, errUnmarshal)
							result = &admissioncontroller.Result{Allowed: true}
						} else {
							result = &hookResult
							log.V(4).Infof("result %s: %v", plug.Name, result)
						}
					} else {
						result = &admissioncontroller.Result{Allowed: true}
					}
				}
			}
		} else {
			result, err = plug.VeradcoPlugin.Execute(kobj, string(r.Operation), *r.DryRun || plug.DryRun, r)
		}

		log.V(4).Infof("result %s: %v", plug.Name, result)

		// rand.Seed(time.Now().UnixNano())
    // n := rand.Intn(3000000000) // n will be between 0 and 3
    // time.Sleep(time.Duration(n)*time.Nanosecond)
		elapsed := time.Since(startTime)
		if err == nil {
			monitoring.AddOperation(plug.Name, plug.Scope, plug.DryRun, result.Allowed, r.Kind.Group, r.Kind.Version, r.Kind.Kind, r.Name, r.Namespace, string(r.Operation), "false")
			monitoring.AddStat(plug.Name, plug.Scope, plug.DryRun, result.Allowed, r.Kind.Group, r.Kind.Version, r.Kind.Kind, r.Name, r.Namespace, string(r.Operation), "false", elapsed)
			// monitoring.AddOperation(plug, r, result)
			if plug.VeradcoPlugin != nil {
				log.Infof(">> Plugin %s execution summary: %s", plug.Name, plug.VeradcoPlugin.Summary())
			}
			if plug.DryRun {
				log.Infof(">> Plugin %s is in dry run mode. Nothing to do!", plug.Name)
			} else {
				if ! result.Allowed {
					return result, nil
				} else {
					globalResult.Msg += result.Msg + "\n"
					globalResult.PatchOps = append(globalResult.PatchOps, result.PatchOps...)
				}
			}
		} else {
			monitoring.AddOperation(plug.Name, plug.Scope, plug.DryRun, false, r.Kind.Group, r.Kind.Version, r.Kind.Kind, r.Name, r.Namespace, string(r.Operation), "true")
			monitoring.AddStat(plug.Name, plug.Scope, plug.DryRun, false, r.Kind.Group, r.Kind.Version, r.Kind.Kind, r.Name, r.Namespace, string(r.Operation), "true", elapsed)
			return result, err
		}
	}
	
	globalResult.Msg = strings.ReplaceAll(strings.Trim(globalResult.Msg, "\n"), "\n", " - ")

	// return &admissioncontroller.Result{Allowed: true}, nil
	return &globalResult, nil

}

func matchRegex(regex string, value string) (*bool, error) {
	// regular expression act as a reverse pattern if it is prefixed by (!~)
	// By example, "(!~)(?i)test" matches that the value does not contain "test" whatever the case is.
	// log.Infof("Evaluate regex %s on %s\n", regex, value)
	
	appliedRegex := regex
	
	matched, err := regexp.MatchString(`^\(!~\).+`, appliedRegex)
	if err != nil {
		log.Errorf("Evaluate regex %s on %s failed: %v", regex, value, err) 
		return nil, err
	}

	inverted := false

	if matched {
		appliedRegex = appliedRegex[4:]
		inverted = true
	}

	matched, err = regexp.MatchString(appliedRegex, value)
	if err != nil {
		log.Errorf("Evaluate regex %s from %s on %s failed: %v", appliedRegex, regex, value, err)
		return nil, err
	}
	
	if inverted {
		matched = ! matched
	}
	
	log.V(3).Infof(">> Evaluate regex %s on %s: %t\n", regex, value, matched)

	return &matched, nil

}

// func main() {
  
// 	fmt.Println("Starting dummy veradco!")

// 	// Load conf from yaml
// 	conf, err := ReadConf("veradco.yaml")
// 	if err != nil {
// 		fmt.Printf("Error: %v\n", err)
// 	} else {
// 		fmt.Printf("Conf: %v\n", conf)
// 	}
// }