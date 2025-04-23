package main

import (
	"cmp"
	"context"
	"encoding/xml"
	"fmt"
	"log"
	"net/http"
	"slices"
	"strconv"
	"strings"
	"time"

	"golang.org/x/net/html/charset"
)

type ValCurs struct {
	XMLName xml.Name `xml:"ValCurs"`
	Date    string   `xml:"Date,attr"`
	Name    string   `xml:"name,attr"`
	Valute  []Valute `xml:"Valute"`
}

type Valute struct {
	ID        string `xml:"ID,attr"`
	NumCode   string `xml:"NumCode"`
	CharCode  string `xml:"CharCode"`
	Nominal   string `xml:"Nominal"`
	Name      string `xml:"Name"`
	Value     string `xml:"Value"`
	VunitRate string `xml:"VunitRate"`
}

type Val struct {
	Date  string
	Name  string
	Value float64
}

func main() {
	fmt.Println("Получение данных ЦБ по курсам валют за 90 дней...")
	// получаем курсы валют за последние 90 дней
	valCurses, err := getValCurses()
	if err != nil {
		fmt.Println(err)
	}

	vals, err := getVals(valCurses)
	if err != nil {
		fmt.Println(err)
	}

	slices.SortFunc(vals, func(a, b *Val) int {
		return cmp.Compare(a.Value, b.Value)
	})

	fmt.Println("Максимальное значение:")
	fmt.Printf("Валюта: %s\nЗначение %.2f\nДата %s\n\n", vals[len(vals)-1].Name, vals[len(vals)-1].Value, vals[len(vals)-1].Date)

	fmt.Println("Минимальное значение:")
	fmt.Printf("Валюта: %s\nЗначение %.2f\nДата %s\n\n", vals[0].Name, vals[0].Value, vals[0].Date)

	var sum float64
	for _, v := range vals {
		sum += v.Value
	}
	fmt.Printf("Среднее значение курса рубля: %.2f", sum/float64(len(vals)))

	time.Sleep(20 * time.Second)
}

func getValCurses() ([]*ValCurs, error) {
	client := &http.Client{}
	valCurses := make([]*ValCurs, 0, 90)
	loc, err := time.LoadLocation("Europe/Moscow")
	if err != nil {
		return nil, err
	}

	for i := range 90 {
		date := time.Now().In(loc).AddDate(0, 0, -i).Format("02/01/2006")
		url := "http://www.cbr.ru/scripts/XML_daily.asp?date_req=" + date
		request, err := http.NewRequestWithContext(context.Background(), http.MethodGet, url, nil)

		if err != nil {
			return nil, err
		}

		request.Header.Set("User-Agent", "Mozilla/5.0 (compatible; MyGoApp/1.0)")

		response, err := client.Do(request)
		if err != nil {
			return nil, err
		}
		defer response.Body.Close()

		decoder := xml.NewDecoder(response.Body)
		decoder.CharsetReader = charset.NewReaderLabel

		var valCurs ValCurs
		err = decoder.Decode(&valCurs)
		if err != nil {
			log.Println("Ошибка")
			return nil, err
		}

		valCurses = append(valCurses, &valCurs)
	}

	return valCurses, nil
}

// Получение всех курсов за 90 дней
func getVals(valCurses []*ValCurs) (vals []*Val, err error) {
	for _, valCurs := range valCurses {
		for _, valute := range valCurs.Valute {
			replacedValue := strings.ReplaceAll(valute.Value, ",", ".")
			newValue, err := strconv.ParseFloat(replacedValue, 64)
			if err != nil {
				return nil, err
			}

			val := &Val{
				Date:  valCurs.Date,
				Name:  valute.Name,
				Value: newValue,
			}

			vals = append(vals, val)
		}
	}
	return vals, nil
}
