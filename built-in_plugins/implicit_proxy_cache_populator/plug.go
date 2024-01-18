package main

import (
	"fmt"
	"regexp"
	"sync"
	"time"

	admissioncontroller "github.com/smart-duck/veradco/admissioncontroller"
	"github.com/smart-duck/veradco/kres"
	"gopkg.in/yaml.v3"
	admission "k8s.io/api/admission/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"

	"strings"

	"context"
	"encoding/base64"

	"github.com/regclient/regclient"
	"github.com/regclient/regclient/config"
	"github.com/regclient/regclient/types/manifest"
	"github.com/regclient/regclient/types/platform"
	"github.com/regclient/regclient/types/ref"

	awsv2Config "github.com/aws/aws-sdk-go-v2/config"
	awsv2ECR "github.com/aws/aws-sdk-go-v2/service/ecr"
	awsv2EcrTypes "github.com/aws/aws-sdk-go-v2/service/ecr/types"

	log "k8s.io/klog/v2"
)

const (
	lookAheadTime         time.Duration = 120 * time.Second
	errNotAnEcrProxyCache               = "not an ecr proxy cache image"
	errNotAnImage                       = "not a docker image"
	// patternEcrProxyCache                = "(^[0-9]+\\.dkr\\.ecr\\.[^.]+\\.amazonaws.com)/proxy_([^/]+)/(.+$)"
	// patternEcrProxyCache = "(^[0-9]+\\.dkr\\.ecr\\.[^.]+\\.amazonaws.com)/(proxy_)(docker\\.io|docker\\.cloudsmith\\.io|docker\\.elastic\\.io|gcr\\.io|ghcr\\.io|k8s\\.gcr\\.io|public\\.ecr\\.aws|quay\\.io|registry\\.k8s\\.io|registry\\.opensource\\.zalan\\.do|us-docker\\.pkg\\.dev|xpkg\\.upbound\\.io)/([^:]+):?(.*)$"
	// ecrRegion            = "eu-central-1"
	// patternRepoTag                      = "(^[^:]+):?(.+)?"
	logLevel = 3
)

var (
	name string = "ImplicitProxyCachePopulator"
	ctx         = context.Background()
)

type ImplicitProxyCachePopulator struct {
	ProxyCaches             []ProxyCache      `yaml:"proxyCaches"`
	MaxNumberOfParallelJobs int               `yaml:"maxNumberOfParallelJobs"`
	summary                 string            `yaml:"-"`
	proceededImages         map[string]string `yaml:"-"`
	proceededImagesLock     sync.Mutex        `yaml:"-"`
	pullQueue               chan QueuedPull   `yaml:"-"`
}

type ProxyCache struct {
	PatternEcrProxyCache string `yaml:"patternEcrProxyCache"`
	Platform             string `yaml:"platform"`
	Region               string `yaml:"region"`
}

// /////////////////
// /////////////////
// Pull Queue
// /////////////////
// /////////////////
type QueuedPull struct {
	url           string
	configuration *ProxyCache
	dryRun        bool
}

///////////////////
///////////////////
// Pull Queue
///////////////////
///////////////////

func (plug *ImplicitProxyCachePopulator) PullQueueConsumer() {
	for {
		log.Infof("[IPCP] Wait for a replication in queue\n")
		item := <-plug.pullQueue
		plug.replicateImage(item.url, *item.configuration, item.dryRun)
	}
}

func (plug *ImplicitProxyCachePopulator) Init(configFile string) error {
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

func (plug *ImplicitProxyCachePopulator) Execute(kobj runtime.Object, operation string, dryRun bool, r *admission.AdmissionRequest) (*admissioncontroller.Result, error) {

	// kobj is supposed to be a pod...
	pod, ok := kobj.(*v1.Pod)
	if !ok {
		plug.summary += "\n" + fmt.Sprintf("Kubernetes resource is not a pod as expected (%s)", kobj.GetObjectKind().GroupVersionKind().Kind)

		if kobj.GetObjectKind().GroupVersionKind().Kind == "Pod" {
			plug.summary += "\n" + "In fact it is a pod, maybe you did not used the pods path. Trying to extract it again..."
			var err error
			pod, err = kres.ParsePod(r)
			if err != nil {
				plug.summary += "\n" + "Definitly, it is not a pod!"
				log.Errorf("[IPCP] Kubernetes resource is not a pod as expected (%s)\n", kobj.GetObjectKind().GroupVersionKind().Kind)
				return nil, fmt.Errorf("kubernetes resource is not a pod as expected (%s)", kobj.GetObjectKind().GroupVersionKind().Kind)
			}
		}
	}

	plug.summary = fmt.Sprintf("Execute plugin %s", name)

	// Browse Init containers
	for _, c := range pod.Spec.InitContainers {
		pProxyCache := plug.retrieveConfigurationIfAny(c.Image)

		if pProxyCache != nil {
			log.Infof("[IPCP] Check that image %s is in the proxy cache\n", c.Image)
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
			log.Infof("[IPCP] Check that image %s is in the proxy cache\n", c.Image)
			plug.summary += "\n" + fmt.Sprintf("Check that image %s is in the proxy cache", c.Image)
			// Put in the Queue channel
			plug.pullQueue <- QueuedPull{url: c.Image, configuration: pProxyCache, dryRun: dryRun}
			// go plug.pullImage(c.Image, *pProxyCache, dryRun)
		}
	}

	return &admissioncontroller.Result{Allowed: true}, nil
}

func (plug *ImplicitProxyCachePopulator) retrieveConfigurationIfAny(url string) *ProxyCache {
	for _, proxyCache := range plug.ProxyCaches {
		re := regexp.MustCompile(proxyCache.PatternEcrProxyCache)

		if re.MatchString(url) {
			return &proxyCache
		}
	}
	return nil
}

func (plug *ImplicitProxyCachePopulator) replicateImage(url string, configuration ProxyCache, dryRun bool) {

	plug.proceededImagesLock.Lock()

	if _, ok := plug.proceededImages[url]; !ok {
		plug.proceededImages[url] = "proceeded"
		plug.proceededImagesLock.Unlock()

		_, hardErr := replicateToEcr(url, configuration, dryRun)

		if hardErr != nil {
			log.Infof("[IPCP] Error replicating image: %v\n", hardErr)
			plug.proceededImagesLock.Lock()
			defer plug.proceededImagesLock.Unlock()
			delete(plug.proceededImages, url)
		}
	} else {
		log.Infof("[IPCP] %s already met and managed\n", url)
		plug.proceededImagesLock.Unlock()
	}
}

func replicateToEcr(imgDest string, configuration ProxyCache, dryRun bool) (error, error) {

	imgToReplicate := ImgCopy{
		EcrProxyImg: imgDest,
		Platform:    configuration.Platform,
		Region:      configuration.Region,
	}

	softErr := imgToReplicate.ParseInput(configuration.PatternEcrProxyCache)
	if softErr != nil {
		log.Infof("[IPCP] softErr=%v\n", softErr)
		return softErr, nil
	}

	_, err := imgToReplicate.GetManifest()

	if err != nil {
		log.Infof("[IPCP] err get manifest=%v\n", err)
	} else {
		// Manifest found
		log.Infof("[IPCP] Image %s%s/%s:%s already exists in registry %s\n", imgToReplicate.prefixEcrRepository, imgToReplicate.publicHostName, imgToReplicate.image, imgToReplicate.tag, imgToReplicate.ecrHostName)
		return nil, nil
	}

	log.Infof("[IPCP] Image %s%s/%s:%s does not exist in registry %s: replicate it\n", imgToReplicate.prefixEcrRepository, imgToReplicate.publicHostName, imgToReplicate.image, imgToReplicate.tag, imgToReplicate.ecrHostName)

	if dryRun { // Dry run
		log.Infof("[IPCP] Dry Run mode, exiting\n")
		return nil, nil
	} else {

		// Check that ecr repository exists or create it
		repo := fmt.Sprintf("%s%s/%s", imgToReplicate.prefixEcrRepository, imgToReplicate.publicHostName, imgToReplicate.image)
		err = imgToReplicate.CheckDockerRepoExistsOrCreateIt(repo)

		if err != nil {
			log.Infof("[IPCP] err repo exists=%v\n", err)
		}

		softErr, hardErr := imgToReplicate.ReplicateImage()

		if softErr != nil {
			log.Infof("[IPCP] softErr=%v\n", softErr)
		}

		if hardErr != nil {
			log.Infof("[IPCP] hardErr=%v\n", hardErr)
		}
	}

	return nil, nil

}

func (plug *ImplicitProxyCachePopulator) Summary() string {
	return plug.summary
}

// exported as symbol named "VeradcoPlugin"
var VeradcoPlugin ImplicitProxyCachePopulator

// /////////////////
// /////////////////
// /////////////////
// Used for implicit pull
// /////////////////
// /////////////////
// /////////////////
type ImgCopy struct {
	EcrProxyImg         string     `yaml:"imgDest"`
	Platform            string     `yaml:"platform"`
	Region              string     `yaml:"region"`
	ecrTokenExpiresAt   *time.Time `yaml:"-"`
	ecrUser             *string    `yaml:"-"`
	ecrPw               *string    `yaml:"-"`
	lockEcr             sync.RWMutex
	ecrHostName         string `yaml:"-"`
	prefixEcrRepository string `yaml:"-"`
	publicHostName      string `yaml:"-"`
	image               string `yaml:"-"`
	tag                 string `yaml:"-"`
}

func decodeAuth(authStr string) (string, string, error) {
	if authStr == "" {
		return "", "", nil
	}
	decoded, err := base64.StdEncoding.DecodeString(authStr)
	if err != nil {
		return "", "", err
	}
	userPass := strings.SplitN(string(decoded), ":", 2)
	if len(userPass) != 2 {
		return "", "", fmt.Errorf("invalid auth configuration file")
	}
	return userPass[0], strings.Trim(userPass[1], "\x00"), nil
}

// Retrieve ECR creds
func (ic *ImgCopy) getEcrCreds() (*string, *string, error) {
	ic.lockEcr.RLock()

	if ic.ecrTokenExpiresAt == nil || time.Now().After(*ic.ecrTokenExpiresAt) {
		// Token never retrieved or near to expire
		ic.lockEcr.RUnlock()
		ic.lockEcr.Lock()
		defer ic.lockEcr.Unlock()
	} else {
		// Token still valid
		ic.lockEcr.RUnlock()
		return ic.ecrUser, ic.ecrPw, nil
	}

	ic.ecrTokenExpiresAt = nil

	awsConfig, err := awsv2Config.LoadDefaultConfig(ctx, awsv2Config.WithRegion(ic.Region))
	if err != nil {
		return nil, nil, err
	}

	ecrSvc := awsv2ECR.NewFromConfig(awsConfig)

	result, err := ecrSvc.GetAuthorizationToken(ctx, &awsv2ECR.GetAuthorizationTokenInput{})
	if err != nil {
		return nil, nil, err
	}

	token := *result.AuthorizationData[0].AuthorizationToken
	// log.Printf("token=%v", token)
	expiresAt := result.AuthorizationData[0].ExpiresAt
	// log.Printf("expiresAt=%v", expiresAt)
	// proxyEndpoint := *result.AuthorizationData[0].ProxyEndpoint
	// log.Printf("proxyEndpoint=%v", proxyEndpoint)

	user, pw, err := decodeAuth(token)

	if err != nil {
		return nil, nil, err
	}

	tokenExpiresAt := (*expiresAt).Add(-lookAheadTime)

	ic.ecrTokenExpiresAt = &tokenExpiresAt
	ic.ecrUser = &user
	ic.ecrPw = &pw

	return ic.ecrUser, ic.ecrPw, nil
}

func (ic *ImgCopy) GetManifest() (*manifest.Manifest, error) {
	client, hardErr := ic.getRegClient()
	if hardErr != nil {
		return nil, hardErr
	}

	img := fmt.Sprintf("%s/%s%s/%s:%s", ic.ecrHostName, ic.prefixEcrRepository, ic.publicHostName, ic.image, ic.tag)

	r, err := ref.New(img)
	if err != nil {
		return nil, err
	}
	manif, err := client.ManifestGet(ctx, r)
	if err != nil {
		return nil, err
	}
	return &manif, nil
}

func (ic *ImgCopy) CheckDockerRepoExistsOrCreateIt(repo string) error {
	awsConfig, err := awsv2Config.LoadDefaultConfig(ctx, awsv2Config.WithRegion(ic.Region))
	if err != nil {
		return err
	}

	ecrSvc := awsv2ECR.NewFromConfig(awsConfig)

	var repoArray []string
	repoArray = append(repoArray, repo)
	// descRepoInput := awsv2ECR.DescribeRepositoriesInput{
	// 	RepositoryNames: repoArray,
	// }
	// awsv2ECR.DescribeRepositories(ctx, &descRepoInput, optFns ...func(*Options)) (*DescribeRepositoriesOutput, error)

	_, err = ecrSvc.DescribeRepositories(
		ctx,
		&awsv2ECR.DescribeRepositoriesInput{
			RepositoryNames: repoArray,
		},
		// nil,
		// func(o *ecr.Options) {
		// 				o.Region = region
		// },
	)
	if err != nil {
		// Does not exists
		_, err = ecrSvc.CreateRepository(
			ctx,
			&awsv2ECR.CreateRepositoryInput{
				RepositoryName: &repo,
				ImageScanningConfiguration: &awsv2EcrTypes.ImageScanningConfiguration{
					ScanOnPush: true,
				},
			},
		)
		if err != nil {
			debug(2, "CheckDockerRepoExistsOrCreateIt failed with repo %s: %v\n", repo, err)
			return err
		}
	}

	return nil
}

func (ic *ImgCopy) ParseInput(patternEcrProxyCache string) error {
	re := regexp.MustCompile(patternEcrProxyCache)
	parsed := re.FindStringSubmatch(ic.EcrProxyImg)
	debug(10, "parsed=%v\n", parsed)
	if len(parsed) == 6 {
		ic.ecrHostName = parsed[1]
		ic.prefixEcrRepository = parsed[2]
		ic.publicHostName = parsed[3]
		ic.image = parsed[4]
		ic.tag = parsed[5]
		if len(ic.tag) == 0 {
			ic.tag = "latest"
		}
		return nil
		// imgTag := parsed[3]
		// re := regexp.MustCompile(patternRepoTag)
		// parsed := re.FindStringSubmatch(imgTag)
		// if len(parsed) == 3 {
		// 	image := parsed[1]
		// 	tag := parsed[2]
		// 	if len(tag) == 0 {
		// 		tag = "latest"
		// 	}
		// 	return ecrHostName, publicHostName, image, tag, nil
		// } else {
		// 	debug(2, "Image %s is not as expected: %s\n", ic.EcrProxyImg, errNotAnImage)
		// 	return "", "", "", "", fmt.Errorf(errNotAnImage)
		// }
	} else {
		debug(2, "Image %s is not as expected: %s\n", ic.EcrProxyImg, errNotAnEcrProxyCache)
		return fmt.Errorf(errNotAnEcrProxyCache)
	}
}

func (ic *ImgCopy) getRegClient() (*regclient.RegClient, error) {
	user, pw, err := ic.getEcrCreds()
	if err != nil {
		debug(2, "getEcrCreds failed: %v\n", err)
		return nil, err
	}

	debug(10, "getEcrCreds succeeded: %s\n", *user)

	rcHosts := []config.Host{
		{
			Name:     ic.ecrHostName,
			Hostname: ic.ecrHostName,
			// Hostname: proxyEndpoint,
			TLS: config.TLSEnabled,
			// TLS:       config.TLSInsecure,
			ReqPerSec: 100,
			User:      *user,
			Pass:      *pw,
			// Token:     token,
		},
		{
			Name:      ic.publicHostName,
			Hostname:  ic.publicHostName,
			TLS:       config.TLSEnabled,
			ReqPerSec: 100,
		},
	}

	hostsOpt := regclient.WithConfigHost(rcHosts...)

	client := regclient.New(hostsOpt)

	return client, nil
}

func (ic *ImgCopy) ReplicateImage() (error, error) {

	client, hardErr := ic.getRegClient()

	if hardErr != nil {
		return nil, hardErr
	}

	imgOpts := []regclient.ImageOpts{}

	imgSrc := fmt.Sprintf("%s/%s:%s", ic.publicHostName, ic.image, ic.tag)
	imgDest := fmt.Sprintf("%s/%s%s/%s:%s", ic.ecrHostName, ic.prefixEcrRepository, ic.publicHostName, ic.image, ic.tag)

	debug(2, "imgSrc=%s\n", imgSrc)
	debug(2, "imgDest=%s\n", imgDest)

	sourceRef, err := ref.New(imgSrc)
	if err != nil {
		return nil, err
	}
	targetRef, err := ref.New(imgDest)
	if err != nil {
		return nil, err
	}

	p, err := platform.Parse(ic.Platform)
	if err != nil {
		return nil, err
	}
	m, err := client.ManifestGet(ctx, sourceRef)
	if err != nil {
		return nil, err
	}
	if m.IsList() {
		d, err := manifest.GetPlatformDesc(m, &p)
		if err != nil {
			return nil, err
		}
		sourceRef.Digest = d.Digest.String()
	}

	if err := client.ImageCopy(ctx, sourceRef, targetRef, imgOpts...); err != nil {
		debug(2, "Image %s unsuccessfuly replicated to %s: %v", imgSrc, imgDest, err)
		return nil, err
	}

	debug(2, "Image %s successfuly replicated to %s", imgSrc, imgDest)

	return nil, nil
}

func debug(level int, format string, a ...any) {
	if level < logLevel {
		log.Infof("[IPCP] " + format, a...)
	}
}

///////////////////
///////////////////
///////////////////
// Used for implicit pull
///////////////////
///////////////////
///////////////////

// Declared not to be bothered by a linter or something else
func main() {}
