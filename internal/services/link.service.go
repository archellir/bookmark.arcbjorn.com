package services

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"golang.org/x/net/html"
)

var retrySchedule = []time.Duration{
	1 * time.Second,
	3 * time.Second,
	10 * time.Second,
}

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

func (service *LinkService) getURLWithRetries(url string) (*http.Response, error) {
	var err error
	var resp *http.Response

	for _, retryInterval := range retrySchedule {
		resp, err = http.Get(url)

		if err == nil {
			break
		}

		fmt.Fprintf(os.Stderr, "Request error: %+v\n", err)
		fmt.Fprintf(os.Stderr, "Retrying in %v\n", retryInterval)
		time.Sleep(retryInterval)
	}

	// all retries failed
	if err != nil {
		return nil, err
	}

	return resp, nil
}

func validateUrl(urlString string) (isValid bool) {
	parsedUrl, err := url.Parse(urlString)
	return err == nil && parsedUrl.Scheme != "" && parsedUrl.Host != ""
}

func (service *LinkService) ValidateLink(url string) (isValid bool, err error) {
	isValid = validateUrl(url)
	if !isValid {
		return false, fmt.Errorf(ErrorTitleUrlNotStaticallyValid)
	}

	response, err := service.getURLWithRetries(url)
	if err != nil {
		return false, fmt.Errorf(ErrorTitleUrlNotValid + err.Error())
	}
	defer response.Body.Close()

	return true, nil
}

// adds protocol prefix if not found
// statically validates url
// dynamically validates url
// extracts document title as a name for bookmark

func (service *LinkService) ProcessLink(urlString string) (isValid bool, title string, err error) {
	url := urlString
	if !strings.Contains(urlString, "https://") {
		url = "https://" + url
	}

	isValid = validateUrl(url)
	if !isValid {
		return false, "", fmt.Errorf(ErrorTitleUrlNotStaticallyValid)
	}

	response, err := service.getURLWithRetries(url)
	if err != nil {
		return false, "", fmt.Errorf(ErrorTitleUrlNotValid + err.Error())
	}
	defer response.Body.Close()

	if isFound, title, err := service.getHtmlTitle(response.Body); err != nil {
		return true, "", err
	} else if isFound {
		return true, title, nil
	}

	return true, "", nil
}
