package monitoring

import (
	"log"
	"net/http"

	"strconv"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"time"
)

var (

	pluginExecTime = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			// Namespace: "our_company",
			// Subsystem: "blob_storage",
			Name: "veradco_plugin_execution_time",
			Help: "Veradco plugin execution time.",
			Buckets: []float64{1000000, 10000000, 100000000, 500000000, 1000000000},
		},
		[]string{"plugin", "scope", "dry_run", "allowed", "group", "version", "kind", "name", "namespace", "operation", "error"},
	)

	pluginExecutions = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "veradco_plugins_executions",
			Help: "Counter of veradco plugins executions",
		},
		[]string{"plugin", "scope", "dry_run", "allowed", "group", "version", "kind", "name", "namespace", "operation", "error"},
	)
)

func Init() {
	// Metrics have to be registered to be exposed:
	prometheus.MustRegister(pluginExecTime)
	prometheus.MustRegister(pluginExecutions)
}

func AddOperation(plugin string, scope string, dryRun bool, allowed	bool, group string, version string, kind string, name string, namespace string, operation string, errStr string) {
	// Not used because it makes metrics too heavy: pods with random suffix
	name = "NotUsed"
	pluginExecutions.With(prometheus.Labels{"plugin": plugin, "scope": scope, "dry_run": strconv.FormatBool(dryRun), "allowed": strconv.FormatBool(allowed), "group": group, "version": version, "kind": kind, "name": name, "namespace": namespace, "operation": operation, "error": errStr}).Inc()
}

func AddStat(plugin string, scope string, dryRun bool, allowed	bool, group string, version string, kind string, name string, namespace string, operation string, errStr string, elapsed time.Duration) {
	// Not used because it makes metrics too heavy: pods with random suffix
	name = "NotUsed"
	pluginExecTime.With(prometheus.Labels{"plugin": plugin, "scope": scope, "dry_run": strconv.FormatBool(dryRun), "allowed": strconv.FormatBool(allowed), "group": group, "version": version, "kind": kind, "name": name, "namespace": namespace, "operation": operation, "error": errStr}).Observe(float64(elapsed))
}

// func AddOperation(plug *conf.Plugin, r *admission.AdmissionRequest, result *admissioncontroller.Result) {
// 	pluginExecutions.With(prometheus.Labels{"plugin": plug.Name, "scope": plug.Scope, "dry_run": strconv.FormatBool(plug.DryRun), "allowed": strconv.FormatBool(result.Allowed), "group": r.Kind.Group, "version": r.Kind.Version, "kind": r.Kind.Kind, "name": r.Name, "namespace": r.Namespace, "operation": string(r.Operation)}).Inc()
// }

func StartMonitoringSvr() {
	http.Handle("/metrics", promhttp.Handler())
	log.Fatal(http.ListenAndServe(":8080", nil))
}