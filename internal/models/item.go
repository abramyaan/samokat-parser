package models

type Product struct {
	UUID  string `json:"uuid"`
	Name  string `json:"name"`
	Slug  string `json:"slug"`
	ParsedAt string `json:"parsed_at"` // Новое поле
	Prices struct {
		Current int64 `json:"current"` // Цена в копейках (9900)
	} `json:"prices"`
}

// Ответ для списка товаров в категории (обычно это массив или объект)
type CategoryResponse struct {
	Products []Product `json:"products"`
}