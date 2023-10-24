package golangtgbot

import (
	"encoding/xml"
	"fmt"
	"io"
	"os"
)

type Companies struct {
	XMLName   xml.Name  `xml:"companies"`
	Companies []Company `xml:"company"`
}

type Company struct {
	XMLName  xml.Name `xml:"company"`
	Category string   `xml:"category,attr"`
	Title    string   `xml:"title"`
	Owner    string   `xml:"owner"`
	Year     string   `xml:"year"`
	Links    Links    `xml:"links"`
}

type Links struct {
	XMLName xml.Name `xml:"links"`
	Links   []Link   `xml:"link"`
}

type Link struct {
	XMLName xml.Name `xml:"link"`
	Name    string   `xml:"name,attr"`
	Href    string   `xml:"link"`
}

func ParseCompFromXml(xmlPath string) Companies {
	// Open xmlFile
	xmlFile, err := os.Open(xmlPath)
	// if we os.Open returns an error then handle it
	if err != nil {
		fmt.Println(err)
	}

	// defer the closing of our xmlFile so that we can parse it later on
	defer xmlFile.Close()

	// read our opened xmlFile as a byte array.
	byteValue, _ := io.ReadAll(xmlFile)

	var companies Companies

	xml.Unmarshal(byteValue, &companies)

	return companies
}

func Filter(companies []Company, fn func(comp Company) bool) []Company {
	var filtered []Company
	for _, comp := range companies {
		if fn(comp) {
			filtered = append(filtered, comp)
		}
	}
	return filtered
}

func MapLinks(data []Link) []string {

	mapped := make([]string, len(data))

	for i, link := range data {
		mapped[i] = link.Href
	}

	return mapped
}
