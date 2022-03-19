package fetcher

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"github.com/fr4nki/lnkElectricityBot/bot/helpers"
	"io"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"time"
)

const (
	htmlUrl        = "https://cebcare.ceb.lk/Incognito/DemandMgmtSchedule"
	dirtyCurrentTZ = "+05:30"
)

type ForecastItem struct {
	StartTime       string `json:"startTime"`
	EndTime         string `json:"endTime"`
	LoadShedGroupID string `json:"loadShedGroupId"`
}

func getForecast(token string, cookie http.Cookie) ([]ForecastItem, error) {
	var items []ForecastItem

	fromTime, toTime, timeErr := helpers.GetDateTimeRange(1)
	if timeErr != nil {
		fmt.Println(timeErr)
		return nil, timeErr
	}

	from := fromTime.Format("2006-01-02")
	to := toTime.Format("2006-01-02")

	client := &http.Client{}

	params := url.Values{}
	params.Set("StartTime", from)
	params.Set("EndTime", to)

	body := bytes.NewBufferString(params.Encode())

	req, respErr := http.NewRequest("POST", "https://cebcare.ceb.lk/Incognito/GetLoadSheddingEvents", body)

	req.Header.Add("Connection", "keep-alive")
	req.Header.Add("Pragma", "no-cache")
	req.Header.Add("Cache-Control", "no-cache")
	req.Header.Add("Requestverificationtoken", token)
	req.Header.Add("Accept-Language", "en-GB,en-US;q=0.9,en;q=0.8")
	req.Header.Add("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/99.0.4844.51 Safari/537.36")
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded; charset=UTF-8")
	req.Header.Add("Accept", "application/json, text/javascript, */*; q=0.01")
	req.Header.Add("X-Requested-With", "XMLHttpRequest")
	req.Header.Add("Sec-Gpc", "1")
	req.Header.Add("Origin", "https://cebcare.ceb.lk")
	req.Header.Add("Sec-Fetch-Site", "same-origin")
	req.Header.Add("Sec-Fetch-Mode", "cors")
	req.Header.Add("Sec-Fetch-Dest", "empty")
	req.Header.Add("Referer", "https://cebcare.ceb.lk/Incognito/DemandMgmtSchedule")
	req.Header.Add("Cookie", fmt.Sprintf("%s=%s", cookie.Name, cookie.Value))

	resp, respErr := client.Do(req)
	if respErr != nil {
		fmt.Println(respErr)
		return nil, respErr
	}

	defer resp.Body.Close()

	data, dataErr := io.ReadAll(resp.Body)
	if dataErr != nil {
		fmt.Println(dataErr)
		return nil, dataErr
	}

	unmarshalError := json.Unmarshal(data, &items)
	if unmarshalError != nil {
		fmt.Println(unmarshalError)
		return nil, unmarshalError
	}

	return items, nil
}

func getRequestToken() (string, http.Cookie, error) {
	signalRCookie := http.Cookie{}

	resp, respErr := http.Get(htmlUrl)
	if respErr != nil {
		return "", signalRCookie, respErr
	}

	defer resp.Body.Close()

	for _, cookie := range resp.Cookies() {
		if strings.HasPrefix(cookie.Name, ".AspNetCore.Antiforgery") {
			signalRCookie.Name = cookie.Name
			signalRCookie.Value = cookie.Value
		}
	}

	if signalRCookie.Name == "" && signalRCookie.Value == "" {
		return "", signalRCookie, errors.New("cannot parse cookie from request")
	}

	doc, docErr := goquery.NewDocumentFromReader(resp.Body)
	if docErr != nil {
		return "", signalRCookie, docErr
	}

	token, ok := doc.Find("input[name=__RequestVerificationToken]").Attr("value")
	if !ok {
		return "", signalRCookie, errors.New("input selector not find")
	}

	return token, signalRCookie, nil
}

func getForecastByArea(items []ForecastItem, area string) (string, error) {
	text := ""
	ranges := make(map[string][]time.Time)

	fromTime, _, timeError := helpers.GetDateTimeRange(1)
	if timeError != nil {
		fmt.Println(timeError)
		text += "Ошибка :("

		return text, timeError
	}

	text += fmt.Sprintf("%s. Группа %s \n\n", fromTime.Format("02.01.2006"), area)

	for _, item := range items {
		if strings.ToLower(item.LoadShedGroupID) == strings.ToLower(area) {
			// It's because api not providing Z within datetime to parse as RFC3339
			sTime := fmt.Sprintf("%s%s", item.StartTime, dirtyCurrentTZ)
			eTime := fmt.Sprintf("%s%s", item.EndTime, dirtyCurrentTZ)
			sTimeParsed, sTimeParsedErr := time.Parse(time.RFC3339, sTime)
			eTimeParsed, eTimeParsedErr := time.Parse(time.RFC3339, eTime)

			if sTimeParsedErr != nil || eTimeParsedErr != nil {
				text += "Ошибка :("
				fmt.Println(sTimeParsedErr, eTimeParsedErr)
			}

			ranges[item.StartTime] = []time.Time{sTimeParsed, eTimeParsed}
		}
	}

	if len(ranges) > 0 {
		text += "Света не будет:\n"
	}

	keys := make([]string, 0, len(ranges))
	for k := range ranges {
		keys = append(keys, k)
	}

	sort.Strings(keys)

	for _, k := range keys {
		text += fmt.Sprintf("c %s по %s \n", ranges[k][0].Format("15:04"), ranges[k][1].Format("15:04"))
	}

	return text, nil
}

func Forecast(area string) (string, error) {
	token, cookie, tokenError := getRequestToken()
	if tokenError != nil {
		fmt.Println(tokenError)
		return "", errors.New("can't get request token")
	}

	items, itemsError := getForecast(token, cookie)
	if itemsError != nil {
		fmt.Println(itemsError)
		return "", errors.New("can't get forecast items")
	}

	text, textError := getForecastByArea(items, area)
	if textError != nil {
		fmt.Println(textError)
		return "", errors.New("can't parse items")
	}

	return text, nil
}
