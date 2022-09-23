package main

import (
	admission "k8s.io/api/admission/v1"
	// v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"github.com/smart-duck/veradco"
	"fmt"
	"gopkg.in/yaml.v3"
	"regexp"
	"sync"
	"time"

	"net/http"
	"crypto/tls"
	"errors"
	"os"
	"encoding/json"
	"strings"

	"io"

	// meta "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var (
	name string = "HarborProxyCachePopulator"
)

type ProxyCache struct {
	RegexURL string `yaml:"regexURL"`
	ReplacementOCI string `yaml:"replacementOCI"`
	ReplacementArch string `yaml:"replacementArch"`
	TargetArch string `yaml:"targetArch"`
	TargetOS string `yaml:"targetOS"`
}

type HarborProxyCachePopulator struct {
	ProxyCaches []ProxyCache `yaml:"proxyCaches"`
	summary string `yaml:"-"`
	proceededImages     map[string]string `yaml:"-"`
	proceededImagesLock sync.Mutex `yaml:"-"`
	proceededImagesNb   int `yaml:"-"`
}

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

func (plug *HarborProxyCachePopulator) retrieveConfigurationIfAny(url string) *ProxyCache {
	for _, proxyCache := range plug.ProxyCaches {
		re := regexp.MustCompile(proxyCache.RegexURL)

		if re.MatchString(url) {
			return &proxyCache
		}
	}
	return nil
}

func (plug *HarborProxyCachePopulator) Init(configFile string) error {
	// Load configuration
	err := yaml.Unmarshal([]byte(configFile), plug)
	if err != nil {
		return err
	}
	if len(plug.ProxyCaches) == 0 {
		return fmt.Errorf("ProxyCaches list shall contain at least one element for plugin %s", name)
	}
	return nil
}

func (plug *HarborProxyCachePopulator) pullImage(url string) {
	for {
		plug.proceededImagesLock.Lock()
		if plug.proceededImagesNb < 1 {
			break
		}
		plug.proceededImagesLock.Unlock()
		// log.Printf("Wait for a slot is freed to pull %s\n", url)
		time.Sleep(5 * time.Second)
	}
	plug.proceededImagesNb++
	defer func() {
		plug.proceededImagesLock.Lock()
		defer plug.proceededImagesLock.Unlock()
		plug.proceededImagesNb--
	}()
	if _, ok := plug.proceededImages[url]; !ok {
		plug.proceededImages[url] = "proceeded"
		plug.proceededImagesLock.Unlock()
		// var errPull error = nil
		// // log.Printf("simulateDockerPull\n")
		// errPull = simulateDockerPull(url)
		
		// if errPull != nil {
		// 	// log.Printf("Error pulling image: %v\n", errPull)
		// 	plug.proceededImagesLock.Lock()
		// 	defer plug.proceededImagesLock.Unlock()
		// 	delete(plug.proceededImages, url)
		// }
	} else {
		// log.Printf("%s already met\n", url)
		plug.proceededImagesLock.Unlock()
	}
}

func (plug *HarborProxyCachePopulator) Execute(kobj runtime.Object, operation string, dryRun bool, r *admission.AdmissionRequest) (*admissioncontroller.Result, error) {

	// // kobj is supposed to be a pod...
	// pod, ok := kobj.(*v1.Pod)
	// if !ok {
	// 	plug.summary += fmt.Sprintf("Kubernetes resource is not a pod as expected (%s)", kobj.GetObjectKind().GroupVersionKind().Kind)
	// 	return nil, fmt.Errorf("Kubernetes resource is not a pod as expected (%s)", kobj.GetObjectKind().GroupVersionKind().Kind)
	// }

	// plug.summary = fmt.Sprintf("Execute plugin %s", name)

	// // Browse containers
	// for _, c := range pod.Spec.Containers {
	// 	// Browse replacements
	// 	for _, pcs := range plug.ProxyCaches {
	// 		find := fr.Find
	// 		replace := fr.Replace
			
	// 		re := regexp.MustCompile(find)
	// 		img := c.Image
	// 		if re.MatchString(img) {
	// 			imgNew := re.ReplaceAllString(img, replace)
	// 			plug.summary = fmt.Sprintf("Log %s", imgNew)
	// 			// add replace patch
	// 			// replaceOp := admissioncontroller.ReplacePatchOperation(fmt.Sprintf("/spec/containers/%d/image", i), imgNew)
	// 			// operations = append(operations, replaceOp)
	// 			// plug.summary += "\n" + fmt.Sprintf("Add repacement operation %v", replaceOp)
	// 			break
	// 		}
	// 	}
	// }

	// TODO: also browse init containers

	return &admissioncontroller.Result{Allowed: true}, nil
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

func simulateDockerPull(url string, regexURL string, replacementOCI string, replacementArch string, targetArch string, targetOS string) (error) {

	// https://harbor.registry.mine.io/v2/proxy_docker.io/library/nginx/manifests/1.22-perl

	// Modify to harbor URL:
	re := regexp.MustCompile(regexURL)

	// replacementOCI := "https://harbor.registry.mine.io/v2/$1/$2/manifests/$3"

	// replacementArch := "https://harbor.registry.mine.io/v2/$1/$2/manifests/ARCHDIGEST"

	if ! re.MatchString(url) {
		errMsg := fmt.Sprintf("%s URL is not as awaited", url)
		return errors.New(errMsg)
	}

	// Query architectures
	urlOCI := re.ReplaceAllString(url, replacementOCI)

	// retrieve AMD 64 digest
	resp, err := queryHarborDockerAPI(urlOCI)

	if err != nil {
		errMsg := fmt.Sprintf("Error querying %s URL: %v", urlOCI, err)
		return errors.New(errMsg)
	}
	// else {
	// 	log.Printf("Resp: %s\n", string(resp))
	// }

	// unmarshall response
	var ociManifest OCIManifest
	err = json.Unmarshal(resp, &ociManifest)

	if err != nil {
		errMsg := fmt.Sprintf("Unable to unmarshal json returned by URL %s: %v", urlOCI, err)
		return errors.New(errMsg)
	}

	fmt.Printf("ociManifest: %v\n", ociManifest)

	digest, err := retrieveTargetArchDigest(ociManifest, targetArch, targetOS)

	if err != nil {
		errMsg := fmt.Sprintf("%v", err)
		return errors.New(errMsg)
	}

	
	// Query target architecture image
	urlTargetImage := re.ReplaceAllString(url, replacementArch)

	urlTargetImage = strings.Replace(urlTargetImage, "ARCHDIGEST", digest, -1)

	fmt.Printf("urlTargetImage: %s\n", urlTargetImage)

	// target arch image
	_, err = queryHarborDockerAPI(urlTargetImage)

	if err != nil {
		errMsg := fmt.Sprintf("Error querying %s URL: %v", urlTargetImage, err)
		return errors.New(errMsg)
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
	return "", errors.New(errMsg)
}

// func main() {
	
// 	url := "https://amazonaws.com/proxy_docker.io/library/alpine:3.11.9"


// 	regex := "^.+amazonaws.com/(proxy_[^:/]+)/([^:]+):(.+$)"

// 	replacementOCI := "https://harbor.registry.mine.io/v2/$1/$2/manifests/$3"

// 	replacementArch := "https://harbor.registry.mine.io/v2/$1/$2/manifests/ARCHDIGEST"

// 	targetArch := "amd64"
	
// 	targetOS := "linux"

// 	// url := "https://amazonaws.com/proxy_docker.io/smartduck/veradco:0.1beta1"

// 	// url := "https://harbor.registry.mine.io/v2/proxy_docker.io/library/alpine/manifests/3.11"
// 	err := simulateDockerPull(url, regex, replacementOCI, replacementArch, targetArch, targetOS)

// 	if err != nil {
// 		log.Printf("Error: %v\n", err)
// 	}
// }

///////////////////
///////////////////
///////////////////
// Used for implicit pull
///////////////////
///////////////////
///////////////////