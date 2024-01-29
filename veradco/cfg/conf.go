package conf

import (
	"os"
	"crypto/tls"
	"crypto/x509"
	"context"
	log "k8s.io/klog/v2"
	"gopkg.in/yaml.v3"
	"encoding/json"
	veradcoplugin "github.com/smart-duck/veradco/plugin"
	goplugin "plugin"
	"fmt"
	"regexp"
	admission "k8s.io/api/admission/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/kubernetes"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/dynamic"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"github.com/smart-duck/veradco/admissioncontroller"
	"github.com/smart-duck/veradco/kres"
	"github.com/smart-duck/veradco/monitoring"
	"time"
	"strings"
	"sync"
  // "math/rand"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
	pb "github.com/smart-duck/veradco/protoc"
)

const (
	timeout = time.Millisecond * 500
)

var (
	reGrpc = regexp.MustCompile(`^[^:]+:[0-9]+$`)
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
	ActivateDiscovery bool `yaml:"activateDiscovery,omitempty"`
	ActivateCRD bool `yaml:"activateCRD,omitempty"`
	// RejectOnPluginError bool `yaml:"rejectOnPluginError"`Managed at k8s level failurePolicy
	Plugins []*Plugin `yaml:"plugins"`
	rwMutexPlugins sync.RWMutex `yaml:"-"`
	mutexCR sync.Mutex `yaml:"-"`
	// alreadyAddedCR string  `yaml:"-"`
	currentCR map[string]*Plugin `yaml:"-"`
}

func (veradcoCfg *VeradcoCfg) ReadConf(cfgFile string) error {

	yfile, err := os.ReadFile(cfgFile)

	if err != nil {

		log.Errorf("Failed to read configuration %s: %v", cfgFile, err)

		return err
	}

	veradcoCfg.rwMutexPlugins.Lock()

	defer veradcoCfg.rwMutexPlugins.Unlock()

	err = yaml.Unmarshal(yfile, &veradcoCfg)

	if err != nil {
		log.Errorf("Failed to load configuration %s: %v", cfgFile, err)

		return err
	}

	veradcoCfg.currentCR = make(map[string]*Plugin)

	go func() {
		// Wait for the CR to be applied
		time.Sleep(15 * time.Second)
		veradcoCfg.DiscoverGrpcPluginsCR(nil)
	}()

	if veradcoCfg.ActivateDiscovery {
		// Launch a go routine to discover services of the veradco-plugins namespace
		go veradcoCfg.discoverGrpcPlugins()
	}

	return nil
}

func (veradcoCfg *VeradcoCfg) loadPluginFromConf(conf []byte) (*Plugin, error) {

	var plugin *Plugin

	err := yaml.Unmarshal(conf, &plugin)

	if err != nil {
		log.Errorf("Failed to load plugin: %v", err)

		return nil, err
	}

	return plugin, nil
}

func (veradcoCfg *VeradcoCfg) addPlugin(plugin *Plugin) {
	veradcoCfg.rwMutexPlugins.Lock()

	defer veradcoCfg.rwMutexPlugins.Unlock()

	log.Infof("Add plugin %s", plugin.Name)

	veradcoCfg.Plugins = append(veradcoCfg.Plugins, plugin)
}

func (veradcoCfg *VeradcoCfg) removePlugin(plugin *Plugin) {
	veradcoCfg.rwMutexPlugins.Lock()

	defer veradcoCfg.rwMutexPlugins.Unlock()

	for index, currPlug := range veradcoCfg.Plugins {
		if currPlug == plugin {
			log.Infof("Delete plugin %s", plugin.Name)
			veradcoCfg.Plugins = append(veradcoCfg.Plugins[:index], veradcoCfg.Plugins[index+1:]...)
		}
	}
}

func (veradcoCfg *VeradcoCfg) DiscoverGrpcPluginsCR(r *admission.AdmissionRequest) {

	getInfoFromCR := func(cr unstructured.Unstructured) (string, string, error) {
		var name string

		metadata, ok := cr.Object["metadata"]
		// If the key exists
		if ok {
			// Do something
			meta, ok := metadata.(map[string]interface{})
			if ok {
				name, ok = meta["name"].(string)
				// fmt.Println(plugin["plugin"])
				if ok {
					log.V(4).Infof("[Discover GRPC Plugins CR] Discover CR %s", name)
				} else {
					return "", "", fmt.Errorf("metadata name not found in CR")
				}
			} else {
				return "", "", fmt.Errorf("metadata name not found in CR")
			}
		} else {
			return "", "", fmt.Errorf("metadata name not found in CR")
		}

		var confPlugin string

		spec, ok := cr.Object["spec"]
		// If the key exists
		if ok {
			// Do something
			pluginMap, ok := spec.(map[string]interface{})
			if ok {
				confPlugin, ok = pluginMap["plugin"].(string)
				// fmt.Println(plugin["plugin"])
				if !ok {
					return "", "", fmt.Errorf("plugin conf not found in CR")
				}
			} else {
				return "", "", fmt.Errorf("plugin conf not found in CR")
			}
		} else {
			return "", "", fmt.Errorf("plugin conf not found in CR")
		}

		return name, confPlugin, nil
		
	}

	if !veradcoCfg.ActivateCRD {
		return
	}

	config, err := rest.InClusterConfig()
	if err != nil {
		log.Errorf("[Discover GRPC Plugins CR] Failed to load kubeconfig: %v", err)
		return
	}
	

	// Create a dynamic client
	client, err := dynamic.NewForConfig(config)
	if err != nil {
		log.Errorf("[Discover GRPC Plugins CR] Failed to create dynamic client: %v", err)
		return
	}

	// Define the API group, version, and resource type of the VeradcoPlugin CR
	groupVersionResource := schema.GroupVersionResource{
		Group:    "smartduck.ovh",
		Version:  "v1",
		Resource: "veradcoplugins",
	}

	veradcoCfg.mutexCR.Lock()
	defer veradcoCfg.mutexCR.Unlock()

	operation := "CREATE"
	resName := ""
	
	if r != nil {
		operation = string(r.Operation)
		resName = r.Name
	}

	log.V(4).Infof("[Discover GRPC Plugins CR] Operation: %s", operation)

	if operation == "DELETE" {
		// Manage plugin deletion
		plugin, exists := veradcoCfg.currentCR[resName]
		if exists {
			veradcoCfg.removePlugin(plugin)
			delete(veradcoCfg.currentCR, resName)
		}
		return
	}

	// Retrieve the content of the custom resource
	result, err := client.Resource(groupVersionResource).Namespace("veradco-plugins").List(context.TODO(), meta.ListOptions{})
	if err != nil {
		log.Errorf("[Discover GRPC Plugins CR] Failed to create client set: %v", err)
		return
	} else {
		log.V(4).Infof("[Discover GRPC Plugins CR] Number of CR found %d", len(result.Items))
		for _, cr := range result.Items {

			name, conf, err := getInfoFromCR(cr)

			if err != nil {
				log.Errorf("[Discover GRPC Plugins CR] Unable to parse CR: %v", err)
				continue
			}

			currPlug, exists := veradcoCfg.currentCR[name]

			if !exists {

				plugin, err := veradcoCfg.loadPluginFromConf([]byte(conf))
				if err != nil {
					log.Errorf("[Discover GRPC Plugins CR] Unable to load configuration for service %s: %v", name, err)
					continue
				}

				_, err = veradcoCfg.loadPlugin(plugin)

				if err != nil {
					log.Errorf("[Discover GRPC Plugins CR] Unable to load plugin %s: %v", name, err)
				}

				veradcoCfg.addPlugin(plugin)


				veradcoCfg.currentCR[name] = plugin

			} else {
				if operation == "UPDATE" {
					// Manage plugin update (recreate)

					if resName == name {

						plugin, err := veradcoCfg.loadPluginFromConf([]byte(conf))
						if err != nil {
							log.Errorf("[Discover GRPC Plugins CR] Unable to load configuration for service %s (UPDATE): %v", name, err)
							continue
						}

						_, err = veradcoCfg.loadPlugin(plugin)

						if err != nil {
							log.Errorf("[Discover GRPC Plugins CR] Unable to load plugin %s (UPDATE): %v", name, err)
						}

						veradcoCfg.removePlugin(currPlug)

						veradcoCfg.addPlugin(plugin)

						veradcoCfg.currentCR[name] = plugin
					} else {
						log.V(4).Infof("[Discover GRPC Plugins CR] CR %s already encountered", name)
					}
				} else {
					log.V(4).Infof("[Discover GRPC Plugins CR] CR %s already encountered", name)
				}
			}
		}
	}
}

func (veradcoCfg *VeradcoCfg) discoverGrpcPlugins() {

	config, err := rest.InClusterConfig()
	if err != nil {
		log.Errorf("[Discover GRPC Plugins] Failed to load kubeconfig: %v", err)
		return
	}
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		log.Errorf("[Discover GRPC Plugins] Failed to create client set: %v", err)
		return
	}

	alreadyAddedNames := make(map[string]*Plugin)

	time.Sleep(10 * time.Second)

	for {
		// List services
		services, err := clientset.CoreV1().Services("veradco-plugins").List(context.TODO(), meta.ListOptions{})
		if err == nil {
			
			// First remove no more existing services
			for name, currPlug := range alreadyAddedNames {
				svcThere := false
				for _, service := range services.Items {
					if service.Name == name {
						svcThere = true
						continue
					}
				}
				if !svcThere {
					// Remove plugin
					veradcoCfg.removePlugin(currPlug)
					delete(alreadyAddedNames, name)
					log.Infof("[Discover GRPC Plugins] Service %s has been removed from plugins", name)
				}
			}

			// Add new plugins
			for _, service := range services.Items {
				_, exists := service.Labels["veradco.discover"]
				if exists {
					log.V(4).Infof("[Discover GRPC Plugins] Service %s has discovery label", service.Name)

					_, exists := alreadyAddedNames[service.Name]

					if !exists {
						log.Infof("[Discover GRPC Plugins] Retrieve configuration from service %s", service.Name)

						// addr := service.Spec.ExternalName + ":" + string(service.Spec.Ports[0].Port)

						addr := fmt.Sprintf("%s.%s:%d", service.Name, service.Namespace, service.Spec.Ports[0].Port)

						confResp, err := veradcoCfg.queryGrpcPluginConfiguration(addr)

						if err != nil {
							log.Errorf("[Discover GRPC Plugins] Unable to retrieve configuration for service %s (addr %s): %v", service.Name, addr, err)
							continue
						}

						rawConf := confResp.Configuration

						if len(rawConf) > 0 {
							plugin, err := veradcoCfg.loadPluginFromConf(rawConf)
							if err != nil {
								log.Errorf("[Discover GRPC Plugins] Unable to load configuration for service %s (addr %s): %v", service.Name, addr, err)
								continue
							}

							_, err = veradcoCfg.loadPlugin(plugin)

							if err != nil {
								log.Errorf("[Discover GRPC Plugins] Unable to load plugin %s (addr %s): %v", service.Name, addr, err)
								continue
							}

							plugin.Path = addr

							veradcoCfg.addPlugin(plugin)

							alreadyAddedNames[service.Name] = plugin
						} else {
							log.Errorf("[Discover GRPC Plugins] Configuration for service %s (addr %s) is empty", service.Name, addr)
							continue
						}
					} else {
						log.V(4).Infof("[Discover GRPC Plugins] Service %s already discovered", service.Name)
					}
					// fmt.Println(service.Name)
				}
			}
		}
		time.Sleep(1 * time.Minute)
	}
}

func (veradcoCfg *VeradcoCfg) queryGrpcPluginConfiguration(addr string) (*pb.ConfigurationResponse, error) {

	clientConn, err := grpc.Dial(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))

	if err != nil {
		return nil, err
	}

	c := pb.NewPluginClient(clientConn)

	// Contact the server and print out its response.
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	var result *pb.ConfigurationResponse

	result, err = c.Discover(ctx, &pb.Empty{})

	return result, err
}

func (plugin *Plugin) queryGrpcPlugin(review []byte) (*pb.AdmissionResponse, error) {
	c := pb.NewPluginClient(plugin.GrpcConn)

	// Contact the server and print out its response.
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	var result *pb.AdmissionResponse
	var err error

	result, err = c.Execute(ctx, &pb.AdmissionReview{Review: review, Configuration: plugin.Configuration, DryRun: plugin.DryRun})

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

func (veradcoCfg *VeradcoCfg) loadPlugin(plugin *Plugin) (int, error) {
	log.Infof(">> Loading plugin  %s", plugin.Name)

	// Test if it is a GRPC plugin or a legacy one
	if reGrpc.MatchString(plugin.Path) {
		// GRPC plugin
		conn, err := plugin.newGrpcConn(plugin.Path)

		if err != nil {
			log.Errorf("Unable to load GRPC plugin %s: %v", plugin.Name, err)
			if ! plugin.DryRun && veradcoCfg.FailOnPluginLoadingFails {
				return 0, err			
			} else {
				return 0, nil
			}
		} else {
			// defer conn.Close(): conn never closed TODO?
			plugin.GrpcConn = conn
			plugin.VeradcoPluginLoaded = true
			return 1, nil
		}
	} else {
		// Legacy plugin: go plugin

		// Try to execute plugins
		plug, err := goplugin.Open(plugin.Path)
		if err != nil {
			log.Errorf("Unable to load plugin %s: %v", plugin.Name, err)
			if ! plugin.DryRun && veradcoCfg.FailOnPluginLoadingFails {
				return 0, err			
			} else {
				return 0, nil
			}
		} else {
			pluginHandler, err := plug.Lookup("VeradcoPlugin")
			if err != nil {
				log.Errorf("Unable to find handler for plugin %s: %v", plugin.Name, err)
				if ! plugin.DryRun && veradcoCfg.FailOnPluginLoadingFails {
					return 0, err			
				} else {
					return 0, nil
				}
			} else {
				var veradcoPlugin veradcoplugin.VeradcoPlugin

				veradcoPlugin, ok := pluginHandler.(veradcoplugin.VeradcoPlugin)
				if !ok {
					log.Errorf("Plugin %s does not implement awaited interface", plugin.Name)
					if ! plugin.DryRun && veradcoCfg.FailOnPluginLoadingFails {
						return 0, fmt.Errorf("Plugin %s does not implement awaited interface\n", plugin.Name)			
					} else {
						return 0, nil
					}
				} else {

					log.Infof(">> Init plugin %s", plugin.Name)
					err := veradcoPlugin.Init(plugin.Configuration)

					if err != nil {
						log.Errorf("Unable to init plugin %s (skipped): %v", plugin.Name, err)
						if ! plugin.DryRun && veradcoCfg.FailOnPluginLoadingFails {
							return 0, err			
						} else {
							return 0, nil
						}
					} else {
						plugin.VeradcoPlugin = veradcoPlugin
						plugin.VeradcoPluginLoaded = true
						return 1, nil
					}
					// log.Infof("Plugin: %v\n", plugin)
				}
			}
		}
	}
}

func (veradcoCfg *VeradcoCfg) LoadPlugins() (int, error) {

	veradcoCfg.rwMutexPlugins.RLock()

	defer veradcoCfg.rwMutexPlugins.RUnlock()

	numberOfPluginsLoaded := 0
	log.Infof(">>>> Loading plugins")

	for _, plugin := range veradcoCfg.Plugins {
		nb, err := veradcoCfg.loadPlugin(plugin)
		if err != nil {
			return numberOfPluginsLoaded, err
		}
		numberOfPluginsLoaded += nb
	}

	if numberOfPluginsLoaded > 0 {
		log.Infof(">> %d plugins loaded over %d", numberOfPluginsLoaded, len(veradcoCfg.Plugins))
		return numberOfPluginsLoaded, nil
	}
	
	// Not an error to load 0 plugin because GRPC plugins can all be discovered
	return numberOfPluginsLoaded, nil
	// return numberOfPluginsLoaded, fmt.Errorf("No plugin loaded")
}

// func (veradcoCfg *VeradcoCfg) GetPlugins(r *admission.AdmissionRequest, kind string, operation string, namespace string, labels map[string]string, annotations map[string]string, scope string) (*[]*Plugin, error) {
func (veradcoCfg *VeradcoCfg) GetPlugins(r *admission.AdmissionRequest, scope string, endpoint string) (*[]*Plugin, error) {
	// log.Infof("Plugins: %v\n", veradcoCfg.Plugins)

	veradcoCfg.rwMutexPlugins.RLock()

	defer veradcoCfg.rwMutexPlugins.RUnlock()

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
		log.Infof(">> Execute plugin %s\n", plug.Name)
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