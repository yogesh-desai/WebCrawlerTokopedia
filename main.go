package main

import (

	"golang.org/x/net/html"
	"net/http"
	"net/url"
	"runtime"
	"strings"
	"context"
	"bytes"
	"flag"
	"sync"
	"time"
	"fmt"
	"log"
	"os"

//	"github.com/PuerkitoBio/fetchbot"
//	"github.com/PuerkitoBio/goquery"
	
	cdp "github.com/knq/chromedp"
	cdpr "github.com/knq/chromedp/runner"
	
)



var (
	baseurl, productFile, urlFile string
	wg sync.WaitGroup
	urls []string

	// Command-line flags
	seed        = flag.String("seed", "https://www.tokopedia.com/", "seed URL")
	cancelAfter = flag.Duration("cancelafter", 0, "automatically cancel the fetchbot after a given time")
	cancelAtURL = flag.String("cancelat", "", "automatically cancel the fetchbot at a given URL")
	stopAfter   = flag.Duration("stopafter", 1 * time.Minute, "automatically stop the fetchbot after a given time")
	stopAtURL   = flag.String("stopat", "", "automatically stop the fetchbot at a given URL")
	memStats    = flag.Duration("memstats", 5 * time.Minute, "display memory statistics at a given interval")

)
/*
func DoExtract(chanURL chan string){

	time.Sleep(2 * time.Millisecond)
	for{

		wg.Add(1)
		go func(){
			// Enqueue the url in chanURL
			//chanURL <- ctx.Cmd.URL().String()
			defer wg.Done()
			url := <- chanURL
			DoCDP(url)
		}()
		wg.Wait();
	}

}*/
func DoExtract(url string){

	time.Sleep(2 * time.Millisecond)
//	for{

		wg.Add(1)
		go func(){
			// Enqueue the url in chanURL
			//chanURL <- ctx.Cmd.URL().String()
			defer wg.Done()
//			url := <- chanURL
			DoCDP(url)
		}()
		wg.Wait();
//	}

}


func main() {

	flag.Parse()

	u, err := url.Parse(*seed)
	check(err, "Error in parsing the seed url")
	log.Println("The URL: ", u)

	baseurl = u.String()
	urlProcessor := make(chan string)
	done := make(chan bool)

	go processURL(urlProcessor, done)
//	go DoExtract(urlProcessor)
	urlProcessor <- u.String() //fmt.Sprint(u) //"https://jeremywho.com"

	// First mem stat print must be right after creating the fetchbot
	if *memStats > 0 {
		// Print starting stats
		printMemStats()
		// Run at regular intervals
		runMemStats(*memStats)
		// On exit, print ending stats after a GC
		defer func() {
			runtime.GC()
			printMemStats()
		}()
	}

	// if a stop or cancel is requested after some duration, launch the goroutine
	// that will stop or cancel.
	if *stopAfter > 0 || *cancelAfter > 0 {
		after := *stopAfter
		stopFunc := true
		if *cancelAfter != 0 {
			after = *cancelAfter
			stopFunc = true
		}

		go func() {
			c := time.After(after)
			<-c
			log.Println("The given timeout has occured. Exiting...")
			done <- stopFunc
		}()
	}

	<-done
	
	log.Println(strings.Repeat("=", 72) + "\n")
	log.Println("\n\nCompleted Crawling & Scrapping the Domain:\n", baseurl)

	// Print the product and URLs file details.
	outFileDetails()
	log.Println(strings.Repeat("=", 72) + "\n") 
}

func outFileDetails() {

	log.Println("Total no. of URLs processed: ", len(urls))
	
	if _, err := os.Stat(productFile); !os.IsNotExist(err) {

		log.Println("The output TSV file location: ", productFile)
		
	} else {
		log.Println("Required data is not present in any of the URLs in the crawled Domain.")
	}

	filePath := WriteProcessedUrlsToFile(urls)
	// Write the processed URLs to a file
	if _, err := os.Stat(urlFile); !os.IsNotExist(err){

		log.Println("The Processed URLs are in the file: ", filePath)
	} else {
		log.Println("Processed URLs fle doesn't exits.")
	}
}

func runMemStats(tick time.Duration) {
	var mu sync.Mutex
	go func() {
		c := time.Tick(tick)
		for _ = range c {
			mu.Lock()
			printMemStats()
			mu.Unlock()
		}
	}()
}

func printMemStats() {

	var mem runtime.MemStats
	runtime.ReadMemStats(&mem)
	buf := bytes.NewBuffer(nil)
	
	buf.WriteString(strings.Repeat("=", 72) + "\n") 
	buf.WriteString("Memory Profile:\n")
	buf.WriteString(fmt.Sprintf("\tAlloc: %d Kb\n", mem.Alloc/1024))
	buf.WriteString(fmt.Sprintf("\tTotalAlloc: %d Kb\n", mem.TotalAlloc/1024))
	buf.WriteString(fmt.Sprintf("\tNumGC: %d\n", mem.NumGC))
	buf.WriteString(fmt.Sprintf("\tGoroutines: %d\n", runtime.NumGoroutine()))
	buf.WriteString(strings.Repeat("=", 72))

	log.Println(buf.String())
}

// processURL checks the url is already visited or not.
//If not visited already, then set map = true and explore page for more links.
func processURL(urlProcessor chan string, done chan bool) {
	visited := make(map[string]bool)
	for {
		select {
		case url := <-urlProcessor:
			if _, ok := visited[url]; ok {
				continue
			} else {
				visited[url] = true
				urls = append(urls, url)
				go exploreURL(url, urlProcessor)
				DoExtract(url)
			}
		case <-time.After(15 * time.Second):
			log.Printf("Explored %d pages\n", len(visited))
			done <- true
			
		}
	}
}

// exploreURL does HTTP GET and tokenize the response
func exploreURL(url string, urlProcessor chan string) {
	log.Printf("Visiting %s.\n", url)

	resp, err := http.Get(url)
	if err != nil {
		log.Println(err)
		return
	}

	defer resp.Body.Close()
	z := html.NewTokenizer(resp.Body)

	for {
		tt := z.Next()
		if tt == html.ErrorToken {
			return
		}

		if tt == html.StartTagToken {
			t := z.Token()

			if t.Data == "a" {
				for _, a := range t.Attr {
					if a.Key == "href" {

						// if link is within jeremywho.com
						if strings.HasPrefix(a.Val, baseurl) {
							urlProcessor <- a.Val
						}
					}
				}
			}
		}
	}
}


//================================================================================
//================================================================================
// getProductInfo extract the required information by using chromedp package
func getProductInfo(urlstr, sel string, res *[]byte, pId, pUrl, url *string) cdp.Tasks {
	return cdp.Tasks{
		cdp.Navigate(urlstr),
		cdp.Sleep(5 * time.Second),
		cdp.WaitVisible(sel, cdp.ByID),
		cdp.EvaluateAsDevTools("document.getElementById('product-id').value;", pId),
		cdp.EvaluateAsDevTools("document.getElementById('product-url').value;", pUrl),
		cdp.EvaluateAsDevTools("document.getElementById('webyclip-widget-3').contentWindow.document.body.outerHTML;", res),
	}
}

// isPresent checks the existance of webyclip-widget-3 element.
func isPresent(url string, res *[]byte) cdp.Tasks {

	return cdp.Tasks{
		cdp.Navigate(url),
		cdp.Sleep(15 * time.Second),
//		cdp.EvaluateAsDevTools("document.getElementById('webyclip-thumbnails').childElementCount;", res),
		cdp.EvaluateAsDevTools("if (document.getElementById('webyclip-thumbnails')) {document.getElementById('webyclip-thumbnails').childElementCount;} else {console.log('0')}", res),
	}

}

//================================================================================
// getVideoLinks returns the Youtube viedo links present in the iframe webyclip-widget-3.
// returns all the links which are comma seperated.
func getVideoLinks(buf []byte) string {

	var videoLinks string

	//Convert byte buffer to String
	innerDoc	:= string(buf[:])
	tmp		:= strings.TrimSpace(innerDoc)

	//Find the videolinks and create one final string
	tmpStr		:= strings.Fields(tmp)
	matchStr	:= "i.ytimg.com/vi/"
	yUrl		:= "https://www.youtube.com/watch?v="

	for _, v := range tmpStr {

		//log.Println("Contains: ", strings.Contains(v, "i.ytimg.com"))
		if strings.Contains(v, matchStr) {

			vv := strings.TrimPrefix(v, "src=\\\"//i.ytimg.com/vi/")
			id := strings.Split(vv, "/")

			//log.Println("https://www.youtube.com/watch?v=" + id[0])
			//log.Println("id: \tlen:\n",len(id), id)

			youtubeLink := yUrl + id[0]
			videoLinks += youtubeLink + ","
		}

	}

	// return the video links
	return videoLinks[:len(videoLinks)-1]
}

//========================================================================================
func WriteToFile(filePath, record string) {

	f, err := os.OpenFile(filePath, os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		//                log.Println("File open failed for writing failure counts")
		//                return
		log.Println("File doesn't exists. File will be created with the headers before adding data.")
		// If file does not exists then create it with the header and write records.
		file, err1 := os.OpenFile(filePath, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0644)
		if err1 != nil {
			log.Println("File Open operation failed.")
			return
		}
		defer file.Close()

		header := fmt.Sprint("Product_ID" + "\t" + "Product_URL" + "\t" + "Youtube_Video_URLs")
		file.WriteString(fmt.Sprintf("%s\n", header))
		file.WriteString(fmt.Sprintf("%s\n", record))
		return

	}
	defer f.Close()

	log.Println("File exists Already. Adding the data for url.")
	f.WriteString(fmt.Sprintf("%s\n", record))
}

func getDomain() string {

	tmp		:= strings.TrimPrefix(baseurl, "https://www.")
	domain		:= strings.Split(tmp, ".")[0]

return domain

}
//================================================================================

func WriteProcessedUrlsToFile(urls []string) string{

	domain		:= getDomain()
	filePath	:= pwd() + "/" + domain + "-ProcessedURLs.csv"
	urlFile = filePath
	
	f, err := os.OpenFile(filePath, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0644)
	check(err, "Error in file Open operation")
	defer f.Close()

	for _, url := range urls {
		
		f.WriteString(fmt.Sprintf("%s\n", url))
		
	}
	return filePath
}

//================================================================================
// check checks the error, panics if not nil
func check(err error, str string){

        if err != nil { log.Fatalln(err, str) }
}

// pwd returns the current working directory through which the binary is invoked.
// used to save the csv file.
func pwd() string {
	
	pwd, err := os.Getwd()
	check(err, "Error in getting current workig dir.")
	return pwd
}
//================================================================================

func DoCDP(url string) {
	
	// create context
	ctxt, cancel := context.WithCancel(context.Background())
	defer cancel()

	// create chrome instance
//	c, err := cdp.New(ctxt, cdp.WithLog(log.Printf), cdp.WithRunnerOptions(cdpr.Flag("disable-web-security", "1")))

	// create chrome instance with cmd line options disable-web-security & headless
	//	c, err := cdp.New(ctxt, cdp.WithRunnerOptions(cdpr.Flag("disable-web-security", "1"), cdpr.Flag("headless", "1")))
		c, err := cdp.New(ctxt, cdp.WithRunnerOptions(cdpr.Flag("disable-web-security", "1")))
	check(err, "Error in creating new cdp instance")
	
	// run task list
	var buf, buf1 []byte
	var pId, pUrl string
	
	// Check for the existence of the webyclip-widget-3 on the page
	err = c.Run(ctxt, isPresent(url, &buf1))
	check(err, "Error in Run method of cdp")


	if (len(buf1) == 0) || (bytes.EqualFold([]byte("0"), buf1)){

		log.Println("No webyclip-widget-3 on page:\n ", url)

		// shutdown chrome
		err = c.Shutdown(ctxt)
		check(err, "Error in shutting down chrome")
	
		// wait for chrome to finish
		err = c.Wait()
		check(err, "Error in wait to shutdown chrome")
		
		return
		//os.Exit(0)

	} else { 
	
	//fmt.Println("In ELSE The status is: \t Len: ", len(buf), "\t", string(buf), " \t", buf)
	// Exit the code if "webyclip-widget-3" is not present.
		err = c.Run(ctxt, getProductInfo(url, `#webyclip-widget-3`, &buf, &pId, &pUrl, &url))
		check(err, "Error in Run method of cdp")

		// shutdown chrome
		err = c.Shutdown(ctxt)
		check(err, "Error in shutting down chrome")
	
		// wait for chrome to finish
		err = c.Wait()
		check(err, "Error in wait to shutdown chrome")
		
		pLinks		:= getVideoLinks(buf)
		record		:= fmt.Sprint(pId + "\t" + pUrl + "\t" + pLinks)
		domain		:= getDomain()
		filePath	:= pwd() + "/" + domain + "-ProductDetails.csv"

		productFile = filePath
		WriteToFile(filePath, record)
	}
}