package movies

import "go.mongodb.org/mongo-driver/bson/primitive"

// Movie คือ Struct ที่จับคู่กับ Document ใน Collection 'movies'
type Movie struct {
	ID           primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Title        string             `bson:"title" json:"title"`
	Genre        string             `bson:"genre" json:"genre"`
	DurationMins int                `bson:"duration_mins" json:"duration_mins"`
	PosterURL    string             `bson:"poster_url" json:"poster_url"`
	BackdropURL  string             `bson:"backdrop_url" json:"backdrop_url"`
	Synopsis     string             `bson:"synopsis" json:"synopsis"`
	Rating       float64            `bson:"rating" json:"rating"`
}
