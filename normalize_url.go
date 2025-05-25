package main

import (
	"fmt"
	"log"
	"net/url"
	"strings"

	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
)

func normalizeURL(rawurl string) (string, error) {
	urlStruct, err := url.Parse(rawurl)
	if err != nil {
		return "", err
	}

	s := urlStruct.User.String() + urlStruct.Host + urlStruct.Path
	if len(s) > 0 && string(s[len(s)-1]) == "/" {
		s = s[:len(s)-1]
	}

	fmt.Println(s)
	return s, nil
}

func getURLsFromHTML(htmlBody string, baseURL *url.URL) ([]string, error) {
	reader := strings.NewReader(htmlBody)
	doc, err := html.Parse(reader)
	if err != nil {
		return nil, err
	}
	urls := make([]string, 0)

	if err != nil {
		log.Fatal(err)
	}

	var traverseNodes func(*html.Node)
	traverseNodes = func(n *html.Node) {
		if n.Type == html.ElementNode && n.DataAtom == atom.A {
			for _, a := range n.Attr {
				if a.Key == "href" {
					href, err := url.Parse(a.Val)
					if err != nil {
						continue
					}
					resolved := baseURL.ResolveReference(href)
					urls = append(urls, resolved.String())

				}
			}
		}

		for child := n.FirstChild; child != nil; child = child.NextSibling {
			traverseNodes(child)
		}

	}
	traverseNodes(doc)
	return urls, nil
}
