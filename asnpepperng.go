package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"regexp"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/go-rod/rod"
)

func main() {
	fs := flag.NewFlagSet(os.Args[0], flag.ExitOnError)

	orgParam := fs.String("org", "", "Filter by organization name")
	outputFileParam := fs.String("output", "", "Output file name")

	fs.Usage = func() {
		printCustomUsage()
	}

	fs.Parse(os.Args[1:])

	if *orgParam == "" {
		printCustomUsage()
		return
	}

	baseURL := "https://bgp.he.net/search?search%%5Bsearch%%5D=%s&commit=Search"
	url := fmt.Sprintf(baseURL, *orgParam)

	page := rod.New().MustConnect().MustPage(url)
	defer page.MustClose()

	page.WaitLoad()

	content := page.MustElement("body").MustHTML()

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(content))
	if err != nil {
		log.Fatal(err)
	}

	reCIDR := regexp.MustCompile(`[0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3}\/[0-9]{1,2}`)

	cidrOrgMap := make(map[string]string)

	doc.Find("tr").Each(func(i int, s *goquery.Selection) {
		cells := s.Find("td")
		if cells.Length() >= 2 {
			cidrHTML := cells.Eq(0).Find("a").Text()
			orgHTML := cells.Eq(1).Text()

			cidrs := reCIDR.FindAllString(cidrHTML, -1)

			if len(cidrs) > 0 {
				organization := strings.TrimSpace(orgHTML)
				cidr := cidrs[0]
				cidrOrgMap[cidr] = organization
			}
		}
	})

	for cidr, organization := range cidrOrgMap {
		fmt.Println(cidr, organization)
	}

	if *outputFileParam != "" {
		if err := saveOutputToFile(*outputFileParam, cidrOrgMap); err != nil {
			log.Fatal(err)
		}
	}
}

func printCustomUsage() {
	str := `
____ ____ _  _ ___  ____ ___  ___  ____ ____
|__| [__  |\ | |__] |___ |__] |__] |___ |__/
  |  ___] | \| |    |___ |    |    |___ |  \

Usages: ./asnpepper --org <organization>
        ./asnpepper --org <organization> --output <file.txt>
`
	fmt.Println(str)
}

func saveOutputToFile(filename string, data map[string]string) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	var cidrs []string
	for cidr := range data {
		cidrs = append(cidrs, cidr)
	}

	_, err = file.WriteString(strings.Join(cidrs, "\n"))
	if err != nil {
		return err
	}

	return nil
}