package model

type Product struct {
	ID          int     `json:"id"`
	Title       string  `json:"title"`
	Price       float64 `json:"price"`
	Description string  `json:"description"`
}
type PhotoSize struct {
	URL  string `json:"url"`
	Type string `json:"type"`
}

type Photo struct {
	Sizes []PhotoSize `json:"sizes"`
}

type Variant struct {
	ID         int     `json:"id"`
	Name       string  `json:"name"`
	Price      float64 `json:"price"`
	Properties []struct {
		Name  string `json:"name"`
		Value string `json:"value"`
	} `json:"properties"`
}

type Property struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

type VariantGrid struct {
	Name     string `json:"name"`
	Type     string `json:"type"`
	Variants []struct {
		VariantID  int    `json:"variant_id"`
		Name       string `json:"name"`
		ItemID     int    `json:"item_id"`
		IsSelected bool   `json:"is_selected"`
	} `json:"variants"`
}

type Item struct {
	ID    int    `json:"id"`
	Title string `json:"title"`
	Price struct {
		Amount string `json:"amount"`
	} `json:"price"`
	Description  string        `json:"description"`
	VariantsGrid []VariantGrid `json:"variants_grid"`
}

type Balloon struct {
	ID             int     `db:"id"`
	Name           string  `db:"name"`
	BalloonPrice   int     `db:"balloon_price"`
	HeliumPortions float64 `db:"helium_portions"`
	HiFloat        int     `db:"hi_float"`
	Sku            int     `db:"sku"`
}

func (Balloon) TableName() string {
	return "balloon"
}
