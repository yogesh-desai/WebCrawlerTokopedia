package main

import (
	"fmt"
	"golang.org/x/net/html"
	"net/http"
	"os"
	"strings"
)

// Helper function to pull the href attribute from a Token
func getHref(t html.Token) (ok bool, href string) {
	// Iterate over all of the Token's attributes until we find an "href"
	for _, a := range t.Attr {
		if a.Key == "href" {
			href = a.Val
			ok = true
		}
	}
	
	// "bare" return will return the variables (ok, href) as defined in
	// the function definition
	return
}

func getProductID(t html.Token) (ok bool, id, url string ){
/*	fmt.Println("Type: ", t.Type)
	fmt.Println("DataAtom: ", t.DataAtom)
	fmt.Println("t.Data: ", t.Data)
//	fmt.Println("t.Namespace: ", t.Namespace)
	fmt.Println("len(t.Attr): ", len(t.Attr))*/
	if t.Data == "input" {
		for _, a := range t.Attr{
			// Get product ID
			if a.Key == "name" && a.Val == "product_id" {
				for _,pid := range t.Attr{
					if pid.Key == "value" {
						id = pid.Val
						ok = true
			//			fmt.Println("\nProduct ID: ", id)
					} //product id
				}
			}
			
			// Get product URL
			if a.Key == "name" && a.Val == "product_link" {
				for _,pid := range t.Attr{
					if pid.Key == "value" {
						url = pid.Val
						ok = true
			//			fmt.Println("\nProduct url: ", url)
					} //product url
				}
			}

//		fmt.Println("a: ", a)
//		fmt.Println("a.Key: ",a.Key, "\t a.Val: ", a.Val, "\n" )
		}
	}
	return  //false, id, url
	
}

// Extract all http** links from a given webpage
func crawl(url string, ch chan string, chFinished chan bool) {
var pid, purl string
	resp, err := http.Get(url)

	defer func() {
		// Notify that we're done after this function
		chFinished <- true
	}()

	if err != nil {
		fmt.Println("ERROR: Failed to crawl \"" + url + "\"")
		return
	}

	b := resp.Body
	defer b.Close() // close Body when the function returns

	z := html.NewTokenizer(b)

	for {
		tt := z.Next()
//		fmt.Println("tt: ", tt)
//		fmt.Println("tt.String(): ", tt.String())

		switch {
		case tt == html.ErrorToken:
			// End of the document, we're done
			return
		case tt == html.StartTagToken:
			t := z.Token()
//			fmt.Println("t: ", t)

			// Check if the token is an <input> tag
		//	isInput := t.Data == "input"
		//	
		//	if !isInput {
		//		continue
		//	}

			// Extract the href value, if there is one
			//			ok, url := getHref(t)
			ok, tid, turl := getProductID(t)
			
			if !ok {
				continue
			}
			if ok {
				if len(tid) > 1 { pid = tid }
				if len(turl) > 1 { purl = turl}
//				fmt.Println("Product ID: ", id, "\nProduct URL: ", url)
			}

			if len(pid) > 1 && len(purl) > 1 {
				fmt.Println("Product ID: ", pid, "\nProduct URL: ", purl)
			}
			// Make sure the url begines in http**
			hasProto := strings.Index(purl, "http") == 0
			if hasProto {
			tmp := fmt.Sprintf("%v, %v", pid, purl)
//			tmp :=  pid + purl
			ch <- tmp
			}
		}
	}
}

func main() {
	foundUrls := make(map[string]bool)
	seedUrls := os.Args[1:]

	// Channels
	chUrls := make(chan string)
	chFinished := make(chan bool) 

	// Kick off the crawl process (concurrently)
	for _, url := range seedUrls {
		go crawl(url, chUrls, chFinished)
	}

	// Subscribe to both channels
	for c := 0; c < len(seedUrls); {
		select {
		case url := <-chUrls:
			foundUrls[url] = true
		case <-chFinished:
			c++
		}
	}

	// We're done! Print the results...

	fmt.Println("\nFound", len(foundUrls), "unique urls:\n")

	for url, _ := range foundUrls {
		fmt.Println(" - " + url)
	}

	close(chUrls)
}
