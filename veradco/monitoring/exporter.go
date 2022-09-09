package monitoring

import (
	"log"
	"net/http"

	"strconv"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
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
	prometheus.MustRegister(pluginExecutions)
}

func AddOperation(plugin string, scope string, dryRun bool, allowed	bool, group string, version string, kind string, name string, namespace string, operation string) {
	pluginExecutions.With(prometheus.Labels{"plugin": plugin, "scope": scope, "dry_run": strconv.FormatBool(dryRun), "allowed": strconv.FormatBool(allowed), "group": group, "version": version, "kind": kind, "name": name, "namespace": namespace, "operation": operation}).Inc()
}

func StartMonitoringSvr() {
	http.Handle("/metrics", promhttp.Handler())
	log.Fatal(http.ListenAndServe(":8080", nil))
}