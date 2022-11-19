package services

import (
	"fmt"
	"io"
	"net/http"

	"golang.org/x/net/html"
)

type LinkService struct{}

func (service *LinkService) isTitleElement(n *html.Node) bool {
	return n.Type == html.ElementNode && n.Data == "title"
}

func (service *LinkService) traverseHtml(node *html.Node) (title string, isTitleFound bool) {
	if service.isTitleElement(node) {
		return node.FirstChild.Data, true
	}

	for child := node.FirstChild; child != nil; child = child.NextSibling {
		title, isTitleFound := service.traverseHtml(child)
		if isTitleFound {
			return title, isTitleFound
		}
	}

	return "", false
}

func (service *LinkService) getHtmlTitle(r io.Reader) (isFound bool, title string, err error) {
	document, err := html.Parse(r)
	if err != nil {
		return false, "", fmt.Errorf("can not parse html: %s", err.Error())
	}

	title, isFound = service.traverseHtml(document)

	return isFound, title, err
}

func (service *LinkService) ProcessLink(url string) (isValid bool, title string, err error) {
	response, err := http.Get(url)
	if err != nil {
		return false, "", fmt.Errorf("can not validate url: " + err.Error())
	}
	defer response.Body.Close()

	if isFound, title, err := service.getHtmlTitle(response.Body); err != nil {
		return true, "", err
	} else if isFound {
		return true, title, nil
	}

	return true, "", nil
}
