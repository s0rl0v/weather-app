package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	owm "github.com/briandowns/openweathermap"
)

// We need an API key for OpenWeather
var apiKey = os.Getenv("OWM_API_KEY")
var namespace = os.Getenv("ENVIRONMENT")
var gitsha = os.Getenv("GITHUB_SHA")

// Template for response on /
var weather_template = `
Temperature in London, Canada is %f C <br/>
`

// Metrics
var totalRequests = prometheus.NewCounterVec(
	prometheus.CounterOpts{
		Name: "http_requests_total",
		Help: "Number of get requests.",
	},
	[]string{"path"},
)

var responseStatus = prometheus.NewCounterVec(
	prometheus.CounterOpts{
		Name: "response_status",
		Help: "Status of HTTP response",
	},
	[]string{"status"},
)

var httpDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
	Name: "http_response_time_seconds",
	Help: "Duration of HTTP requests.",
}, []string{"path"})

// Custom responseWriter for Prometheus middleware
type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func NewResponseWriter(w http.ResponseWriter) *responseWriter {
	return &responseWriter{w, http.StatusOK}
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

// Logic to fetch weather data
func get_weather_by_city_county_code(city string, county_code string) (*owm.CurrentWeatherData, error) {
	w, err := owm.NewCurrent("C", "EN", apiKey) // (internal - OpenWeatherMap reference for kelvin) with English output
	if err != nil {
		log.Fatalln(err)
	}

	w.CurrentByName(fmt.Sprintf("%s,%s", city, county_code))
	return w, nil
}

// Path handlers
func weather_handler(w http.ResponseWriter, r *http.Request) {
	weather_data, err := get_weather_by_city_county_code("London", "CA")
	// Check if something wrong got on OpenWeather side (mainly, incorrect API key etc)
	if err != nil || (weather_data.GeoPos.Latitude == 0.0 && weather_data.GeoPos.Longitude == 0.0) {
		fmt.Fprint(w, http.StatusInternalServerError)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	body := fmt.Sprintf(weather_template, weather_data.Main.Temp)

	w.Header().Set("Content-Type", "text/html")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(body))
}

func ping_handler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("PONG"))
}

func healthcheck_handler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status":"ok"}`))
}

func version_handler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(fmt.Sprintf("%s:%s", namespace, gitsha)))
}

func prometheusMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		route := mux.CurrentRoute(r)
		path, _ := route.GetPathTemplate()

		timer := prometheus.NewTimer(httpDuration.WithLabelValues(path))
		rw := NewResponseWriter(w)
		next.ServeHTTP(rw, r)

		statusCode := rw.statusCode

		responseStatus.WithLabelValues(strconv.Itoa(statusCode)).Inc()
		totalRequests.WithLabelValues(path).Inc()

		timer.ObserveDuration()
	})
}

func init() {
	prometheus.Register(totalRequests)
	prometheus.Register(responseStatus)
	prometheus.Register(httpDuration)
}

func main() {
	router := mux.NewRouter()
	// Inject middlware
	router.Use(prometheusMiddleware)

	// Prometheus endpoint
	router.Path("/metrics").Handler(promhttp.Handler())

	// Logic handlers
	router.HandleFunc("/", weather_handler)
	router.HandleFunc("/ping", ping_handler)
	router.HandleFunc("/health", healthcheck_handler)
	router.HandleFunc("/version", version_handler)

	fmt.Println("Serving requests on port 8080")
	err := http.ListenAndServe(":8080", router)
	log.Fatal(err)
}
