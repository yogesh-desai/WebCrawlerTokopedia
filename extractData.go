// This script takes url as argument and extract the required data.
// Usage:
//./extractData.go <URL to process>
// Example:
// ./extractData.go http://www.tokopedia.com/chocoapple/ready-stock-bnib-iphone-128gb-7-plus-jet-black-garansi-apple-1-tahun-10?src=topads
//================================================================================
package main

import (
	"context"
	//	"io/ioutil"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	cdp "github.com/knq/chromedp"
	cdpr "github.com/knq/chromedp/runner"
)

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

//================================================================================
// getVideoLinks returns the Youtube viedo links present in the iframe webyclip-widget-3.
// returns all the links which are comma seperated.
func getVideoLinks(buf []byte) string {

	var videoLinks string
	//Convert byte buffer to String
	innerDoc := string(buf[:])
	tmp := strings.TrimSpace(innerDoc)

	//Find the videolinks and create one final string
	tmpStr := strings.Fields(tmp)
	matchStr := "i.ytimg.com/vi/"
	yUrl := "https://www.youtube.com/watch?v="

	for _, v := range tmpStr {

		//fmt.Println("Contains: ", strings.Contains(v, "i.ytimg.com"))
		if strings.Contains(v, matchStr) {

			vv := strings.TrimPrefix(v, "src=\\\"//i.ytimg.com/vi/")
			id := strings.Split(vv, "/")

			//fmt.Println("https://www.youtube.com/watch?v=" + id[0])
			//fmt.Println("id: \tlen:\n",len(id), id)

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
		//                fmt.Println("File open failed for writing failure counts")
		//                return
		fmt.Println("File doesn't exists. File will be created with the headers before adding data.")
		// If file does not exists then create it with the header and write records.
		file, err1 := os.OpenFile(filePath, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0644)
		if err1 != nil {
			fmt.Println("File Open operation failed.")
			return
		}
		defer file.Close()

		header := fmt.Sprint("Product_ID" + "\t" + "Product_URL" + "\t" + "Youtube_Video_URLs")
		file.WriteString(fmt.Sprintf("%s\n", header))
		file.WriteString(fmt.Sprintf("%s\n", record))
		return

	}
	defer f.Close()
	fmt.Println("File exists Already. Adding the data.")
	f.WriteString(fmt.Sprintf("%s\n", record))
}

//================================================================================

// HasValidCLI checks the validity of required no. of arguments to run program
func HasValidCLI(Args []string, maxNo int, msg string) bool {

	if (len(Args) == 1) || (len(Args) > maxNo) {
		Usage(msg)
		return false
	}

	return true
}

// Usage prints the Usage.
func Usage(msg string) {

	fmt.Println("\n Usage: \n", os.Args[0], msg)
}

// ValidArgs checks the validity of the arguments.
func ValidateArgs() {

	if !strings.HasPrefix(os.Args[1], "https://www.tokopedia.com/") {
		os.Exit(1)
	}

}

// check checks the error, panics if not nil
func check(err error){

        if err != nil { log.Fatal(err) }
}

// pwd returns the current working directory through which the binary is invoked.
// used to save the csv file.
func pwd() string {
	
	pwd, err := os.Getwd()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	return pwd
}

//========================================================================================
func main() {

	msg := fmt.Sprintln("URL to process.\nExample:\n./extractData.go http://www.tokopedia.com/chocoapple/ready-stock-bnib-iphone-128gb-7-plus-jet-black-garansi-apple-1-tahun-10?src=topads\n")

	if !HasValidCLI(os.Args, 2, msg) { os.Exit(1) }
	ValidateArgs()

	//url := "https://www.tokopedia.com/chocoapple/ready-stock-bnib-iphone-128gb-7-plus-jet-black-garansi-apple-1-tahun-10?src=topads";
	url := os.Args[1]

	// create context
	ctxt, cancel := context.WithCancel(context.Background())
	defer cancel()

	// create chrome instancefunc(map[string]interface{}) error
	c, err := cdp.New(ctxt, cdp.WithRunnerOptions(cdpr.Flag("disable-web-security", "1")))
	check(err)
	
	// run task list
	var buf []byte
	var pId, pUrl string
	err = c.Run(ctxt, getProductInfo(url, `#webyclip-widget-3`, &buf, &pId, &pUrl, &url))
	check(err)

	// shutdown chrome
	err = c.Shutdown(ctxt)
	check(err)
	
	// wait for chrome to finish
	err = c.Wait()
	check(err)

	pLinks := getVideoLinks(buf)
	record := fmt.Sprint(pId + "\t" + pUrl + "\t" + pLinks)
	filePath :=  pwd() + "/TokoProductDetails.csv"
	WriteToFile(filePath, record)
}

//================================================================================
