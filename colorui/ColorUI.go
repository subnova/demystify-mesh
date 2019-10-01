package main

import (
	"encoding/json"
	"fmt"
	"github.com/alecthomas/kingpin"
	"github.com/labstack/echo"
	"html/template"
	"io"
	"io/ioutil"
	"net/http"
)

var (
	remote1 = kingpin.Arg("remote1", "The first remote address to call").Required().String()
	remote2 = kingpin.Arg("remote2", "The second remote address to call").Required().String()
	address = kingpin.Flag("address", "The address to bind to").Default(":8080").String()
)

type Color struct {
	Title    string `json:"title"`
	Color    string `json:"color"`
	External *Color `json:"external,omitempty"`
}

func main() {
	kingpin.Parse()
	t := &Template{
		templates: template.Must(template.ParseGlob("*.html")),
	}
	e := echo.New()
	e.Renderer = t
	e.GET("/", ColorUI)
	err := e.Start(*address)
	if err != nil {
		panic(err)
	}
}

type RenderInfo struct {
	Color1 *Color
	Color2 *Color
}

func ColorUI(c echo.Context) error {
	color1, err := readColor(*remote1)
	if err != nil {
		fmt.Printf("unable to read color1: %v\n", err)
	}
	color2, err := readColor(*remote2)
	if err != nil {
		fmt.Printf("unable to read color2: %v\n", err)
	}

	return c.Render(200, "ui.html", RenderInfo{
		color1,
		color2,
	})
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

type Template struct {
	templates *template.Template
}

func (t *Template) Render(w io.Writer, name string, data interface{}, c echo.Context) error {
	return t.templates.ExecuteTemplate(w, name, data)
}
