package main

import (
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/PuerkitoBio/goquery"
)

type game struct {
	league   string
	status   string
	homeTeam string
	awayTeam string
	link     string
	score    string
	hour     string
	home00   float64
}

var selectedGames []game

func main() {
	url := "https://www.academiadasapostasbrasil.com/"

	resp, err := http.Get(url)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		panic(fmt.Sprintf("Erro: %d", resp.StatusCode))
	}

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		log.Fatal(err)
	}

	//Finding the games table...
	doc.Find("table.competition-today.dskt tbody").Each(func(i int, item *goquery.Selection) {
		listGames(item)
	})

	filteredGamesBefore20 := filterGamesBefore20(selectedGames)

	filteredGamesByMainRule := filterGamesByMainRule(filteredGamesBefore20)

	for _, v := range filteredGamesByMainRule {
		fmt.Println(v.status, v.hour, v.homeTeam, v.score, v.awayTeam, v.home00)
	}

}

func filterGamesByMainRule(games []game) []game {
	var filteredGames []game

	for _, v := range games {
		resp, err := http.Get(v.link)
		if err != nil {
			log.Fatal(err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			panic(fmt.Sprintf("Erro: %d", resp.StatusCode))
		}

		doc, err := goquery.NewDocumentFromReader(resp.Body)
		if err != nil {
			log.Fatal(err)
		}

		doc.Find("table.ajax_load_stats").Each(func(i int, item *goquery.Selection) {
			tableName := item.Children().First().Children().First().Children().First().Children().First().Children().First().Text()
			tableName = strings.TrimSpace(tableName)

			if strings.TrimSpace(tableName) == "Resultado" {
				item.Find(".ajax-container td.mobile_single_column").Each(func(i int, tableScores *goquery.Selection) {
					homeTeamScores := strings.TrimSpace(tableScores.Text())
					fmt.Println(homeTeamScores)
				})
			}

			// if item.Children().First().Is("span") {
			// 	teamTableStats := item.Children().First().Text()
			// 	if teamTableStats == v.homeTeam {

			// 	}
			// }
			// teamTableStats := strings.TrimSpace(item.Children().First().Text())
			// fmt.Println(teamTableStats)

		})

		break

	}
	return filteredGames
}

func filterGamesBefore20(games []game) []game {
	var filteredGames []game
	for _, v := range games {
		if v.status == "" {
			continue
		}
		if v.status != "Não Iniciado" {
			minutesPlayed, erro := strconv.Atoi(v.status)
			if erro != nil {
				fmt.Println(erro)
			}
			if minutesPlayed > 20 {
				continue
			}
		}
		filteredGames = append(filteredGames, v)
	}
	return filteredGames
}

func listGames(t *goquery.Selection) {
	t.Find("tr").Each(func(i int, row *goquery.Selection) {
		writeRow(row)
	})
}

func writeRow(row *goquery.Selection) {
	currentGame := game{}

	row.Find("td").Each(func(i int, item *goquery.Selection) {

		attr, _ := item.Attr("class")
		attr = strings.TrimSpace(attr)

		if attr == "flag tipsy-active" {
			league, _ := item.Children().First().Attr("title")
			currentGame.league = league
		}

		if attr == "status" {
			statusText := strings.TrimSpace(item.Children().First().Text())
			if statusText == "" {
				statusText = "Não Iniciado"
				currentGame.status = statusText
			}
		}

		if attr == "status gameinlive" {
			statusText := strings.TrimSpace(item.Children().First().Text())
			if statusText != "Meio Tempo" {
				statusText = strings.TrimSpace(item.Children().First().Text())
			}
			currentGame.status = statusText
		}

		if strings.Contains(attr, "team-a") {
			homeTeam := item.Children().First().Text()
			currentGame.homeTeam = homeTeam
		}

		if attr == "score" {
			currentGame.score = " vs "
			link, _ := item.Children().First().Attr("href")
			currentGame.link = link
		}

		if attr == "score gameinlive" {
			link, _ := item.Children().First().Attr("href")
			homeScore := item.Children().First().Children().First().Text()
			awayScore := item.Children().First().Children().Last().Text()
			currentGame.link = link
			currentGame.score = strings.TrimSpace(homeScore) + " - " + strings.TrimSpace(awayScore)
		}

		if strings.Contains(attr, "team-b") {
			awayTeam := item.Children().First().Text()
			currentGame.awayTeam = awayTeam
		}

		if attr == "hour" {
			hourStr := strings.TrimSpace(item.Children().First().Text())
			hour, _ := strconv.ParseInt(hourStr, 10, 64)
			t := time.Unix(hour, 0)

			currentGame.hour = strings.TrimSpace(t.UTC().Format("15:04"))
		}
	})

	selectedGames = append(selectedGames, currentGame)
}

func FixLength(s, c string, n int) string {
	// Contar caracteres como runas.
	lenStr := utf8.RuneCountInString(s)

	// Cortar se for maior que o desejado.
	if lenStr > n {
		runes := []rune(s)
		return string(runes[:n])
	}

	// Preencher com espaços se for menor.
	if lenStr < n {
		return s + strings.Repeat(c, n-lenStr)
	}

	// Retornar a string como está se já tem o tamanho exato.
	return s
}
