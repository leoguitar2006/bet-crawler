package main

import (
	"fmt"
	"io"
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
	away00   float64
	homeft00 float64
	awayft00 float64
}

var selectedGames []game

func main() {
	url := "https://www.academiadasapostasbrasil.com/"
	page := getUrl(url)

	doc, err := goquery.NewDocumentFromReader(page)
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
		fmt.Println(v.league, v.status, v.hour, v.homeTeam, v.score, v.awayTeam, v.home00, v.away00, v.homeft00, v.awayft00)
	}

	page.Close()
}

func getUrl(url string) io.ReadCloser {
	resp, err := http.Get(url)
	if err != nil {
		log.Fatal(err)
	}

	if resp.StatusCode != http.StatusOK {
		panic(fmt.Sprintf("Erro: %d", resp.StatusCode))
	}
	return resp.Body
}

func filterGamesByMainRule(games []game) []game {
	var filteredGames []game
	var gameAccepted bool

	for _, v := range games {
		statsLink := getUrl(v.link)

		doc, err := goquery.NewDocumentFromReader(statsLink)
		if err != nil {
			log.Fatal(err)
		}

		gameAccepted = false
		doc.Find("table.ajax_load_stats").Each(func(i int, item *goquery.Selection) {
			tableName := item.Children().First().Children().First().Children().First().Children().First().Children().First().Text()
			tableName = strings.TrimSpace(tableName)

			if strings.TrimSpace(tableName) == "Resultado" {
				item.Find(".ajax-container td.mobile_single_column").Each(func(i int, tableScores *goquery.Selection) {
					teamName := getTeamName(tableScores.Children().First())
					if strings.Contains(teamName, v.homeTeam) || strings.Contains(teamName, v.awayTeam) {
						tableScores.Find("table table.stat-correctscore").Each(func(i int, halfOrFull *goquery.Selection) {
							gamePart := halfOrFull.Children().First().Children().First().Children().First().Text()
							if strings.Contains(gamePart, "intervalo") || strings.Contains(gamePart, "final") {
								halfOrFull.Find("tbody.stat-quarts-padding").Each(func(i int, table *goquery.Selection) {
									table.Find("tr").Each(func(i int, row *goquery.Selection) {
										gameScore := strings.TrimSpace(row.Children().First().Text())
										if gameScore == "0-0" {
											percent00 := strings.TrimSpace(row.Children().Last().Text())
											percentPosition := strings.Index(percent00, "%")
											percent00 = percent00[:percentPosition]

											if teamName == v.homeTeam {
												if strings.Contains(gamePart, "intervalo") {
													v.home00, _ = strconv.ParseFloat(percent00, 64)
												} else {
													v.homeft00, _ = strconv.ParseFloat(percent00, 64)
												}

											} else if teamName == v.awayTeam {
												if strings.Contains(gamePart, "intervalo") {
													v.away00, _ = strconv.ParseFloat(percent00, 64)
												} else {
													v.awayft00, _ = strconv.ParseFloat(percent00, 64)
												}
											}
											gameAccepted = true
										}
									})
								})
							}
						})
					}
				})
			}
		})
		if gameAccepted {
			filteredGames = append(filteredGames, v)
		}
		statsLink.Close()
	}
	return filteredGames
}

func getTeamName(spans *goquery.Selection) string {
	teamName, _ := spans.Html()
	spacePosition := strings.Index(teamName, "<")
	teamName = teamName[:spacePosition]
	teamName = strings.TrimSpace(teamName)
	return teamName
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

		if attr == "score" {
			currentGame.score = " vs "

			link, _ := item.Children().First().Attr("href")
			currentGame.link = link
			currentGame.homeTeam, currentGame.awayTeam = fillTeams(link)
		}

		if attr == "score gameinlive" {
			homeScore := item.Children().First().Children().First().Text()
			awayScore := item.Children().First().Children().Last().Text()
			currentGame.score = strings.TrimSpace(homeScore) + " - " + strings.TrimSpace(awayScore)

			link, _ := item.Children().First().Attr("href")
			currentGame.link = link
			currentGame.homeTeam, currentGame.awayTeam = fillTeams(link)
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

func fillTeams(link string) (homeTeam, awayTeam string) {
	statsPage := getUrl(link)
	defer statsPage.Close()

	doc, err := goquery.NewDocumentFromReader(statsPage)
	if err != nil {
		log.Fatal(err)
	}

	nameHomeTeam, _ := doc.Find(".mobile_single_column table thead tr th.first").Html()
	nameAwayTeam, _ := doc.Find(".mobile_single_column table thead tr th.last").Html()

	return strings.TrimSpace(nameHomeTeam), strings.TrimSpace(nameAwayTeam)
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
