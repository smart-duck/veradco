package main

import (
	admission "k8s.io/api/admission/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"github.com/smart-duck/veradco"
	"github.com/smart-duck/veradco/kres"
	"fmt"
	"gopkg.in/yaml.v3"
	"regexp"
	"sync"
	"time"
	"math/rand"

	"net/http"
	"crypto/tls"
	"errors"
	"os"
	"encoding/json"
	"strings"

	"io"

	log "k8s.io/klog/v2"

	// meta "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var (
	name string = "HarborProxyCachePopulator"
)

type HarborProxyCachePopulator struct {
	ProxyCaches []ProxyCache `yaml:"proxyCaches"`
	MaxNumberOfParallelJobs int `yaml:"maxNumberOfParallelJobs"`
	summary string `yaml:"-"`
	proceededImages     map[string]string `yaml:"-"`
	proceededImagesLock sync.Mutex `yaml:"-"`
	pullQueue chan QueuedPull `yaml:"-"`
}

type ProxyCache struct {
	RegexURL string `yaml:"regexURL"`
	ReplacementOCI string `yaml:"replacementOCI"`
	ReplacementArch string `yaml:"replacementArch,omitempty"`
	TargetArch string `yaml:"targetArch"`
	TargetOS string `yaml:"targetOS"`
}

///////////////////
///////////////////
// Pull Queue
///////////////////
///////////////////
type QueuedPull struct {     
	url string
	configuration *ProxyCache
	dryRun bool
}
///////////////////
///////////////////
// Pull Queue
///////////////////
///////////////////


///////////////////
///////////////////
///////////////////
// For OCI json
///////////////////
///////////////////
///////////////////
type OCIManifest struct {     
	MediaType string `json:"mediaType"`
	SchemaVersion int `json:"schemaVersion"`
	Manifests []DigestPlatform `json:"manifests"`
}

type DigestPlatform struct {     
	Digest string `json:"digest"`
	MediaType string `json:"mediaType"`
	Size int `json:"size"`
	Platform Platform `json:"platform"`
}

type Platform struct {     
	Architecture string `json:"architecture"`
	OS string `json:"os"`
}
///////////////////
///////////////////
///////////////////
// For OCI json
///////////////////
///////////////////
///////////////////

func (plug *HarborProxyCachePopulator) PullQueueConsumer() {
	for {
		log.Infof("[HPCP] Wait for a pull request")
		log.Flush()
		item := <- plug.pullQueue
		plug.pullImageV2(item.url, *item.configuration, item.dryRun)
	}
}

func (plug *HarborProxyCachePopulator) Init(configFile string) error {
	// Create map of already successfully proceeded images
	plug.proceededImages = make(map[string]string)
	// Load configuration
	err := yaml.Unmarshal([]byte(configFile), plug)
	if err != nil {
		return err
	}
	if len(plug.ProxyCaches) == 0 {
		return fmt.Errorf("ProxyCaches list shall contain at least one element for plugin %s", name)
	}
	// Pull queue channel
	plug.pullQueue = make(chan QueuedPull)

	for i := 0; i < plug.MaxNumberOfParallelJobs; i++ {
		go plug.PullQueueConsumer()
	}
	return nil
}

func (plug *HarborProxyCachePopulator) Execute(kobj runtime.Object, operation string, dryRun bool, r *admission.AdmissionRequest) (*admissioncontroller.Result, error) {

	// Both environment variables hUSER and hPW shall be defined!!!
	_, okUser := os.LookupEnv("hUSER")
	_, okPw := os.LookupEnv("hPW")

	if ! okUser || ! okPw {
		plug.summary = fmt.Sprintf("hUSER and hPW environment variables shall be defined")
		log.Errorf("[HPCP] hUSER and hPW environment variables shall be defined")
		log.Flush()
		return nil, fmt.Errorf("Error: %s", plug.summary)
	}

	// kobj is supposed to be a pod...
	pod, ok := kobj.(*v1.Pod)
	if !ok {
		plug.summary += "\n" + fmt.Sprintf("Kubernetes resource is not a pod as expected (%s)", kobj.GetObjectKind().GroupVersionKind().Kind)
		
		if kobj.GetObjectKind().GroupVersionKind().Kind == "Pod" {
			plug.summary += "\n" + fmt.Sprintf("In fact it is a pod, maybe you did not used the pods path. Trying to extract it again...")
			var err error
			pod, err = kres.ParsePod(r)
			if err != nil {
				plug.summary += "\n" + fmt.Sprintf("Definitly, it is not a pod!")
				log.Errorf("[HPCP] Kubernetes resource is not a pod as expected (%s)", kobj.GetObjectKind().GroupVersionKind().Kind)
				log.Flush()
				return nil, fmt.Errorf("Kubernetes resource is not a pod as expected (%s)", kobj.GetObjectKind().GroupVersionKind().Kind)
			}
		}
	}

	plug.summary = fmt.Sprintf("Execute plugin %s", name)

	// Browse Init containers
	for _, c := range pod.Spec.InitContainers {
		pProxyCache := plug.retrieveConfigurationIfAny(c.Image)
		
		if pProxyCache != nil {
			log.Infof("[HPCP] Check that image %s is in the proxy cache", c.Image)
			log.Flush()
			plug.summary += "\n" + fmt.Sprintf("Check that image %s is in the proxy cache", c.Image)
			// Put in queue
			plug.pullQueue <- QueuedPull{url: c.Image, configuration: pProxyCache, dryRun: dryRun}
			// go plug.pullImage(c.Image, *pProxyCache, dryRun)
		}
	}

	// Browse containers
	for _, c := range pod.Spec.Containers {
		pProxyCache := plug.retrieveConfigurationIfAny(c.Image)
		
		if pProxyCache != nil {
			log.Infof("[HPCP] Check that image %s is in the proxy cache", c.Image)
			log.Flush()
			plug.summary += "\n" + fmt.Sprintf("Check that image %s is in the proxy cache", c.Image)
			// Put in the Queue channel
			plug.pullQueue <- QueuedPull{url: c.Image, configuration: pProxyCache, dryRun: dryRun}
			// go plug.pullImage(c.Image, *pProxyCache, dryRun)
		}
	}

	return &admissioncontroller.Result{Allowed: true}, nil
}

func (plug *HarborProxyCachePopulator) retrieveConfigurationIfAny(url string) *ProxyCache {
	for _, proxyCache := range plug.ProxyCaches {
		re := regexp.MustCompile(proxyCache.RegexURL)

		if re.MatchString(url) {
			return &proxyCache
		}
	}
	return nil
}

func (plug *HarborProxyCachePopulator) pullImageV2(url string, configuration ProxyCache, dryRun bool) {

	plug.proceededImagesLock.Lock()

	if _, ok := plug.proceededImages[url]; !ok {
		plug.proceededImages[url] = "proceeded"
		plug.proceededImagesLock.Unlock()
		var errPull error = nil
		errPull = pullImageFromProxyCache(url, configuration.RegexURL, configuration.ReplacementOCI, configuration.ReplacementArch, configuration.TargetArch, configuration.TargetOS, dryRun)
		
		if errPull != nil {
			log.Infof("[HPCP] Error pulling image: %v", errPull)
			plug.proceededImagesLock.Lock()
			defer plug.proceededImagesLock.Unlock()
			delete(plug.proceededImages, url)
			log.Flush()
		}
	} else {
		log.Infof("[HPCP] %s already met and managed", url)
		plug.proceededImagesLock.Unlock()
		log.Flush()
	}
}

func (plug *HarborProxyCachePopulator) Summary() string {
	return plug.summary
}

// exported as symbol named "VeradcoPlugin"
var VeradcoPlugin HarborProxyCachePopulator



///////////////////
///////////////////
///////////////////
// Used for implicit pull
///////////////////
///////////////////
///////////////////

func queryHarborDockerAPI(url string) ([]byte, error) {
	user := os.Getenv("hUSER")
	pw := os.Getenv("hPW")

	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		// log.Fatal(err)
		return nil, err
	}

	req.SetBasicAuth(user, pw)

	resp, err := client.Do(req)
	if err != nil {
		// log.Printf("Errored when sending request to the server: %v\n", err)
		return nil, err
	}

	defer resp.Body.Close()

	// log.Printf("queryHarborDockerAPI resp.Status = %v", resp.Status)
	// log.Println(string(responseBody))

	if resp.StatusCode == 200 {
		responseBody, err := io.ReadAll(resp.Body)
		if err != nil {
			// log.Fatal(err)
			return nil, err
		}
		return responseBody, nil
	} else {
		return nil, errors.New(fmt.Sprintf("Request failed with status code %d", resp.StatusCode))
	}

}

func pullImageFromProxyCache(url string, regexURL string, replacementOCI string, replacementArch string, targetArch string, targetOS string, dryRun bool) (error) {

	// Pulling an image from proxy cache triggers the pull from the cached registry if not already present.

	HARBORPCP_DEBUG := os.Getenv("HARBORPCP_DEBUG")

	if len(HARBORPCP_DEBUG) > 0 {
		rand.Seed(time.Now().UnixNano())
		n := 5 + rand.Intn(10)
		log.Infof("[HPCP] simulate pullImageFromProxyCache to pull %s - Wait %d seconds",url, n)
		log.Flush()
    time.Sleep(time.Duration(n)*time.Second)
		return nil
	}

	// https://harbor.registry.mine.io/v2/proxy_docker.io/library/nginx/manifests/1.22-perl

	// Modify to harbor URL:
	re := regexp.MustCompile(regexURL)

	// replacementOCI := "https://harbor.registry.mine.io/v2/$1/$2/manifests/$3"

	// replacementArch := "https://harbor.registry.mine.io/v2/$1/$2/manifests/ARCHDIGEST"

	if ! re.MatchString(url) {
		errMsg := fmt.Sprintf("%s URL is not as awaited", url)
		log.Error("[HPCP] " + errMsg)
		log.Flush()
		return errors.New(errMsg)
	}

	// Query architectures
	urlOCI := re.ReplaceAllString(url, replacementOCI)

	var resp []byte
	var err error

	if dryRun && len(replacementArch) == 0 {
		log.Infof("[HPCP] DRYRUN / Should query %s", urlOCI)
		log.Flush()
	} else {
		// retrieve AMD 64 digest
		resp, err = queryHarborDockerAPI(urlOCI)

		if err != nil {
			errMsg := fmt.Sprintf("Error querying %s URL: %v", urlOCI, err)
			log.Error("[HPCP] " + errMsg)
			log.Flush()
			return errors.New(errMsg)
		}
		// else {
		// 	log.Printf("Resp: %s\n", string(resp))
		// }

	}
	
	if len(replacementArch) > 0 {

		// unmarshall response
		var ociManifest OCIManifest
		err = json.Unmarshal(resp, &ociManifest)

		if err != nil {
			errMsg := fmt.Sprintf("Unable to unmarshal json returned by URL %s: %v", urlOCI, err)
			log.Error("[HPCP] " + errMsg)
			log.Flush()
			return errors.New(errMsg)
		}

		// log.Infof("ociManifest: %v\n", ociManifest)

		digest, err := retrieveTargetArchDigest(ociManifest, targetArch, targetOS)

		if err != nil {
			errMsg := fmt.Sprintf("%v", err)
			log.Error("[HPCP] " + errMsg)
			log.Flush()
			return errors.New(errMsg)
		}

		// Query target architecture image
		urlTargetImage := re.ReplaceAllString(url, replacementArch)

		urlTargetImage = strings.Replace(urlTargetImage, "ARCHDIGEST", digest, -1)

		if ! dryRun {
			log.Infof("[HPCP] urlTargetImage: %s\n", urlTargetImage)
			log.Flush()
			// target arch image
			_, err = queryHarborDockerAPI(urlTargetImage)

			if err != nil {
				errMsg := fmt.Sprintf("Error querying %s URL: %v", urlTargetImage, err)
				log.Error("[HPCP] " + errMsg)
				log.Flush()
				return errors.New(errMsg)
			}
		} else {
			log.Infof("[HPCP] DryRun urlTargetImage: %s", urlTargetImage)
			log.Flush()
		}
	}

	return nil
}

func retrieveTargetArchDigest(ociManifest OCIManifest, targetArch string, targetOS string) (string, error) {
	for _, digestPlatform := range ociManifest.Manifests {
		if digestPlatform.Platform.Architecture == "amd64" && digestPlatform.Platform.OS == "linux" {
			return digestPlatform.Digest, nil
		}
	}
	errMsg := fmt.Sprintf("Unable to retrieve digest for arch %s and OS %s", targetArch, targetOS)
	log.Error("[HPCP] " + errMsg)
	log.Flush()
	return "", errors.New(errMsg)
}

///////////////////
///////////////////
///////////////////
// Used for implicit pull
///////////////////
///////////////////
///////////////////