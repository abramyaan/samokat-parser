package scraper

import (
	"context"
	"fmt"
	"math/rand"
	"strconv"
	"strings"
	"time"

	"github.com/chromedp/chromedp"
	"samokat-parser/internal/models"
)

func (s *Scraper) FetchCategoryWithBrowser(categorySlug string) ([]models.Product, error) {
	
	browserPath := `C:\Program Files (x86)\Microsoft\Edge\Application\msedge.exe`

	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.ExecPath(browserPath),
		chromedp.Flag("headless", false), 
		chromedp.Flag("disable-web-security", true),
		chromedp.Flag("disable-blink-features", "AutomationControlled"),
		chromedp.Flag("excludeSwitches", "enable-automation"),
		chromedp.UserAgent("Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/121.0.0.0 Safari/537.36 Edg/121.0.0.0"),
	)

	allocCtx, cancel := chromedp.NewExecAllocator(context.Background(), opts...)
	defer cancel()
	ctx, cancel := chromedp.NewContext(allocCtx)
	defer cancel()

	ctx, cancel = context.WithTimeout(ctx, 240*time.Second)
	defer cancel()

	var rawData []struct {
		Name  string
		Price string
	}

	targetURL := fmt.Sprintf("https://samokat.ru/category/%s", categorySlug)

	// Хелпер для случайных пауз
	humanSleep := func(min, max int) chromedp.Action {
		return chromedp.Sleep(time.Duration(rand.Intn(max-min)+min) * time.Millisecond)
	}

	fmt.Println("🤖 Запуск финальной версии (адаптация под новую разметку)...")

	err := chromedp.Run(ctx,
		chromedp.Navigate(targetURL),
		humanSleep(5000, 7000),

		// 1. СБРОС АДРЕСА
		chromedp.Evaluate(`
			(function() {
				const btn = Array.from(document.querySelectorAll('button')).find(el => el.innerText.includes('Нет') || el.innerText.includes('другой'));
				if (btn) btn.click();
			})()
		`, nil),
		humanSleep(3000, 4000),

		// 2. ВВОД ГОРОДА
		chromedp.ActionFunc(func(ctx context.Context) error {
			fmt.Println("🏙️ Шаг 1: Ввожу город...")
			return nil
		}),
		chromedp.WaitVisible(`[class*="AddressSuggest_root"] input`, chromedp.ByQuery),
		chromedp.SendKeys(`[class*="AddressSuggest_root"] input`, "Оренбург"),
		humanSleep(4000, 5000),

		chromedp.Evaluate(`
			(function() {
				const items = Array.from(document.querySelectorAll('[class*="Suggest_suggestItem"]'));
				const target = items.find(el => el.innerText.trim() === 'Оренбург');
				if (target) {
					target.scrollIntoView();
					target.click();
					return true;
				}
				return false;
			})()
		`, nil),
		humanSleep(3000, 4000),

		// 3. ВВОД УЛИЦЫ
		chromedp.ActionFunc(func(ctx context.Context) error {
			fmt.Println("🏠 Шаг 2: Ввожу улицу...")
			return nil
		}),
		chromedp.WaitVisible(`[class*="addressSuggest"] input`, chromedp.ByQuery),
		chromedp.SendKeys(`[class*="addressSuggest"] input`, "Карагандинская улица, 22"),
		humanSleep(6000, 8000),

		// ВЫБОР ТОЧНОГО АДРЕСА
		chromedp.Evaluate(`
			(function() {
				const items = Array.from(document.querySelectorAll('[class*="Suggest_suggestItem"]'));
				const target = items.find(el => {
					const text = el.innerText.replace(/\s+/g, ' ').trim();
					return text.includes('Карагандинская') && text.includes('22') && !text.includes('22/') && !text.includes('22а');
				}) || items[0];

				if (target) {
					target.scrollIntoView();
					['mousedown', 'mouseup', 'click'].forEach(type => {
						target.dispatchEvent(new MouseEvent(type, { bubbles: true, cancelable: true, view: window }));
					});
					return true;
				}
				return false;
			})()
		`, nil),
		humanSleep(4000, 5000),

		// 4. ПОДТВЕРЖДЕНИЕ
		chromedp.ActionFunc(func(ctx context.Context) error {
			fmt.Println("✅ Шаг 3: Нажимаю 'Да, всё верно'...")
			return nil
		}),
		chromedp.Evaluate(`
			(function() {
				const btns = Array.from(document.querySelectorAll('button'));
				const confirm = btns.find(el => 
					el.innerText.includes('всё верно') || 
					el.innerText.includes('Доставить')
				);
				if (confirm) {
					confirm.scrollIntoView();
					confirm.click();
				}
			})()
		`, nil),
		humanSleep(8000, 10000),

		// 5. СБОР ТОВАРОВ
		chromedp.ActionFunc(func(ctx context.Context) error {
			fmt.Println("📜 Мы на странице. Собираю товары из списка...")
			return nil
		}),
		chromedp.WaitVisible(`[class*="ProductsList_productList"]`, chromedp.ByQuery),
		// Скроллим вниз несколько раз, чтобы подгрузить всё (lazy load)
		chromedp.Evaluate(`window.scrollBy(0, 2000);`, nil),
		humanSleep(2000, 3000),
		chromedp.Evaluate(`window.scrollBy(0, 2000);`, nil),
		humanSleep(2000, 3000),

		chromedp.Evaluate(`
			(function() {
				const cards = document.querySelectorAll('a[href^="/product/"]');
				return Array.from(cards).map(card => {
					const img = card.querySelector('img');
					const priceContainer = card.querySelector('[class*="ProductCardActions_text"]');
					
					let priceStr = "0";
					if (priceContainer) {
						const oldPriceEl = priceContainer.querySelector('[class*="oldPrice"]');
						let text = priceContainer.innerText;
						if (oldPriceEl) {
							text = text.replace(oldPriceEl.innerText, "").trim();
						}
						priceStr = text.replace(/[^\d]/g, ""); 
					}

					return {
						Name: img ? img.alt : "Наименование не найдено",
						Price: priceStr
					};
				});
			})()
		`, &rawData),
	)

	if err != nil {
		return nil, fmt.Errorf("ошибка парсинга: %v", err)
	}

	// ФИНАЛЬНАЯ ОБРАБОТКА ДАННЫХ
	var finalProducts []models.Product
	uniqueNames := make(map[string]bool)
	parsedAt := time.Now().Format("2006-01-02 15:04:05")

	// Список слов-исключений для очистки от мусора
	junkKeywords := []string{"стакан", "кружка", "лопатка", "салатник", "тарелка", "кисть", "набор", "таймер"}

	for _, item := range rawData {
		// 1. Проверка на пустые данные
		if item.Name == "Наименование не найдено" || item.Price == "" || item.Price == "0" {
			continue
		}

		// 2. Убираем дубликаты (проверка по имени)
		if _, exists := uniqueNames[item.Name]; exists {
			continue
		}

		// 3. Фильтр мусора (непищевые товары)
		isJunk := false
		lowerName := strings.ToLower(item.Name)
		for _, word := range junkKeywords {
			if strings.Contains(lowerName, word) {
				isJunk = true
				break
			}
		}
		if isJunk {
			continue
		}

		// 4. Конвертация цены
		priceVal, _ := strconv.ParseInt(item.Price, 10, 64)

		// 5. Создание чистого объекта
		p := models.Product{
			Name:     item.Name,
			Slug:     item.Price + " ₽",
			ParsedAt: parsedAt,
		}
		p.Prices.Current = priceVal * 100 // В копейки для БД/аналитики

		finalProducts = append(finalProducts, p)
		uniqueNames[item.Name] = true
	}

	fmt.Printf("🎉 Готово! Успешно собрано %d уникальных товаров (отфильтровано).\n", len(finalProducts))
	return finalProducts, nil
}