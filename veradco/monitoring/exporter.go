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
			Name: "plugin_execution_time",
			Help: "Plugin execution time.",
			Buckets: []float64{1000000, 10000000, 100000000, 500000000, 1000000000},
		},
		[]string{"plugin", "scope", "dry_run", "allowed", "group", "version", "kind", "name", "namespace", "operation"},
	)

	pluginExecutions = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "veradco_plugins_executions",
			Help: "Counter of plugins executions",
		},
		[]string{"plugin", "scope", "dry_run", "allowed", "group", "version", "kind", "name", "namespace", "operation"},
	)
)

func Init() {
	// Metrics have to be registered to be exposed:
	prometheus.MustRegister(pluginExecTime)
	prometheus.MustRegister(pluginExecutions)
}

func AddOperation(plugin string, scope string, dryRun bool, allowed	bool, group string, version string, kind string, name string, namespace string, operation string) {
	pluginExecutions.With(prometheus.Labels{"plugin": plugin, "scope": scope, "dry_run": strconv.FormatBool(dryRun), "allowed": strconv.FormatBool(allowed), "group": group, "version": version, "kind": kind, "name": name, "namespace": namespace, "operation": operation}).Inc()
}

func AddStat(plugin string, scope string, dryRun bool, allowed	bool, group string, version string, kind string, name string, namespace string, operation string, elapsed time.Duration) {
	pluginExecTime.With(prometheus.Labels{"plugin": plugin, "scope": scope, "dry_run": strconv.FormatBool(dryRun), "allowed": strconv.FormatBool(allowed), "group": group, "version": version, "kind": kind, "name": name, "namespace": namespace, "operation": operation}).Observe(float64(elapsed))
}

// func AddOperation(plug *conf.Plugin, r *admission.AdmissionRequest, result *admissioncontroller.Result) {
// 	pluginExecutions.With(prometheus.Labels{"plugin": plug.Name, "scope": plug.Scope, "dry_run": strconv.FormatBool(plug.DryRun), "allowed": strconv.FormatBool(result.Allowed), "group": r.Kind.Group, "version": r.Kind.Version, "kind": r.Kind.Kind, "name": r.Name, "namespace": r.Namespace, "operation": string(r.Operation)}).Inc()
// }

func StartMonitoringSvr() {
	http.Handle("/metrics", promhttp.Handler())
	log.Fatal(http.ListenAndServe(":8080", nil))
}