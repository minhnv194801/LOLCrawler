package crawler

import (
	"context"
	"encoding/csv"
	"log"
	"os"
	"strings"

	"lolcrawl/model"

	"github.com/PuerkitoBio/goquery"
	"github.com/chromedp/cdproto/dom"
	"github.com/chromedp/chromedp"
)

type LPLCrawler struct {
	Tasks      chromedp.Action
	Context    context.Context
	CancelFunc context.CancelFunc
}

const (
	url = "https://gol.gg/tournament/tournament-matchlist/LPL%20Spring%202023/"
)

func (crawler *LPLCrawler) Start() error {
	return chromedp.Run(crawler.Context, crawler.Tasks)
}

func (crawler *LPLCrawler) Cancel() {
	crawler.CancelFunc()
}

func NewLPLCrawler() *LPLCrawler {
	crawler := new(LPLCrawler)
	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.Flag("headless", true),
		chromedp.Flag("start-fullscreen", false),
	)

	allocCtx, _ := chromedp.NewExecAllocator(context.Background(), opts...)
	crawler.Context, crawler.CancelFunc = chromedp.NewContext(allocCtx, chromedp.WithLogf(log.Printf))

	crawler.Tasks = chromedp.Tasks{
		chromedp.Navigate(url),
		chromedp.ActionFunc(crawlAllMatchInfo),
	}

	return crawler
}

func crawlAllMatchInfo(ctx context.Context) error {
	matchUrls, err := crawlMatchUrls(ctx)
	if err != nil {
		return err
	}
	var csvData [][]string
	for _, matchUrl := range matchUrls {
		match, err := crawlMatchInfo(ctx, matchUrl)
		if err != nil {
			log.Println(err)
			continue
		}
		csvData = append(csvData, match.ToCsvData()...)
		csvData = append(csvData, []string{})
	}

	f, err := os.Create("lplstat.csv")
	if err != nil {
		log.Fatal("Failed to create file", err)
	}
	defer f.Close()

	if err != nil {
		log.Fatalln("failed to open file", err)
	}

	w := csv.NewWriter(f)
	err = w.WriteAll(csvData)

	if err != nil {
		log.Fatal(err)
	}

	return nil
}

func crawlMatchUrls(ctx context.Context) ([]string, error) {
	node, err := dom.GetDocument().Do(ctx)
	if err != nil {
		log.Fatal(err)
		return nil, err
	}
	res, err := dom.GetOuterHTML().WithNodeID(node.NodeID).Do(ctx)
	if err != nil {
		log.Fatal(err)
		return nil, err
	}
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(res))
	if err != nil {
		log.Fatal(err)
		return nil, err
	}

	matchUrls := make([]string, 0)
	doc.Find("body > div > main > div:nth-child(7) > div > div:nth-child(5) > div > section > div > div > table > tbody > tr > td > a").Each(func(index int, info *goquery.Selection) {
		url, _ := info.Attr("href")
		if strings.Contains(url, "page-summary") {
			url = strings.Replace(url, "..", "https://gol.gg", -1)
			matchUrls = append(matchUrls, url)
		}
	})

	return matchUrls, nil
}

func crawlMatchInfo(ctx context.Context, matchUrl string) (model.Match, error) {
	var match model.Match
	match.Team = make([]string, 2)
	chromedp.Navigate(matchUrl).Do(ctx)

	node, err := dom.GetDocument().Do(ctx)
	if err != nil {
		log.Fatal(err)
		return model.Match{}, err
	}
	res, err := dom.GetOuterHTML().WithNodeID(node.NodeID).Do(ctx)
	if err != nil {
		log.Fatal(err)
		return model.Match{}, err
	}
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(res))
	if err != nil {
		log.Fatal(err)
		return model.Match{}, err
	}

	match.Team[0] = doc.Find("body > div > main > div:nth-child(4) > div > div.row.rowbreak.fond-main-cadre.p-4 > div > div.col-cadre.pb-4 > div.row.pb-3 > div:nth-child(1) > h1 > a").First().Text()
	match.Team[1] = doc.Find("body > div > main > div:nth-child(4) > div > div.row.rowbreak.fond-main-cadre.p-4 > div > div.col-cadre.pb-4 > div.row.pb-3 > div:nth-child(3) > h1 > a").First().Text()
	var gameUrls []string
	doc.Find("a[href]").Each(func(index int, info *goquery.Selection) {
		url, _ := info.Attr("href")
		if strings.Contains(url, "page-game") {
			url = strings.Replace(url, "..", "https://gol.gg", -1)
			gameUrls = append(gameUrls, url)
		}
	})

	for _, gameUrl := range gameUrls {
		game, err := crawlGameInfo(ctx, gameUrl)
		if err != nil {
			log.Println(err)
			continue
		}
		match.Game = append(match.Game, game)
	}

	return match, nil
}

func crawlGameInfo(ctx context.Context, gameUrl string) (model.Game, error) {
	var game model.Game
	game.Team = make([]string, 2)
	game.Players = make([]model.Player, 10)
	chromedp.Navigate(gameUrl).Do(ctx)
	node, err := dom.GetDocument().Do(ctx)
	if err != nil {
		log.Fatal(err)
		return model.Game{}, err
	}
	res, err := dom.GetOuterHTML().WithNodeID(node.NodeID).Do(ctx)
	if err != nil {
		log.Fatal(err)
		return model.Game{}, err
	}
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(res))
	if err != nil {
		log.Fatal(err)
		return model.Game{}, err
	}
	game.Time = doc.Find("body > div > main > div:nth-child(4) > div > div:nth-child(4) > div > div > div > div:nth-child(1) > div > div > div:nth-child(1) > div.col-6.text-center > h1").First().Text()
	game.Team[0] = doc.Find("body > div > main > div:nth-child(4) > div > div:nth-child(4) > div > div > div > div:nth-child(1) > div > div > div:nth-child(2) > div:nth-child(1) > div.row.rowbreak.pb-3 > div").First().Text()
	game.Team[1] = doc.Find("body > div > main > div:nth-child(4) > div > div:nth-child(4) > div > div > div > div:nth-child(1) > div > div > div:nth-child(2) > div:nth-child(2) > div.row.rowbreak.pb-3 > div").First().Text()
	game.Team[0] = strings.Trim(game.Team[0], "\n")
	game.Team[0] = strings.Trim(game.Team[0], " ")
	game.Team[1] = strings.Trim(game.Team[1], "\n")
	game.Team[1] = strings.Trim(game.Team[1], " ")

	url := strings.Replace(gameUrl, "page-game", "page-fullstats", -1)
	chromedp.Navigate(url).Do(ctx)
	node, err = dom.GetDocument().Do(ctx)
	if err != nil {
		log.Fatal(err)
		return model.Game{}, err
	}
	res, err = dom.GetOuterHTML().WithNodeID(node.NodeID).Do(ctx)
	if err != nil {
		log.Fatal(err)
		return model.Game{}, err
	}
	doc, err = goquery.NewDocumentFromReader(strings.NewReader(res))
	if err != nil {
		log.Fatal(err)
		return model.Game{}, err
	}

	doc.Find("table > thead > tr > th > img").Each(func(index int, info *goquery.Selection) {
		champ, _ := info.Attr("alt")
		if strings.Compare(champ, "K") == 0 {
			champ = "KSante"
		}
		game.Players[index].Champion = champ
	})

	doc.Find("table > tbody > tr:nth-child(2) > td").Each(func(index int, info *goquery.Selection) {
		if index == 0 {
			return
		}
		role := info.Text()
		game.Players[index-1].Role = role
	})

	doc.Find("table > tbody > tr:nth-child(4) > td").Each(func(index int, info *goquery.Selection) {
		if index == 0 {
			return
		}
		kills := info.Text()
		game.Players[index-1].Kills = kills
	})

	doc.Find("table > tbody > tr:nth-child(5) > td").Each(func(index int, info *goquery.Selection) {
		if index == 0 {
			return
		}
		deaths := info.Text()
		game.Players[index-1].Deaths = deaths
	})

	doc.Find("table > tbody > tr:nth-child(6) > td").Each(func(index int, info *goquery.Selection) {
		if index == 0 {
			return
		}
		assists := info.Text()
		game.Players[index-1].Assists = assists
	})

	doc.Find("table > tbody > tr:nth-child(7) > td").Each(func(index int, info *goquery.Selection) {
		if index == 0 {
			return
		}
		kda := info.Text()
		game.Players[index-1].KDA = kda
	})

	doc.Find("table > tbody > tr:nth-child(8) > td").Each(func(index int, info *goquery.Selection) {
		if index == 0 {
			return
		}
		cs := info.Text()
		game.Players[index-1].CS = cs
	})

	doc.Find("table > tbody > tr:nth-child(11) > td").Each(func(index int, info *goquery.Selection) {
		if index == 0 {
			return
		}
		golds := info.Text()
		game.Players[index-1].Golds = golds
	})

	doc.Find("table > tbody > tr:nth-child(25) > td").Each(func(index int, info *goquery.Selection) {
		if index == 0 {
			return
		}
		totalDamage := info.Text()
		game.Players[index-1].TotalDamage = totalDamage
	})

	doc.Find("table > tbody > tr:nth-child(51) > td").Each(func(index int, info *goquery.Selection) {
		if index == 0 {
			return
		}
		totalDamageTaken := info.Text()
		game.Players[index-1].TotalDamageTaken = totalDamageTaken
	})

	return game, nil
}
