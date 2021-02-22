package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
)

func check(e error) {
	if e != nil {
		panic(e)
	}
}


var code_list = []string{	"ATA", "AKM", "BED", "BIL", "BIO", "BLG", "BUS", "CAB", "CEV", "CHZ", "CIE",
							"CMP", "COM", "DEN", "DFH", "DNK", "DUI", "EAS", "ECO", "ECN", "EHB", "EHN",
							"EKO", "ELE", "ELH", "ELK", "END", "ENR", "ESL", "ETH", "ETK", "EUT", "FIZ",
							"GED", "GEM", "GEO", "GID", "GMI", "GSB", "GUV", "HUK", "HSS", "ICM", "ILT",
							"IML", "ING", "INS", "ISE", "ISL", "ISH", "ITB", "JDF", "JEF", "JEO", "KIM",
							"KMM", "KMP", "KON", "MAD", "MAK", "MAL", "MAR", "MAT", "MCH", "MEK", "MEN",
							"MET", "MIM", "MOD", "MRE", "MRT", "MTO", "MTH", "MTM", "MTR", "MST", "MUH",
							"MUK", "MUT", "MUZ", "NAE", "NTH", "PAZ", "PEM", "PET", "PHE", "PHY", "RES",
							"SBP", "SES", "STA", "STI", "TDW", "TEB", "TEK", "TEL", "TER", "TES", "THO",
							"TUR", "UCK", "UZB", "YTO", "YZV",
}
var classes map[string][][]string

func main() {
	classes = make(map[string][][]string)

	APIURL := "http://www.sis.itu.edu.tr/TR/ogrenci/lisans/onsartlar/onsartlar.php"

	for _, code := range code_list{
		res, err := http.PostForm(APIURL, url.Values{
			"derskodu": {code},
		})

		fmt.Println(code)

		if err != nil {
			panic(err)
		}
		defer res.Body.Close()

		if res.StatusCode != 200 {
			log.Fatalf("status code error: %d %s", res.StatusCode, res.Status)
		}

		// Load the HTML document
		doc, err := goquery.NewDocumentFromReader(res.Body)
		if err != nil {
			log.Fatal(err)
		}

		done := make(chan int)
		class_ch := make(chan string)
		deps_ch := make(chan string)
		go func() {
			for class := range class_ch {
				deps := <-deps_ch
				deps = strings.ReplaceAll(deps, "MIN DD", "")
				deps = strings.ReplaceAll(deps, "MIN BB", "")
				deps = strings.ReplaceAll(deps, "Yok", "")
				ve := strings.Split(deps, "ve ")
				ves := [][]string{}

				for _, s := range ve {
					veyas := strings.Split(s, "veya ")
					use_these := []string{}
					for i := range veyas {
						veyas[i] = strings.Trim(veyas[i], "()")
						veyas[i] = strings.TrimSpace(veyas[i])
						if len(veyas[i]) <= 12 {
							use_these = append(use_these, veyas[i])
						}
					}
					ves = append(ves, use_these)
				}

				if _, ok := classes[class]; ok {
					//classes[class] = append(classes[class], []string{deps})
					panic("Fix here sometime")
				} else {
					classes[class] = ves
				}
			}
			done <- 0
		}()

		doc.Find(".table > tbody:nth-child(2) > tr").Each(func(i int, s *goquery.Selection) {
			s.Find("td").Each(func(i int, s *goquery.Selection) {
				sec := s.Text()
				switch i {
				case 0:
					class_ch <- sec
				case 2:
					deps_ch <- sec
				}
			})
		})
		close(class_ch)
		close(deps_ch)
		<-done
	}

	jsonString, err := json.Marshal(classes)

	check(err)


	f, err := os.Create("./test.json")
	check(err)

	defer f.Close()

	w := bufio.NewWriter(f)
	n4, err := w.WriteString(string(jsonString))
	check(err)
	fmt.Printf("wrote %d bytes\n", n4)

	w.Flush()
}
