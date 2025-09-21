package models

// Recommendations represents track recommendations
type Recommendations struct {
	Seeds  []RecommendationSeed `json:"seeds"`
	Tracks []Track              `json:"tracks"`
}

// RecommendationSeed represents a seed used for recommendations
type RecommendationSeed struct {
	AfterFilteringSize int    `json:"afterFilteringSize"`
	AfterRelinkingSize int    `json:"afterRelinkingSize"`
	Href               string `json:"href"`
	ID                 string `json:"id"`
	InitialPoolSize    int    `json:"initialPoolSize"`
	Type               string `json:"type"`
}

// AvailableGenreSeeds represents available genre seeds
type AvailableGenreSeeds struct {
	Genres []string `json:"genres"`
}