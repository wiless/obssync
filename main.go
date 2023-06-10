package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
)

type File struct {
	ID              string    `json:"id"`
	URL             string    `json:"url"`
	DownloadURL     string    `json:"download_url"`
	Name            string    `json:"name"`
	CreateTime      time.Time `json:"create_time"`
	MimeType        string    `json:"mime_type"`
	Status          string    `json:"status"`
	Size            int       `json:"size"`
	ContentEncoding any       `json:"content_encoding"`
}

type Notebook struct {
	Props struct {
		PageProps struct {
			InitialNotebook struct {
				Files         []File                   `json:"files"`
				Has_importers bool                     `json:"has_importers"`
				Nodes         []map[string]interface{} `json:"nodes"`
			} `json:"initialNotebook"`
		} `json:"pageProps"`
	} `json:"props"`
}

func CreateOJS(doc *goquery.Document) {

	// // Find the review items
	// doc.Find(".left-content article .post-title").Each(func(i int, s *goquery.Selection) {
	// 	// For each item found, get the title
	// 	title := s.Find("a").Text()
	// 	fmt.Printf("Review %d: %s\n", i, title)
	// })
	var obsnode Notebook

	nodebook := make(map[string]interface{})
	doc.Find("script").Each(func(i int, s *goquery.Selection) {
		// For each item found, get the title
		id, exists := s.Attr("id")
		// fmt.Printf("Attribute %v? %v\n", id, exists)
		if exists && id == "__NEXT_DATA__" {
			title := s.Text() // s.Find("a").Text()
			//	os.WriteFile("notebook.json", []byte(title), 0644)
			fmt.Print(json.Unmarshal([]byte(title), &nodebook))
			fmt.Print(json.Unmarshal([]byte(title), &obsnode))
			//	fmt.Printf("Review %d: %s\n", i, title)
		}

	})

	//	fmt.Printf("\n\nContent %v", nodebook["props"])

	var fw *os.File

	if output != "" {
		var ferr error
		fw, ferr = os.Create(output)
		log.Println("File Error : ", ferr)
		defer fw.Close()
	}

	// 	/*
	// FileAttachments:
	//   a.txt: ./a.txt
	//   gh.csv: ./gh.csv
	//   gv.csv: ./gv.csv
	//   gcsinfo.csv: ./gcsinfo.csv
	//   ueids.json: ./ueids.json
	//   ueids.csv: ./ueids.csv
	//   ueids2.csv: ./something.csv
	// */

	if len(obsnode.Props.PageProps.InitialNotebook.Files) > 0 {

		fw.WriteString("/*\nFileAttachments:")
		for _, f := range obsnode.Props.PageProps.InitialNotebook.Files {
			fmt.Fprintf(fw, "\n %v: ./%v", f.Name, f.Name)

		}
		fw.WriteString("\n*/\n\n")
	}

	log.Println("extracting cells  : ")
	for k, v := range obsnode.Props.PageProps.InitialNotebook.Nodes {
		fmt.Printf("%d ", k)
		// fmt.Printf("%s", v["value"])

		var cellstr string
		if v["mode"] == "md" {
			cellstr = fmt.Sprintf("md`%s`", v["value"])
			// fw.WriteString(cellstr)
		} else {

			cellstr = v["value"].(string)
		}
		// DATAflow requires absolute path for imported notebooks
		if strings.HasPrefix(cellstr, "import") {
			parts := strings.Split(cellstr, " ")
			etc, _ := strconv.Unquote(parts[len(parts)-1])
			fmt.Printf("\n Fixing  import %v \n ", etc)
			if !strings.HasPrefix(etc, "@") {
				etc = "d/" + etc
			}
			parts[len(parts)-1] = strconv.Quote("https://observablehq.com/" + etc)
			cellstr = strings.Join(parts, " ")
			// log.Printf("\n ======== PARTS \n %#v", strings.Join(parts, " "))
		}
		if output != "" {
			fw.WriteString(cellstr)
			fw.WriteString("\n\n")
		}
	}
	if output != "" {
		fw.Close()
	}

	if download {
		for indx, f := range obsnode.Props.PageProps.InitialNotebook.Files {
			fmt.Printf("\n\n%d Content Files %#v", indx, f.Name)
			if res, err := http.Get(f.DownloadURL); err == nil {
				if wr, err := os.Create(f.Name); err == nil {
					defer wr.Close()
					io.Copy(wr, res.Body)
				}
			}

		}
	} else {
		log.Println("The notebook contains file attachments but did not download.. Did you forget ?  -d true ")
	}

}

var url string
var filename string
var output string
var download bool

func init() {
	flag.StringVar(&url, "url", "", "Pass the url of the observablehq e.g. https://observablehq.com/d/e753000a71a027b2")
	flag.StringVar(&filename, "file", "obs.html", "html file downloaded from obs")
	flag.StringVar(&output, "o", "", "o file e.g. filename.ojs for dataflow")
	flag.BoolVar(&download, "d", false, "d Download files (default false)")
	flag.Parse()
}
func main() {
	var doc *goquery.Document
	if url != "" {
		// Request the HTML page.
		res, err := http.Get(url)
		if err != nil {
			log.Fatal(err)
		}
		defer res.Body.Close()
		if res.StatusCode != 200 {
			log.Printf("status code error: %d %s", res.StatusCode, res.Status)
		}
		doc, err = goquery.NewDocumentFromReader(res.Body)
		if err != nil {
			log.Fatal(err)
		}
	} else {
		if filename == "" {
			log.Print("Pass a valid filename ")
			return
		}
		data, err := os.ReadFile(filename)
		doc, err = goquery.NewDocumentFromReader(bytes.NewBuffer(data))
		if err != nil {
			log.Fatal(err)
		}

	}
	CreateOJS(doc)
}
