package conf

import (
	"io/ioutil"
	log "k8s.io/klog/v2"
	"gopkg.in/yaml.v3"
)

type Plugin struct {
	Name string
	Path string
	Resources []string
	Operations []string
	Configuration string
	Scope string
}

type VeradcoCfg struct {
	Plugins []Plugin
}

func ReadConf(cfgFile string) (VeradcoCfg, error) {

	result := VeradcoCfg{}

	yfile, err := ioutil.ReadFile(cfgFile)

	if err != nil {

		log.Errorf("Failed to read configuration %s: %v", cfgFile, err)

		return result, err
	}

	err = yaml.Unmarshal(yfile, &result)

	if err != nil {
		log.Errorf("Failed to load configuration %s: %v", cfgFile, err)

		return result, err
	}

	return result, nil
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