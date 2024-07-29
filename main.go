package main

import (
	"bufio"
	"context"
	"fmt"
	"github.com/gocolly/colly"
	"github.com/xuri/excelize/v2"
	"log"
	"os"
	"reflect"
	"runtime"
	"strconv"
	"strings"
	"time"
	"unicode"

	"github.com/PuerkitoBio/goquery"
	"github.com/chromedp/cdproto/dom"
	"github.com/chromedp/chromedp"
)

type Product struct {
	title  string
	vendor string
	price  int
	link   string
}

func (p *Product) SetPrice(price int) {
	p.price = price
}

func (p *Product) SetVendor(vendor string) {
	p.vendor = vendor
}

func (p *Product) SetTitle(title string) {
	p.title = title
}

func (p *Product) SetLink(link string) {
	p.link = link
}

func main() {
	ParseZZap()

	f := excelize.NewFile()
	defer func() {
		if err := f.Close(); err != nil {
			fmt.Println(err)
		}
	}()
	execFunc := []func() []Product{YandexMarket, Ozon, VseInstrumenty, Yanimag, OilStore, Wildberries, Megamarket}

	for _, products := range execFunc {
		funcName := strings.Split(runtime.FuncForPC(reflect.ValueOf(products).Pointer()).Name(), ".")[1]
		if f.GetSheetName(0) == "Sheet1" {
			err := f.SetSheetName("Sheet1", funcName)
			if err != nil {
				log.Println(err)
			}
			for j, product := range products() {
				_ = f.SetColWidth(funcName, "A", "A", 75)
				_ = f.SetColWidth(funcName, "B", "B", 20)
				_ = f.SetColWidth(funcName, "C", "C", 50)
				_ = f.SetColWidth(funcName, "D", "D", 120)
				horiz, _ := f.NewStyle(&excelize.Style{
					Alignment: &excelize.Alignment{
						Horizontal: "center",
					},
				})
				_ = f.SetCellStyle(funcName, "A1", "D50", horiz)
				_ = f.SetSheetRow(funcName, "A1", &[]interface{}{"Название продукта", "Цена", "Продавец", "Ссылка на страницу продукта"})
				currentRow := "A" + strconv.Itoa(j+2)
				err = f.SetSheetRow(funcName, currentRow, &[]interface{}{product.title, product.price, product.vendor, product.link})
			}
		} else {
			_, err := f.NewSheet(funcName)
			if err != nil {
				log.Println(err)
			}
			for j, product := range products() {
				_ = f.SetColWidth(funcName, "A", "A", 75)
				_ = f.SetColWidth(funcName, "B", "B", 20)
				_ = f.SetColWidth(funcName, "C", "C", 50)
				_ = f.SetColWidth(funcName, "D", "D", 120)
				horiz, _ := f.NewStyle(&excelize.Style{
					Alignment: &excelize.Alignment{
						Horizontal: "center",
					},
				})
				_ = f.SetCellStyle(funcName, "A1", "D50", horiz)
				_ = f.SetSheetRow(funcName, "A1", &[]interface{}{"Название продукта", "Цена", "Продавец", "Ссылка на страницу продукта"})
				currentRow := "A" + strconv.Itoa(j+2)
				err = f.SetSheetRow(funcName, currentRow, &[]interface{}{product.title, product.price, product.vendor, product.link})
			}
		}

	}

	fileName := "RESULT.xlsx"
	fmt.Println("\nСоздается файл в текущей папке: " + fileName)
	if err := f.SaveAs(fileName); err != nil {
		log.Println(err)
		log.Println("Закройте файл " + fileName)
		os.Exit(1)
	}
	fmt.Println("Finish! Press ENTER")
	bufio.NewReader(os.Stdin).ReadString('\n')

	// ZZap() !!!
}

func VseInstrumenty() []Product {
	fmt.Println("\nЧтение www.vseinstrumenti.ru\n")
	opts := []chromedp.ExecAllocatorOption{
		chromedp.ExecPath(`C:\Program Files\Google\Chrome\Application\chrome.exe`),
		chromedp.UserAgent("Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/126.0.0.0 Safari/537.36"),
		chromedp.WindowSize(1920, 1080),
		chromedp.NoFirstRun,
		chromedp.NoDefaultBrowserCheck,
		chromedp.Headless,
		chromedp.DisableGPU,
	}
	products := []Product{}
	product := Product{}
	ctx, cancel := chromedp.NewExecAllocator(context.Background(), opts...)
	defer cancel()

	ctx, cancel = chromedp.NewContext(ctx)
	defer cancel()
	var html string
	err := chromedp.Run(ctx,
		// visit the target page
		chromedp.Navigate("https://www.vseinstrumenti.ru/brand/lubrigard--2102564/?asc=asc&orderby=price"),
		// wait for the page to load
		chromedp.Sleep(1000*time.Millisecond),
		// extract the raw HTML from the page
		chromedp.ActionFunc(func(ctx context.Context) error {
			// select the root node on the page
			rootNode, err := dom.GetDocument().Do(ctx)
			if err != nil {
				return err
			}
			html, err = dom.GetOuterHTML().WithNodeID(rootNode.NodeID).Do(ctx)
			return err
		}),
	)
	if err != nil {
		log.Println("Error while performing the automation logic:", err)
	}

	htmlReader := strings.NewReader(html)
	doc, err := goquery.NewDocumentFromReader(htmlReader)
	if err != nil {
		fmt.Println(err)
	}

	doc.Find("[data-qa=\"products-tile\"]").Each(func(i int, selection *goquery.Selection) {
		// Title
		product.SetTitle(selection.Find("a[title]").Text())
		//Price
		result := ""
		for _, v := range selection.Find("[data-qa=\"product-price-current\"]").Text() {
			if string(v) == "." {
				break
			}
			if unicode.IsNumber(v) {
				result += string(v)
			}
		}
		conv, err := strconv.Atoi(result)
		if err != nil {
			log.Println(err)
		}
		product.SetPrice(conv)
		// Vendor
		product.SetVendor("www.vseinstrumenti.ru")
		// Link
		link, _ := selection.Find("[data-qa=\"product-name\"]").Attr("href")
		product.SetLink(link)
		fmt.Println(fmt.Sprintf("Прочитан продукт: 1) %s 2) %s 3) %d руб.", product.title, product.vendor, product.price))
		products = append(products, product)
	})
	return products
}

func Autodoc() []Product {
	fmt.Println("\nЧтение www.autodoc.ru\n")
	opts := []chromedp.ExecAllocatorOption{
		chromedp.ExecPath(`C:\Program Files\Google\Chrome\Application\chrome.exe`),
		chromedp.UserAgent("Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/126.0.0.0 Safari/537.36"),
		chromedp.WindowSize(1920, 1080),
		chromedp.NoFirstRun,
		chromedp.NoDefaultBrowserCheck,
		chromedp.Headless,
		chromedp.DisableGPU,
	}
	isSuccess := false
	products := []Product{}
	product := Product{}
	ctx, cancel := chromedp.NewExecAllocator(context.Background(), opts...)
	defer cancel()

	ctx, cancel = chromedp.NewContext(ctx)
	defer cancel()
	var html string
	err := chromedp.Run(ctx,
		// visit the target page
		chromedp.Navigate("https://www.autodoc.ru/catalogs/universal/maslo-motornoe-780?searchString=LUBRIGARD&page=goods&order=2"),
		chromedp.WaitReady(".cell"),
		chromedp.OuterHTML(".product-items", &html),
	)
	if err != nil {
		log.Println("Error while performing the automation logic:", err)
	}

	htmlReader := strings.NewReader(html)
	doc, err := goquery.NewDocumentFromReader(htmlReader)
	if err != nil {
		fmt.Println(err)
	}
	doc.Find(".cell").Each(func(i int, selection *goquery.Selection) {
		isSuccess = true
		//Title
		product.SetTitle(selection.Find(".properties-text").Text())
		//Price
		result := ""
		for _, v := range selection.Find(".price-container").Find("span").Text() {
			if string(v) == "." {
				break
			}
			if unicode.IsNumber(v) {
				result += string(v)
			}
		}
		conv, err := strconv.Atoi(result)
		if err != nil {
			log.Println(err)
		}
		product.SetPrice(conv)
		// Vendor
		product.SetVendor("autodoc.ru")
		// Link
		link, _ := selection.Find(".product-properties").Attr("href")
		product.SetLink("https://www.autodoc.ru" + link)
		fmt.Println(fmt.Sprintf("Прочитан продукт: 1) %s 2) %s 3) %d руб.", product.title, product.vendor, product.price))
		products = append(products, product)
	})
	if !isSuccess {
		warn := "Слишком много запросов к сайту, нас заблокировали. Попробуйте через 5 минут."
		fmt.Println(warn)
		product.SetTitle(warn)
	}
	return products
}

func Megamarket() []Product {
	fmt.Println("\nЧтение megamarket.ru\n")
	opts := []chromedp.ExecAllocatorOption{
		chromedp.ExecPath(`C:\Program Files\Google\Chrome\Application\chrome.exe`),
		chromedp.UserAgent("Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/126.0.0.0 Safari/537.36"),
		chromedp.WindowSize(1920, 1080),
		chromedp.NoFirstRun,
		chromedp.NoDefaultBrowserCheck,
		chromedp.Headless,
		chromedp.DisableGPU,
	}
	products := []Product{}
	product := Product{}
	ctx, cancel := chromedp.NewExecAllocator(context.Background(), opts...)
	defer cancel()

	ctx, cancel = chromedp.NewContext(ctx)
	defer cancel()
	var html string
	err := chromedp.Run(ctx,
		// visit the target page
		chromedp.Navigate("https://megamarket.ru/catalog/?q=lubrigard#?sort=1"),
		// wait for the page to load
		chromedp.Sleep(1000*time.Millisecond),
		// extract the raw HTML from the page
		chromedp.ActionFunc(func(ctx context.Context) error {
			// select the root node on the page
			rootNode, err := dom.GetDocument().Do(ctx)
			if err != nil {
				return err
			}
			html, err = dom.GetOuterHTML().WithNodeID(rootNode.NodeID).Do(ctx)
			return err
		}),
	)
	if err != nil {
		log.Println("Error while performing the automation logic:", err)
	}

	htmlReader := strings.NewReader(html)
	doc, err := goquery.NewDocumentFromReader(htmlReader)
	if err != nil {
		fmt.Println(err)
	}

	doc.Find(".catalog-item-desktop").Each(func(i int, selection *goquery.Selection) {
		if i < 25 {
			// Title
			title := selection.Find(".catalog-item-regular-desktop__title-link").Text()
			product.SetTitle(strings.TrimSpace(title))
			// Price
			result := ""
			for _, v := range selection.Find(".catalog-item-regular-desktop__price").Text() {
				if unicode.IsNumber(v) {
					result += string(v)
				} else if string(v) == "." {
					break
				}
			}
			conv, _ := strconv.Atoi(result)
			product.SetPrice(conv)
			// Vendor
			vendor := selection.Find(".merchant-info__name").Text()
			product.SetVendor(strings.TrimSpace(vendor))
			// Link
			link, _ := selection.Find(".ddl_product_link").Attr("href")
			product.SetLink("https://megamarket.ru" + link)
			fmt.Println(fmt.Sprintf("Прочитан продукт: 1) %s 2) %s 3) %d руб.", product.title, product.vendor, product.price))
			products = append(products, product)
		}
	})
	return products
}

func Yanimag() []Product {
	fmt.Println("\nЧтение yunimag.ru\n")
	c := colly.NewCollector(colly.UserAgent("Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/58.0.3029.110 Safari/537.36"))
	products := []Product{}
	product := Product{}
	c.OnHTML("#catalog", func(e *colly.HTMLElement) {
		e.ForEach("tr", func(i int, element *colly.HTMLElement) {
			// Price
			if element.ChildText(".price") == "" {
				return
			}
			result := ""
			for _, v := range element.ChildText(".price") {
				if unicode.IsNumber(v) {
					result += string(v)
				} else if string(v) == "." {
					break
				}
			}
			conv, _ := strconv.Atoi(result)
			product.SetPrice(conv)
			// Title
			url := element.ChildAttr("a", "href")
			err := c.Visit("https://yunimag.ru" + url)
			if err != nil {
				log.Println(err)
			}
			// Vendor
			product.SetVendor("yunimag.ru")
			// Link
			product.SetLink("https://yunimag.ru" + url)
			fmt.Println(fmt.Sprintf("Прочитан продукт: 1) %s 2) %s 3) %d руб.", product.title, product.vendor, product.price))
			products = append(products, product)
		})
	})

	c.OnHTML("h1", func(element *colly.HTMLElement) {
		product.SetTitle(strings.TrimSpace(element.Text))
	})

	err := c.Visit("https://yunimag.ru/search/?q=lubrigard&brand=&price")
	if err != nil {
		log.Println(err)
	}
	return products
}

func OilStore() []Product {
	fmt.Println("\nЧтение oil-store.ru\n")
	c := colly.NewCollector(colly.UserAgent("Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/58.0.3029.110 Safari/537.36"))
	products := []Product{}
	product := Product{}
	c.OnHTML("#res-products", func(e *colly.HTMLElement) {
		e.ForEach(".product-layout", func(i int, element *colly.HTMLElement) {
			// Price
			result := ""
			for _, v := range element.ChildText(".common-price") {
				if unicode.IsNumber(v) {
					result += string(v)
				}
			}
			conv, _ := strconv.Atoi(result)

			product.SetPrice(conv)
			// Title
			product.SetTitle(element.ChildText("h4 a"))
			// Vendor
			product.SetVendor("oil-store.ru")
			// Link
			link := element.ChildAttr("h4 a", "href")
			product.SetLink(link)
			fmt.Println(fmt.Sprintf("Прочитан продукт: 1) %s 2) %s 3) %d руб.", product.title, product.vendor, product.price))
			products = append(products, product)
		})
	})

	err := c.Visit("https://oil-store.ru/search/?sort=p.price&order=ASC&search=lubrigard")
	if err != nil {
		log.Println(err)
	}
	return products
}

func Ozon() []Product {
	fmt.Println("\nЧтение ozon.ru\n")
	opts := []chromedp.ExecAllocatorOption{
		chromedp.ExecPath(`C:\Program Files\Google\Chrome\Application\chrome.exe`),
		chromedp.UserAgent("Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/126.0.0.0 Safari/537.36"),
		chromedp.WindowSize(1920, 1080),
		chromedp.NoFirstRun,
		chromedp.NoDefaultBrowserCheck,
		chromedp.Headless,
		chromedp.DisableGPU,
	}
	scrollingScript := `
        // scroll down the page 8 times
        const scrolls = 8
        let scrollCount = 0
        
        // scroll down and then wait for 0.5s
        const scrollInterval = setInterval(() => {
          window.scrollTo(0, document.body.scrollHeight)
          scrollCount++
        
          if (scrollCount === numScrolls) {
           clearInterval(scrollInterval)
          }
        }, 500)
       `
	ctx, cancel := chromedp.NewExecAllocator(context.Background(), opts...)
	defer cancel()

	ctx, cancel = chromedp.NewContext(ctx)
	defer cancel()

	var html string
	products := []Product{}
	product := Product{}
	err := chromedp.Run(ctx,
		// visit the target page
		chromedp.Navigate("https://www.ozon.ru/search/?brand=100262122&deny_category_prediction=true&from_global=true&sorting=price&text=lubrigard"),
		// wait for the page to load
		chromedp.Sleep(2*time.Second),
		chromedp.Click(".rb", chromedp.ByQuery),
		chromedp.WaitVisible(`.tile-hover-target`),
		chromedp.Evaluate(scrollingScript, nil),
		chromedp.Sleep(3000*time.Millisecond),
		// extract the raw HTML from the page
		chromedp.OuterHTML(`#paginatorContent`, &html),
	)
	if err != nil {
		log.Println("Error while performing the automation logic:", err)
	}

	htmlReader := strings.NewReader(html)
	doc, err := goquery.NewDocumentFromReader(htmlReader)
	if err != nil {
		log.Println(err)
	}
	var html1 string
	doc.Find(".tile-root").Each(func(i int, selection *goquery.Selection) {
		//Title
		product.SetTitle(selection.Find(".tsBody500Medium").Text())
		// Link
		link, _ := selection.Find("a.tile-hover-target").Attr("href")
		product.SetLink("https://ozon.ru" + link)
		// Price
		result := ""
		for _, v := range selection.Find(".tsHeadline500Medium").Text() {
			if unicode.IsNumber(v) {
				result += string(v)
			}
		}
		conv, err := strconv.Atoi(result)
		product.SetPrice(conv)
		// Vendor
		err = chromedp.Run(ctx,
			// visit the target page
			chromedp.Navigate("https://ozon.ru"+link),
			// wait for the page to load
			chromedp.WaitVisible("[data-widget=\"webCurrentSeller\"]"),
			// extract the raw HTML from the page
			chromedp.OuterHTML("[data-widget=\"webCurrentSeller\"]", &html1),
		)
		if err != nil {
			log.Println("Error while performing the automation logic:", err)
		}

		htmlReader1 := strings.NewReader(html1)
		doc1, err := goquery.NewDocumentFromReader(htmlReader1)
		if err != nil {
			log.Println(err)
		}
		product.SetVendor(doc1.Find("a[title]").Text())
		fmt.Println(fmt.Sprintf("Прочитан продукт: 1) %s 2) %s 3) %d руб.", product.title, product.vendor, product.price))
		products = append(products, product)
	})
	return products
}

func Wildberries() []Product {
	fmt.Println("\nЧтение www.wildberries.ru\n")
	opts := []chromedp.ExecAllocatorOption{
		chromedp.ExecPath(`C:\Program Files\Google\Chrome\Application\chrome.exe`),
		chromedp.UserAgent("Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/126.0.0.0 Safari/537.36"),
		chromedp.WindowSize(1920, 1080),
		chromedp.NoFirstRun,
		chromedp.NoDefaultBrowserCheck,
		chromedp.Headless,
		chromedp.DisableGPU,
	}

	ctx, cancel := chromedp.NewExecAllocator(context.Background(), opts...)
	defer cancel()

	ctx, cancel = chromedp.NewContext(ctx)
	defer cancel()

	var html string
	products := []Product{}
	product := Product{}
	err := chromedp.Run(ctx,
		// visit the target page
		chromedp.Navigate("https://www.wildberries.ru/brands/310529574-lubrigard/masla-motornye?sort=priceup&page=1"),
		// wait for the page to load
		chromedp.Sleep(2000*time.Millisecond),
		// extract the raw HTML from the page
		chromedp.ActionFunc(func(ctx context.Context) error {
			// select the root node on the page
			rootNode, err := dom.GetDocument().Do(ctx)
			if err != nil {
				log.Println(err)
			}
			html, err = dom.GetOuterHTML().WithNodeID(rootNode.NodeID).Do(ctx)
			return err
		}),
	)
	if err != nil {
		log.Println("Error while performing the automation logic:", err)
	}

	htmlReader := strings.NewReader(html)
	doc, err := goquery.NewDocumentFromReader(htmlReader)
	if err != nil {
		log.Println(err)
	}

	doc.Find("article").Each(func(i int, selection *goquery.Selection) {
		// Title
		title := selection.Find(".product-card__name").Text()
		title = strings.TrimSpace(title)
		title = strings.Replace(title, "/ ", "", 1)
		product.SetTitle(title)
		// Price
		result := ""
		for _, v := range selection.Find(".wallet-price").Text() {
			if unicode.IsNumber(v) {
				result += string(v)
			}
		}
		conv, err := strconv.Atoi(result)
		if err != nil {
			log.Println(err)
		}
		product.SetPrice(conv)
		// Vendor
		url, _ := selection.Find(".product-card__link").Attr("href")
		err = chromedp.Run(ctx,
			// visit the target page
			chromedp.Navigate(url),
			// wait for the page to load
			chromedp.Sleep(1000*time.Millisecond),
			// extract the raw HTML from the page
			chromedp.ActionFunc(func(ctx context.Context) error {
				// select the root node on the page
				rootNode, err := dom.GetDocument().Do(ctx)
				if err != nil {
					return err
				}
				html, err = dom.GetOuterHTML().WithNodeID(rootNode.NodeID).Do(ctx)
				return err
			}),
		)
		if err != nil {
			log.Println("Error while performing the automation logic:", err)
		}

		htmlReader = strings.NewReader(html)
		doc, err = goquery.NewDocumentFromReader(htmlReader)
		if err != nil {
			log.Println(err)
		}
		vendor := strings.TrimSpace(doc.Find(".seller-info__name").Nodes[0].FirstChild.Data)
		product.SetVendor(vendor)
		// Link
		link, _ := selection.Find(".product-card__link").Attr("href")
		product.SetLink(link)
		fmt.Println(fmt.Sprintf("Прочитан продукт: 1) %s 2) %s 3) %d руб.", product.title, product.vendor, product.price))
		products = append(products, product)

	})
	return products
}

func YandexMarket() []Product {
	fmt.Println("\nЧитаем market.yandex.ru\n")
	opts := []chromedp.ExecAllocatorOption{
		chromedp.ExecPath(`C:\Program Files\Google\Chrome\Application\chrome.exe`),
		chromedp.UserAgent("Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/126.0.0.0 Safari/537.36"),
		chromedp.WindowSize(1920, 1080),
		chromedp.NoFirstRun,
		chromedp.NoDefaultBrowserCheck,
		chromedp.Headless,
		chromedp.DisableGPU,
	}
	products := []Product{}
	product := Product{}
	ctx, cancel := chromedp.NewExecAllocator(context.Background(), opts...)
	defer cancel()

	ctx, cancel = chromedp.NewContext(ctx)
	defer cancel()
	var html string
	err := chromedp.Run(ctx,
		// visit the target page
		chromedp.Navigate("https://market.yandex.ru/catalog--tovary-dlia-avto-i-mototekhniki/54418/list?text=lubrigard&hid=90402&how=aprice&rs=eJxFzMFKwzAAxvFmdpBVhBzHBC07FTw0SVOzePQieBJfIGRZZiNdV5oWYQdR2GngUTwKewKfYQffQ2968BkcVfHy57v8Pno2PAl6eTOu7JWqJuj1_W1n6EMP9dp6P428a28BU0wp4yS5A-tNdwU6EFyAhQ-_Nt0X4J0-g4AEEAIE-iAEqDM41KWSLpuX0ha6qeWNrTOpy5l0tSm0zcPtbUBbAvt7oY_QYH-snNVSq2reOJPLMle2kM6oSmfh-uM2-nxYgYC1ZvfXHGjlGpX_o0Ta2szcH3ta6ujx_giBc6gEYwQLfkkJp0SMUkZHnLIY8-lIG5YIYlLKJ8l2aS2mglJi8DHGMYnJNxVRTok%2C&glfilter=7893318%3A50224713"),
		// wait for the page to load
		chromedp.Sleep(1000*time.Millisecond),
		// extract the raw HTML from the page
		chromedp.ActionFunc(func(ctx context.Context) error {
			// select the root node on the page
			rootNode, err := dom.GetDocument().Do(ctx)
			if err != nil {
				return err
			}
			html, err = dom.GetOuterHTML().WithNodeID(rootNode.NodeID).Do(ctx)
			return err
		}),
	)
	if err != nil {
		log.Println("Error while performing the automation logic:", err)
	}

	htmlReader := strings.NewReader(html)
	doc, err := goquery.NewDocumentFromReader(htmlReader)
	if err != nil {
		fmt.Println(err)
	}
	doc.Find("[data-baobab-name=\"productSnippet\"]").Each(func(i int, selection *goquery.Selection) {
		if i < 7 {
			return
		}
		if i < 19 {
			// Title
			product.SetTitle(selection.Find("[data-auto=\"snippet-title\"]").Text())
			// Price
			result := ""
			for _, v := range selection.Find("[data-auto=\"snippet-price-current\"]").Text() {
				if unicode.IsNumber(v) {
					result += string(v)
				}
			}
			conv, err := strconv.Atoi(result)
			if err != nil {
				log.Println(err)
			}
			product.SetPrice(conv)
			// Vendor
			link, _ := selection.Find("[data-auto=\"snippet-link\"]").Attr("href")
			linkFormat := fmt.Sprintf("https://market.yandex.ru%s", link)
			var html string
			err = chromedp.Run(ctx,
				// visit the target page
				chromedp.Navigate(linkFormat),
				chromedp.WaitReady("[data-baobab-name=\"productSnippet\"]"),
				// extract the raw HTML from the page
				chromedp.ActionFunc(func(ctx context.Context) error {
					// select the root node on the page
					rootNode, err := dom.GetDocument().Do(ctx)
					if err != nil {
						return err
					}
					html, err = dom.GetOuterHTML().WithNodeID(rootNode.NodeID).Do(ctx)
					return err
				}),
			)
			if err != nil {
				log.Println("Error while performing the automation logic:", err)
			}

			htmlReader := strings.NewReader(html)
			doc, err := goquery.NewDocumentFromReader(htmlReader)
			if err != nil {
				fmt.Println(err)
			}
			vendor := doc.Find("[data-auto=\"shop-info-container\"]").Find("[data-zone-name=\"shop-name\"]").Find("a").Text()
			product.SetVendor(vendor)
			// Link
			product.SetLink(linkFormat)
			fmt.Println(fmt.Sprintf("Прочитан продукт: 1) %s 2) %s 3) %d руб.", product.title, product.vendor, product.price))
			products = append(products, product)
		}
	})
	return products
}

func ParseZZap() {
	opts := []chromedp.ExecAllocatorOption{
		chromedp.ExecPath(`C:\Program Files\Google\Chrome\Application\chrome.exe`),
		chromedp.UserAgent("Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:91.0) Gecko/20100101 Firefox/91.0"),
		chromedp.WindowSize(1920, 1080),
		chromedp.NoFirstRun,
		chromedp.NoDefaultBrowserCheck,
		//chromedp.Headless,
		chromedp.DisableGPU,
	}

	ctx, cancel := chromedp.NewExecAllocator(context.Background(), opts...)
	defer cancel()

	ctx, cancel = chromedp.NewContext(ctx)
	defer cancel()
	var html string
	product := Product{}
	err := chromedp.Run(ctx,
		// visit the target page
		chromedp.Navigate("https://www.zzap.ru/public/catalogs/catalogoils.aspx#tag_v=(%D0%BB%D1%8E%D0%B1%D0%BE%D0%B9)&tag_w=(%D0%BB%D1%8E%D0%B1%D0%BE%D0%B9)&tag_s=(%D0%BB%D1%8E%D0%B1%D0%BE%D0%B9)&class_man=LUBRIGARD&price_min=null&price_max=null&code_cur=1"),
		// wait for the page to load
		chromedp.Sleep(4000*time.Millisecond),
		// extract the raw HTML from the page
		chromedp.ActionFunc(func(ctx context.Context) error {
			// select the root node on the page
			rootNode, err := dom.GetDocument().Do(ctx)
			if err != nil {
				return err
			}
			html, err = dom.GetOuterHTML().WithNodeID(rootNode.NodeID).Do(ctx)
			return err
		}),
	)
	if err != nil {
		log.Println("Error while performing the automation logic:", err)
	}

	htmlReader := strings.NewReader(html)
	doc, err := goquery.NewDocumentFromReader(htmlReader)
	if err != nil {
		fmt.Println(err)
	}

	doc.Find(".dxgvDataRow_ZZapAqua").Each(func(i int, selection *goquery.Selection) {
		if i < 4 {
			var html1 string
			url, _ := selection.Find("a.f14b").Attr("href")
			test := selection.Find(".f12b").Text()
			fmt.Println(test)
			err = chromedp.Run(ctx,
				// visit the target page
				chromedp.Navigate("https://www.zzap.ru"+url),
				// wait for the page to load
				chromedp.Sleep(4000*time.Millisecond),
				// extract the raw HTML from the page
				chromedp.ActionFunc(func(ctx context.Context) error {
					// select the root node on the page
					rootNode, err := dom.GetDocument().Do(ctx)
					if err != nil {
						return err
					}
					html1, err = dom.GetOuterHTML().WithNodeID(rootNode.NodeID).Do(ctx)
					return err
				}),
			)
			if err != nil {
				log.Println("Error while performing the automation logic:", err)
			}

			htmlReader1 := strings.NewReader(html1)
			doc1, err := goquery.NewDocumentFromReader(htmlReader1)
			if err != nil {
				log.Println(err)
			}
			doc1.Find(".dxgvDataRow_ZZapAqua").Each(func(i int, selection *goquery.Selection) {
				if i < 4 {
					// Price
					result := ""
					for _, v := range selection.Find(".right").Find("span.f14b").Text() {
						if unicode.IsNumber(v) {
							result += string(v)
						}
					}
					if len(result) < 1 {
						return
					}
					conv, err := strconv.Atoi(result)
					if err != nil {
						log.Println(err)
					}
					product.SetPrice(conv)
					// Vendor
					vendor := selection.Find(".dxgv").Find("a.f14b").Text()
					product.SetVendor(vendor)
					//product.GetAll()
				}

			})
		}
	})
}

//func Exist() {
//	fmt.Println("\nЧтение www.exist.ru\n")
//	c := colly.NewCollector()
//	c.OnRequest(func(request *colly.Request) {
//		c.Cookies("o=,2,")
//		// Manually set cookies
//		cookies := []*http.Cookie{
//			{
//				Name:   "o",
//				Value:  ",2,",
//				Domain: "www.exist.ru",
//			},
//			{
//				Name:   "_ref",
//				Value:  "https://www.exist.ru/Catalog/Goods/7/3/F511116F",
//				Domain: "www.exist.ru",
//			},
//		}
//		//request.Headers.Set("o", ",2,")
//		// Or set the cookies directly on the collector
//		err := c.SetCookies("www.exist.ru", cookies)
//		if err != nil {
//			log.Println(err)
//		}
//	})
//	c.OnResponse(func(r *colly.Response) {
//		// Access the cookies
//		cookies := c.Cookies(r.Request.URL.String())
//		for _, cookie := range cookies {
//			log.Println("Cookie:", cookie.Name, "Value:", cookie.Value)
//		}
//	})
//	c.OnHTML("#unicat", func(element *colly.HTMLElement) {
//		element.ForEach(".cell2", func(i int, element *colly.HTMLElement) {
//			test, test1, test3 := element.ChildText(".wrap p"),
//				element.ChildText(".descr"),
//				element.ChildText(".ucatprc")
//			fmt.Println(test, test1, test3)
//		})
//	})
//	//c.OnHTML(".cell2", func(element *colly.HTMLElement) {
//	//	test, test1, test3 := element.ChildText(".wrap p"),
//	//		element.ChildText(".descr"),
//	//		element.ChildText(".ucatprc")
//	//	fmt.Println(test, test1, test3)
//	//})
//
//	err := c.Visit("https://www.exist.ru/Catalog/Goods/7/3?1_-10=6651")
//	if err != nil {
//		log.Println(err)
//	}
//}
