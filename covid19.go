package main

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"time"

	"github.com/PuerkitoBio/goquery"
)

func check(e error) {
	if e != nil {
		panic(e)
	}
}

func writeToFile(fname string, data string) {
	ioutil.WriteFile(fname, []byte(data), 0644)
}

func readFromFile(fname string) string {
	if _, err := os.Stat(fname); err == nil {
		dat, err := ioutil.ReadFile(fname)

		check(err)

		return string(dat)
	}

	return "{}"
}

func convertToInterface(jsonStr string) map[string]interface{} {
	var covidObject map[string]interface{}

	err := json.Unmarshal([]byte(jsonStr), &covidObject)

	check(err)

	return covidObject
}

func parseJsonWithRegexp(script string) string {
	r, _ := regexp.Compile("{(.*?)}")
	jsonString := r.FindString(script)

	return jsonString
}

func prepareMessage(newCovidObject map[string]interface{}) string {
	msg := "Tarih: " + newCovidObject["tarih"].(string) +
		"\nGünlük test: " + newCovidObject["gunluk_test"].(string) +
		"\nGünlük vaka: " + newCovidObject["gunluk_vaka"].(string)

	return url.QueryEscape(msg)
}

func sendMessage(msg string) {
	userId := "telegramUserId"
	botApiKey := "telegramBotApiKey"

	res, err := http.Get("https://api.telegram.org/bot" + botApiKey + "/sendMessage?text=" + msg + "&chat_id=" + userId)

	check(err)

	defer res.Body.Close()

	if res.StatusCode != 200 {
		panic(res.StatusCode)
	}
}

func getDocFromPage(pageUrl string) *goquery.Document {
	res, err := http.Get(pageUrl)

	check(err)

	defer res.Body.Close()

	if res.StatusCode != 200 {
		panic(res.StatusCode)
	}

	doc, err := goquery.NewDocumentFromReader(res.Body)

	check(err)

	return doc
}

func main() {
	fname := "covid19.json"
	previousData := readFromFile(fname)
	previousCovidObject := convertToInterface(previousData)
	currentTime := time.Now()

	if currentTime.Format("02.01.2006") != previousCovidObject["tarih"] {
		doc := getDocFromPage("https://covid19.saglik.gov.tr/")
		script := doc.Find("script[type='text/javascript']").Last().Text()
		jsonString := parseJsonWithRegexp(script)
		newCovidObject := convertToInterface(jsonString)

		if newCovidObject["tarih"] != previousCovidObject["tarih"] {
			msg := prepareMessage(newCovidObject)

			sendMessage(msg)

			writeToFile(fname, jsonString)
		}
	}
}
