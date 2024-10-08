package main

import (
	"encoding/xml"
)

type Node struct {
	XMLName xml.Name
	Attrs   []xml.Attr `xml:",any,attr"`
	Content []byte     `xml:",innerxml"`
	Nodes   []Node     `xml:",any"`
}

func parseXMLTree(xmlString string) (*Node, error) {
	data := []byte(xmlString)
	var root Node

	err := xml.Unmarshal(data, &root)
	if err != nil {
		return nil, err
	}

	return &root, nil
}
