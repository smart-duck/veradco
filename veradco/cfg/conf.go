package conf

import (
	"io/ioutil"
	log "k8s.io/klog/v2"
	"gopkg.in/yaml.v3"
	veradcoplugin "github.com/smart-duck/veradco/plugin"
	goplugin "plugin"
	"fmt"
	"regexp"
	admission "k8s.io/api/admission/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"github.com/smart-duck/veradco"
	"github.com/smart-duck/veradco/kres"
	"github.com/smart-duck/veradco/monitoring"
)

type Plugin struct {
	Name string `yaml:"name"`
	Path string `yaml:"path"`
	Code string `yaml:"code,omitempty"`
	Kinds string `yaml:"kinds"`
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

func (veradcoCfg *VeradcoCfg) LoadPlugins() (int, error) {
	numberOfPluginsLoaded := 0
	log.Infof(">>>> Loading plugins\n")
	for _, plugin := range veradcoCfg.Plugins {
		log.Infof(">> Loading plugin  %s\n", plugin.Name)

		// Try to execute plugins
		plug, err := goplugin.Open(plugin.Path)
		if err != nil {
			log.Errorf("Unable to load plugin %s: %v\n", plugin.Name, err)
			if ! plugin.DryRun && veradcoCfg.FailOnPluginLoadingFails {
				return numberOfPluginsLoaded, err			
			}
		} else {
			pluginHandler, err := plug.Lookup("VeradcoPlugin")
			if err != nil {
				log.Errorf("Unable to find handler for plugin %s: %v\n", plugin.Name, err)
				if ! plugin.DryRun && veradcoCfg.FailOnPluginLoadingFails {
					return numberOfPluginsLoaded, err			
				}
			} else {
				var veradcoPlugin veradcoplugin.VeradcoPlugin

				veradcoPlugin, ok := pluginHandler.(veradcoplugin.VeradcoPlugin)
				if !ok {
					log.Errorf("Plugin %s does not implement awaited interface\n", plugin.Name)
					if ! plugin.DryRun && veradcoCfg.FailOnPluginLoadingFails {
						return numberOfPluginsLoaded, fmt.Errorf("Plugin %s does not implement awaited interface\n", plugin.Name)			
					}
				} else {

					log.Infof(">> Init plugin %s\n", plugin.Name)
					err := veradcoPlugin.Init(plugin.Configuration)

					if err != nil {
						log.Errorf("Unable to init plugin %s (skipped): %v", plugin.Name, err)
						if ! plugin.DryRun && veradcoCfg.FailOnPluginLoadingFails {
							return numberOfPluginsLoaded, err			
						}
					} else {
						plugin.VeradcoPlugin = veradcoPlugin
						numberOfPluginsLoaded++
					}
					// log.Infof("Plugin: %v\n", plugin)
				}
			}
		}
	}

	if numberOfPluginsLoaded > 0 {
		log.Infof(">> %d plugins loaded over %d\n", numberOfPluginsLoaded, len(veradcoCfg.Plugins))
		return numberOfPluginsLoaded, nil
	}
	return numberOfPluginsLoaded, fmt.Errorf("No plugin loaded")
}

// func (veradcoCfg *VeradcoCfg) GetPlugins(r *admission.AdmissionRequest, kind string, operation string, namespace string, labels map[string]string, annotations map[string]string, scope string) (*[]*Plugin, error) {
func (veradcoCfg *VeradcoCfg) GetPlugins(r *admission.AdmissionRequest, scope string) (*[]*Plugin, error) {
	// log.Infof("Plugins: %v\n", veradcoCfg.Plugins)

	log.V(2).Infof(">>>> GetPlugins called\n")

	result := []*Plugin{}

	// Browse all plugins to filter the relevant ones
	for _, plugin := range veradcoCfg.Plugins {

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

	log.Infof(">> Number of plugins selected: %d\n", len(result))

	return &result, nil
}

func (veradcoCfg *VeradcoCfg) ProceedPlugins(kobj runtime.Object, r *admission.AdmissionRequest, scope string) (*admissioncontroller.Result, error) {

	plugins, err := veradcoCfg.GetPlugins(r, scope)

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
		result, err := plug.VeradcoPlugin.Execute(kobj, string(r.Operation), *r.DryRun || plug.DryRun, r)
		if err == nil {
			monitoring.AddOperation(plug.Name, plug.Scope, plug.DryRun, result.Allowed, r.Kind.Group, r.Kind.Version, r.Kind.Kind, r.Name, r.Namespace, string(r.Operation))
			log.Infof(">> Plugin %s execution summary: %s\n", plug.Name, plug.VeradcoPlugin.Summary())
			if plug.DryRun {
				log.Infof(">> Plugin %s is in dry run mode. Nothing to do!\n", plug.Name)
			} else {
				if ! result.Allowed {
					return result, err
				} else {
					globalResult.Msg += result.Msg
					globalResult.PatchOps = append(globalResult.PatchOps, result.PatchOps...)
				}
			}
		} else {
			return result, err
		}
	}
	
	// return &admissioncontroller.Result{Allowed: true}, nil
	return &globalResult, nil

}

func matchRegex(regex string, value string) (*bool, error) {
	// regular expression act as a reverse pattern if it is prefixed by (!~)
	// By example, "(!~)(?i)test" matches that the value does not contain "test" whatever the case is.
	// log.Infof("Evaluate regex %s on %s\n", regex, value)
	
	appliedRegex := regex
	
	matched, err := regexp.MatchString(`\(!~\).+`, appliedRegex)
	if err != nil {
		log.Errorf("Evaluate regex %s on %s failed: %v\n", regex, value, err) 
		return nil, err
	}

	inverted := false

	if matched {
		appliedRegex = appliedRegex[4:]
		inverted = true
	}

	matched, err = regexp.MatchString(appliedRegex, value)
	if err != nil {
		log.Errorf("Evaluate regex %s from %s on %s failed: %v\n", appliedRegex, regex, value, err)
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