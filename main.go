package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"

	"github.com/tebeka/selenium"
)

type RespJson struct {
	Price float64
	Error bool
}

func main() {
	portServer := os.Args[1]
	if portServer == "" {
		portServer = "20333"
	}

	portSelenium := os.Args[2]
	if portSelenium == "" {
		portSelenium = "20334"
	}

	portSeleniumInt, _ := strconv.ParseInt(portSelenium, 10, 32)

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		url := r.URL.Query().Get("url")
		price, err := getPrice(url, int(portSeleniumInt))
		fmt.Println(err)
		resp := RespJson{price, err != nil}
		jsonStr, _ := json.Marshal(resp)
		w.Header().Set("Content-Type", "application/json")
		w.Write(jsonStr)
	})
	fmt.Println("Server is listening " + portServer)
	fmt.Println("Make request: http://localhost:" + portServer + "/?url={{url_price}}")
	http.ListenAndServe("localhost:"+portServer, nil)
}

func getPrice(url string, ports int) (price float64, err error) {
	const (
		// These paths will be different on your system.
		seleniumPath    = "vendor/selenium-server-standalone-3.14.0.jar"
		geckoDriverPath = "vendor/geckodriver-v0.23.0-linux64"
		port            = 8080
	)
	opts := []selenium.ServiceOption{
		selenium.StartFrameBuffer(),           // Start an X frame buffer for the browser to run in.
		selenium.GeckoDriver(geckoDriverPath), // Specify the path to GeckoDriver in order to use Firefox.
		selenium.Output(os.Stderr),            // Output debug information to STDERR.
	}
	selenium.SetDebug(false)
	service, err := selenium.NewSeleniumService(seleniumPath, port, opts...)
	if err != nil {
		return
	}
	defer service.Stop()

	// Connect to the WebDriver instance running locally.
	caps := selenium.Capabilities{
		"browserName": "firefox",
		"server":      "OFF",
		"browser":     "OFF",
		"client":      "OFF",
		"driver":      "OFF",
		"performance": "OFF",
		"profiler":    "OFF",
	}
	wd, err := selenium.NewRemote(caps, fmt.Sprintf("http://localhost:%d/wd/hub", port))
	if err != nil {
		return
	}
	defer wd.Quit()

	// Navigate to the simple playground interface.
	err = wd.Get(url)
	if err != nil {
		return
	}

	// Get a reference to the text box containing code.
	elem, err := wd.FindElement(selenium.ByCSSSelector, "[itemprop=\"price\"]")
	if err != nil {
		return
	}

	priceStr, err := elem.Text()
	if err != nil {
		return
	}

	price, err = strconv.ParseFloat(priceStr, 64)

	return
}
