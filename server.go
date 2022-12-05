package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	owm "github.com/briandowns/openweathermap"
)

var apiKey = os.Getenv("OWM_API_KEY")
var namespace = os.Getenv("ENVIRONMENT")

var weather_template = `
Temperature in London, Canada is %f C <br/>
`

func get_weather_by_city_county_code(city string, county_code string) (*owm.CurrentWeatherData, error) {
	w, err := owm.NewCurrent("C", "EN", apiKey) // (internal - OpenWeatherMap reference for kelvin) with English output
	if err != nil {
		log.Fatalln(err)
	}

	w.CurrentByName(fmt.Sprintf("%s,%s", city, county_code))
	return w, nil
}

func weather_handler(w http.ResponseWriter, r *http.Request) {
	weather_data, err := get_weather_by_city_county_code("London", "CA")
	if err != nil {
		fmt.Fprint(w, http.StatusInternalServerError)
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
	w.Write([]byte(namespace))
}

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/", weather_handler)
	mux.HandleFunc("/ping", ping_handler)
	mux.HandleFunc("/health", healthcheck_handler)
	mux.HandleFunc("/version", version_handler)

	server := http.Server{
		Addr:         ":8080",
		Handler:      mux,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	fmt.Println("starting server at :8080")
	server.ListenAndServe()
}
