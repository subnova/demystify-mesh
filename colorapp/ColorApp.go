package main

import (
	"encoding/json"
	"fmt"
	"github.com/alecthomas/kingpin"
	"github.com/labstack/echo"
	"io/ioutil"
	"net/http"
)

var (
	title    = kingpin.Arg("title", "Title of the application").Required().String()
	color    = kingpin.Arg("color", "The color to return").Required().String()
	external = kingpin.Arg("external", "The address of a remote server to read data from").String()
	address  = kingpin.Flag("address", "The address to bind to").Default(":8080").String()
)

type Color struct {
	Title    string `json:"title"`
	Color    string `json:"color"`
	External *Color `json:"external,omitempty"`
}

func main() {
	kingpin.Parse()
	e := echo.New()
	e.GET("/", ColorApp)
	err := e.Start(*address)
	if err != nil {
		panic(err)
	}
}

func ColorApp(c echo.Context) error {
	var externalColor *Color

	if external != nil && *external != "" {
		externalColor, _ = readColor(*external)
	}

	return c.JSON(200, Color{*title, *color, externalColor})
}

func readColor(address string) (*Color, error) {
	var color Color

	resp, err := http.Get(address)
	if err != nil || resp == nil {
		return nil, fmt.Errorf("unable to retrieve data from %s: %v", address, err)
	}

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("unable to read response body: %v", err)
	}

	err = json.Unmarshal(data, &color)
	if err != nil {
		return nil, fmt.Errorf("unable to unmarshal data: %v", err)
	}

	return &color, nil
}
