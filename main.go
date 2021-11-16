package main

import (
	"flag"
	"net/http"
	_ "net/http/pprof"
	"strings"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/ucef29/sensor-exporter/internal/sensors"
)

var (
	fanspeedDesc = prometheus.NewDesc(
		"sensor_lm_fan_speed_rpm",
		"fan speed (rotations per minute).",
		[]string{"fantype", "chip", "adaptor"},
		nil)

	voltageDesc = prometheus.NewDesc(
		"sensor_lm_voltage_volts",
		"voltage in volts",
		[]string{"intype", "chip", "adaptor"},
		nil)

	powerDesc = prometheus.NewDesc(
		"sensor_lm_power_watts",
		"power in watts",
		[]string{"powertype", "chip", "adaptor"},
		nil)

	temperatureDesc = prometheus.NewDesc(
		"sensor_lm_temperature_celsius",
		"temperature in celsius",
		[]string{"temptype", "chip", "adaptor"},
		nil)

	hddTempDesc = prometheus.NewDesc(
		"sensor_hddsmart_temperature_celsius",
		"temperature in celsius",
		[]string{"device", "id"},
		nil)
)

func main() {
	var (
		listenAddress = flag.String("web.listen-address", ":9255", "Address on which to expose metrics and web interface.")
		metricsPath   = flag.String("web.telemetry-path", "/metrics", "Path under which to expose metrics.")
	)
	flag.Parse()

	lmscollector := NewLmSensorsCollector()
	lmscollector.Init()
	prometheus.MustRegister(lmscollector)

	//http.Handle(*metricsPath, prometheus.Handler())
	http.Handle(*metricsPath, promhttp.Handler())

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`<html>
			<head><title>Sensor Exporter</title></head>
			<body>
			<h1>Sensor Exporter</h1>
			<p><a href="` + *metricsPath + `">Metrics</a></p>
			</body>
			</html>`))
	})
	http.ListenAndServe(*listenAddress, nil)
}

type (
	LmSensorsCollector struct{}
)

func NewLmSensorsCollector() *LmSensorsCollector {
	return &LmSensorsCollector{}
}

func (l *LmSensorsCollector) Init() {
	sensors.Init()
}

// Describe implements prometheus.Collector.
func (l *LmSensorsCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- fanspeedDesc
	ch <- powerDesc
	ch <- temperatureDesc
	ch <- voltageDesc
}

// Collect implements prometheus.Collector.
func (l *LmSensorsCollector) Collect(ch chan<- prometheus.Metric) {
	for _, chip := range sensors.GetDetectedChips() {
		chipName := chip.String()
		adaptorName := chip.AdapterName()
		for _, feature := range chip.GetFeatures() {
			if strings.HasPrefix(feature.Name, "fan") {
				ch <- prometheus.MustNewConstMetric(fanspeedDesc,
					prometheus.GaugeValue,
					feature.GetValue(),
					feature.GetLabel(), chipName, adaptorName)
			} else if strings.HasPrefix(feature.Name, "temp") {
				ch <- prometheus.MustNewConstMetric(temperatureDesc,
					prometheus.GaugeValue,
					feature.GetValue(),
					feature.GetLabel(), chipName, adaptorName)
			} else if strings.HasPrefix(feature.Name, "in") {
				ch <- prometheus.MustNewConstMetric(voltageDesc,
					prometheus.GaugeValue,
					feature.GetValue(),
					feature.GetLabel(), chipName, adaptorName)
			} else if strings.HasPrefix(feature.Name, "power") {
				ch <- prometheus.MustNewConstMetric(powerDesc,
					prometheus.GaugeValue,
					feature.GetValue(),
					feature.GetLabel(), chipName, adaptorName)
			}
		}
	}
}
