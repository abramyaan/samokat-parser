package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"samokat-parser/internal/scraper"
	"time"
)

func main() {
	s := &scraper.Scraper{}
	
	category := "molochnoe-i-yaytsa" // Категория
	
	fmt.Println("=== Запуск парсера Самокат ===")
	products, err := s.FetchCategoryWithBrowser(category)
	if err != nil {
		log.Fatal(err)
	}


	outputDir := "results"
	_ = os.MkdirAll(outputDir, os.ModePerm)

	
	fileName := fmt.Sprintf("%s_%s.json", category, time.Now().Format("2006-01-02"))
	filePath := filepath.Join(outputDir, fileName)

	fileData, _ := json.MarshalIndent(products, "", "  ")
	err = os.WriteFile(filePath, fileData, 0644)
	if err != nil {
		fmt.Printf("❌ Ошибка сохранения: %v\n", err)
	} else {
		fmt.Printf("\n💾 Данные сохранены в: %s\n", filePath)
	}

	// Вывод в консоль для проверки
	for i, p := range products {
		if i >= 10 { break }
		fmt.Printf("[%d] %s — %d коп. (Parsed: %s)\n", i+1, p.Name, p.Prices.Current, p.ParsedAt)
	}
}