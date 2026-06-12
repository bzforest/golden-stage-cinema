package models

// Movie คือ Struct ที่จับคู่กับ Document ใน Collection 'movies'
type Movie struct {
	ID           string `bson:"_id,omitempty" json:"id"`
	Title        string `bson:"title" json:"title"`
	Genre        string `bson:"genre" json:"genre"`
	DurationMins int    `bson:"duration_mins" json:"durationMins"`
	PosterURL    string `bson:"poster_url" json:"posterURL"`
	Synopsis     string `bson:"synopsis" json:"synopsis"`
}
