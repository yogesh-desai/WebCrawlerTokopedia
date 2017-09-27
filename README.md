# WebCrawlerTokopedia
It is a web crawler and scrapper for https://www.Tokopedia.com. It is fully automated code where you just need to give input URL to get started.

The program extract the following,

* product-ID,
* product-URL,
* product-videos-URLs

It has fetcher and extractor functions. The strucutre of the webpage is considered and the code is written specifically for that purpose. One need to change the extractor, `DoCDP()` function to get the required results.


## Dependencies

It uses the `chromdp` package. You can check it [here](https://github.com/knq/chromedp).

## Usage

```
$ go run main.go

```

## Output

The code generates two files.
    1. File to store product details.
    2. File to store visited URLs.

Following is the example of the code when ran for a single webpage.

```

Product_ID	Product_URL	Youtube_Video_URLs
146347138	https://www.tokopedia.com/chocoapple/ready-stock-bnib-iphone-128gb-7-plus-jet-black-garansi-apple-1-tahun-10	https://www.youtube.com/watch?v=oKR2fh09Nic,https://www.youtube.com/watch?v=12JBG20n3jI,https://www.youtube.com/watch?v=mWEG1nu2rVY,https://www.youtube.com/watch?v=wgZ7Q4ywOl8

```

## Features

* It has fecther and extractor functions.
* It uses goroutines and channels to make tasks parallel and faster.
* It has Flags, with bydefault values. You can give your own values at runtime.
* It also has the Memory Stats to keep track of memory being used by the program.

## TODOs

* Currently, it uses GUI mode of the Google-Chrome. Need to implement the `--headless` functionality.
* Make the code more Faster and stable.
* More Testing and profiling.


Please feel free to generate pull requests or issues. :)
