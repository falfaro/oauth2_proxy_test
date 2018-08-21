// file: test.go
package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"regexp"

	"golang.org/x/net/html"

	"github.com/PuerkitoBio/goquery"
)

type Jar struct {
	cookies []*http.Cookie
}

func (j *Jar) SetCookies(u *url.URL, cookies []*http.Cookie) {
	for _, cookie := range cookies {
		j.cookies = append(j.cookies, cookie)
	}
}

func (j *Jar) Cookies(u *url.URL) (cookies []*http.Cookie) {
	return j.cookies
}

// Wrapper for http.Get() that shows some debugging information
func Get(baseUrl string) *http.Response {
	fmt.Printf("GET %s...\n", baseUrl)
	resp, err := http.Get(baseUrl)
	if err != nil {
		log.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		log.Fatal("Unexpected response code ", resp.StatusCode)
	}
	return resp
}

// Wrapper for http.PostForm() that shows some debugging information
func Post(baseUrl string, formParams url.Values) *http.Response {
	fmt.Printf("POST %s...\n", baseUrl)
	resp, err := http.PostForm(baseUrl, formParams)
	if err != nil {
		log.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		log.Fatal("Unexpected response code ", resp.StatusCode)
	}
	return resp
}

// Returns the URL that is used for OAuth2 authentication.
func getAuthLink(resp *http.Response, base *url.URL) *url.URL {
	r := regexp.MustCompile(`/dex/auth/mock`)
	for _, v := range getLinks(resp.Body) {
		if r.MatchString(v) {
			f, err := base.Parse(v)
			if err != nil {
				log.Fatal(err)
			}
			return f
		}
	}
	return nil
}

// Collects all links from response body and returns them as an
// array of strings.
func getLinks(body io.Reader) []string {
	var links []string
	z := html.NewTokenizer(body)
	for {
		tt := z.Next()

		switch tt {
		case html.ErrorToken:
			//todo: links list shoudn't contain duplicates
			return links
		case html.StartTagToken, html.EndTagToken:
			token := z.Token()
			if "a" == token.Data {
				for _, attr := range token.Attr {
					if attr.Key == "href" {
						links = append(links, attr.Val)
					}
				}
			}
		}
	}
}

// Ensures that the final response after authentication leads to the
// protected resource by inspecting the title of the returned HTML
// response and looking for the "Authorization Successful!" string.
func ensureAuthenticationSuccess(resp *http.Response) {
	const successString = "Authorization Successful!"

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		log.Fatal(err)
	}
	if titleString, _ := doc.Find("title").Html(); titleString != successString {
		log.Fatal("Unexpected HTML response (%s) != (%s)", titleString, successString)
	}
	fmt.Println("Success!")
}

// Collects all form buttons from response body and selects the one
// that corresponds to "Grant Access", and returns all form parameters
// related to it.
func getFormParams(resp *http.Response, base *url.URL) url.Values {
	// Load the HTML document
	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		log.Fatal(err)
	}

	formParams := url.Values{}
	doc.Find("form").Each(func(i int, s *goquery.Selection) {
		params := url.Values{}
		s.Find("input").Each(func(i int, s *goquery.Selection) {
			name, _ := s.Attr("name")
			value, _ := s.Attr("value")
			params.Set(name, value)
		})
		// This button correspnds to the "Grant Access" button...
		if params.Get("approval") == "approve" {
			formParams = params
		}
	})
	return formParams
}

func main() {
	// URL for the protected resource
	baseUrl := "http://172.30.0.4:4180"

	// Use our custom CookieJar for handling of cookies
	http.DefaultClient = &http.Client{Jar: &Jar{}}

	// Try to access protected resource, which will trigger a redirect to
	// the OAuth2 initial authentication mechanism from Dex (log-in with
	// e-mail and password or use the example/mock credentials)
	resp := Get(baseUrl)

	// Search for an HTTP link that looks like "/dex/auth/mock"
	authUrl := getAuthLink(resp, resp.Request.URL)
	if authUrl == nil {
		log.Fatal("No valid link to proceed with authentication was found in the response")
	}

	// Initiate OAuth2 authentication (mock) against Dex
	resp = Get(authUrl.String())

	// Parse parameters from the return form looking for the form button that
	// corresponds to the "Grant Access" and retrieves the form parameters
	formParams := getFormParams(resp, resp.Request.URL)
	resp = Post(resp.Request.URL.String(), formParams)
	if resp.StatusCode != http.StatusOK {
		log.Fatal("Grant Access failed!")
	}
	//Ensure final response matches the HTML from the protected resource
	ensureAuthenticationSuccess(resp)
}
