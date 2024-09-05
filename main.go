package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"
)

const openWeatherMapAPIKey = "3ff28abae14130d3f70c8f32decfcb2a"

type ForecastResponse struct {
	List []struct {
		Main struct {
			Temp     float64 `json:"temp"`
			Humidity int     `json:"humidity"`
		} `json:"main"`
		Wind struct {
			Speed float64 `json:"speed"`
		} `json:"wind"`
		Weather []struct {
			Description string `json:"description"`
		} `json:"weather"`
	} `json:"list"`
}

func fetchWeatherForecast(city string) (ForecastResponse, error) {
	url := fmt.Sprintf("http://api.openweathermap.org/data/2.5/forecast?q=%s&appid=%s&units=metric", city, openWeatherMapAPIKey)
	resp, err := http.Get(url)
	if err != nil {
		return ForecastResponse{}, err
	}
	defer resp.Body.Close()

	var forecast ForecastResponse
	err = json.NewDecoder(resp.Body).Decode(&forecast)
	if err != nil {
		return ForecastResponse{}, err
	}

	return forecast, nil
}

func getCurrentWeather(city string) (string, error) {
	url := fmt.Sprintf("http://api.openweathermap.org/data/2.5/weather?q=%s&appid=%s&units=metric", city, openWeatherMapAPIKey)
	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var weatherData map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&weatherData)
	if err != nil {
		return "", err
	}

	description := weatherData["weather"].([]interface{})[0].(map[string]interface{})["description"].(string)
	temp := weatherData["main"].(map[string]interface{})["temp"].(float64)
	humidity := weatherData["main"].(map[string]interface{})["humidity"].(float64)
	windSpeed := weatherData["wind"].(map[string]interface{})["speed"].(float64)

	return fmt.Sprintf("Current weather: %s, Temperature: %.2f°C, Humidity: %.0f%%, Wind Speed: %.2f m/s", description, temp, humidity, windSpeed), nil
}

func handler(w http.ResponseWriter, r *http.Request) {
	// Serve the HTML file
	http.ServeFile(w, r, "index.html")
}

func weatherHandler(w http.ResponseWriter, r *http.Request) {
	location := r.URL.Query().Get("location")

	forecast, err := fetchWeatherForecast(location)
	if err != nil {
		http.Error(w, "Failed to fetch weather forecast", http.StatusInternalServerError)
		return
	}

	currentWeather, err := getCurrentWeather(location)
	if err != nil {
		http.Error(w, "Failed to fetch current weather", http.StatusInternalServerError)
		return
	}

	// Respond with JSON containing weather information
	response := struct {
		Location       string             `json:"location"`
		CurrentWeather string             `json:"currentWeather"`
		Forecast       []ForecastListItem `json:"forecast"`
	}{
		Location:       location,
		CurrentWeather: currentWeather,
		Forecast:       make([]ForecastListItem, 5),
	}

	for i := 0; i < 5; i++ {
		date := time.Now().AddDate(0, 0, i).Format("2006-01-02")
		response.Forecast[i] = ForecastListItem{
			Date:        date,
			Temperature: forecast.List[i].Main.Temp,
			Humidity:    forecast.List[i].Main.Humidity,
			WindSpeed:   forecast.List[i].Wind.Speed,
			Description: forecast.List[i].Weather[0].Description,
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

type ForecastListItem struct {
	Date        string  `json:"date"`
	Temperature float64 `json:"temperature"`
	Humidity    int     `json:"humidity"`
	WindSpeed   float64 `json:"windSpeed"`
	Description string  `json:"description"`
}

func main() {
	http.HandleFunc("/", handler)
	http.HandleFunc("/weather", weatherHandler)
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	log.Printf("Server listening on port %s", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
