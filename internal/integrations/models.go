package integrations

type CatBreedWeight struct {
	Imperial string `json:"imperial"`
	Metric   string `json:"metric"`
}

type CatBreed struct {
	ID             string         `json:"id"`
	Name           string         `json:"name"`
	Origin         string         `json:"origin"`
	Temperament    string         `json:"temperament"`
	Description    string         `json:"description"`
	LifeSpan       string         `json:"life_span"`
	WikipediaURL   string         `json:"wikipedia_url"`
	ReferenceImage string         `json:"reference_image_id"`
	Weight         CatBreedWeight `json:"weight"`
	EnergyLevel    int            `json:"energy_level"`
	AffectionLevel int            `json:"affection_level"`
	ChildFriendly  int            `json:"child_friendly"`
	DogFriendly    int            `json:"dog_friendly"`
	HealthIssues   int            `json:"health_issues"`
	Grooming       int            `json:"grooming"`
	Intelligence   int            `json:"intelligence"`
	Hypoallergenic int            `json:"hypoallergenic"`
}

type CatImage struct {
	ID     string     `json:"id"`
	URL    string     `json:"url"`
	Width  int        `json:"width"`
	Height int        `json:"height"`
	Breeds []CatBreed `json:"breeds,omitempty"`
}

type CatFact struct {
	Fact   string `json:"fact"`
	Length int    `json:"length"`
}

type CareTip struct {
	ID       string `json:"id"`
	Category string `json:"category"`
	Emoji    string `json:"emoji"`
	Title    string `json:"title"`
	Text     string `json:"text"`
	Source   string `json:"source"`
}

type CataasImage struct {
	ID       string   `json:"id"`
	Tags     []string `json:"tags"`
	URL      string   `json:"url"`
	MIMEType string   `json:"mimetype"`
}

type FoodProduct struct {
	Barcode         string         `json:"barcode"`
	Name            string         `json:"name"`
	GenericName     string         `json:"generic_name"`
	Brands          string         `json:"brands"`
	Quantity        string         `json:"quantity"`
	Categories      string         `json:"categories"`
	IngredientsText string         `json:"ingredients_text"`
	ImageURL        string         `json:"image_url"`
	NutriScore      string         `json:"nutriscore_grade"`
	NovaGroup       string         `json:"nova_group"`
	Nutriments      map[string]any `json:"nutriments"`
	SourceURL       string         `json:"source_url"`
}

type AdoptionAnimal struct {
	ID          int      `json:"id"`
	Name        string   `json:"name"`
	Age         string   `json:"age"`
	Gender      string   `json:"gender"`
	Size        string   `json:"size"`
	Status      string   `json:"status"`
	Description string   `json:"description"`
	URL         string   `json:"url"`
	Breeds      []string `json:"breeds"`
	Photos      []string `json:"photos"`
	Contact     string   `json:"contact"`
}
