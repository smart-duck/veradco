package conf

import (
	"io/ioutil"
	log "k8s.io/klog/v2"
	"gopkg.in/yaml.v3"
	veradcoplugin "github.com/smart-duck/veradco/plugin"
	goplugin "plugin"
	"fmt"
)

type Plugin struct {
	Name string `yaml:"name"`
	Path string `yaml:"path"`
	Code string `yaml:"code,omitempty"`
	Resources string `yaml:"resources"`
	Operations string `yaml:"operations"`
	Namespaces string `yaml:"namespaces"`
	Labels []map[string]string `yaml:"labels,omitempty"`
	Annotations []map[string]string `yaml:"annotations,omitempty"`
	DryRun bool `yaml:"dryRun"`
	Configuration string `yaml:"configuration"`
	Scope string `yaml:"scope"`
	VeradcoPlugin veradcoplugin.VeradcoPlugin `yaml:"-"`
}

type VeradcoCfg struct {
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

func (veradcoCfg *VeradcoCfg) LoadPlugins() error {
	numberOfPluginsLoaded := 0
	for _, plugin := range veradcoCfg.Plugins {
		log.Infof("Loading plugin  %s\n", plugin.Name)

		// Try to execute plugins
		plug, err := goplugin.Open(plugin.Path)
		if err != nil {
			log.Errorf("Unable to load plugin %s: %v\n", plugin.Name, err)
		} else {
			pluginHandler, err := plug.Lookup("VeradcoPlugin")
			if err != nil {
				log.Errorf("Unable to find handler for plugin %s: %v\n", plugin.Name, err)
			} else {
				var veradcoPlugin veradcoplugin.VeradcoPlugin

				veradcoPlugin, ok := pluginHandler.(veradcoplugin.VeradcoPlugin)
				if !ok {
					log.Errorf("Plugin %s does not implement awaited interface\n", plugin.Name)
				} else {
					plugin.VeradcoPlugin = veradcoPlugin
					numberOfPluginsLoaded++
					// log.Infof("Plugin: %v\n", plugin)
				}
			}
		}
	}

	if numberOfPluginsLoaded > 0 {
		log.Infof("%d plugins loaded over %d\n", numberOfPluginsLoaded, len(veradcoCfg.Plugins))
		return nil
	}
	return fmt.Errorf("No plugin loaded")
}

func (veradcoCfg *VeradcoCfg) GetPlugins(kind string, operation string, namespace string, labels map[string]string, annotations map[string]string, scope string) (*[]*Plugin) {
	// log.Infof("Plugins: %v\n", veradcoCfg.Plugins)
	return &veradcoCfg.Plugins
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