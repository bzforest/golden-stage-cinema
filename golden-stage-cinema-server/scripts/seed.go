package main

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"os"
	"time"

	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func main() {
	// โหลดไฟล์ .env
	err := godotenv.Load("../.env")
	if err != nil {
		log.Println("Warning: Error loading .env file from ../.env. Using default values.")
	}

	mongoURI := os.Getenv("MONGO_URI")
	if mongoURI == "" {
		mongoURI = "mongodb://127.0.0.1:27017"
	}
	dbName := os.Getenv("MONGO_DB_NAME")
	if dbName == "" {
		dbName = "golden_stage_db"
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute) // ให้เวลา 2 นาทีเผื่อ insert นาน
	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(mongoURI))
	if err != nil {
		log.Fatal(err)
	}
	defer client.Disconnect(ctx)

	db := client.Database(dbName)

	log.Println("--- 1. ล้างข้อมูลเก่า (Dropping old collections) ---")
	db.Collection("movies").Drop(ctx)
	db.Collection("cinemas").Drop(ctx)
	db.Collection("halls").Drop(ctx)
	db.Collection("showtimes").Drop(ctx)
	db.Collection("showtime_seats").Drop(ctx)

	log.Println("--- 2. สร้างข้อมูล Movies (12 เรื่อง) ---")
	moviesData := []map[string]interface{}{
		{
			"_id":           "6a2be959e5986753bf0f48db",
			"title":         "Dune: Part Two",
			"genre":         "Sci-Fi/Adventure",
			"duration_mins": 166,
			"poster_url":    "https://m.media-amazon.com/images/M/MV5BNTc0YmQxMjEtODI5MC00NjFiLTlkMWUtOGQ5NjFmYWUyZGJhXkEyXkFqcGc@._V1_FMjpg_UX1000_.jpg",
			"synopsis":      "Paul Atreides unites with Chani and the Fremen while on a warpath of revenge.",
			"backdrop_url":  "https://images.hdqwalls.com/wallpapers/dune-part-two-poster-5k-pb.jpg",
			"rating":        7.5,
		},
		{
			"_id":           "6a2be959e5986753bf0f48dc",
			"poster_url":    "https://spacebar.th/api/media/file/deadpool_and_wolverine_new_poster_SPACEBAR_Photo_V01_d8afe018ba.jpg",
			"synopsis":      "Wolverine is recovering from his injuries when he crosses paths with the loudmouth, Deadpool.",
			"title":         "Deadpool & Wolverine",
			"genre":         "Action/Comedy",
			"duration_mins": 127,
			"backdrop_url":  "https://images4.alphacoders.com/138/1382563.jpg",
			"rating":        8.5,
		},
		{
			"_id":           "6a2be959e5986753bf0f48dd",
			"title":         "Inside Out 2",
			"genre":         "Animation/Family",
			"duration_mins": 96,
			"poster_url":    "https://lumiere-a.akamaihd.net/v1/images/io2_payoff_squish_poster-s_29270ace.jpeg",
			"synopsis":      "Follows Riley, in her teenage years, encountering new emotions.",
			"backdrop_url":  "https://4kwallpapers.com/images/wallpapers/inside-out-2-3840x2160-17097.jpg",
			"rating":        9.2,
		},
		{
			"_id":           "6a2be959e5986753bf0f48de",
			"title":         "Godzilla x Kong: The New Empire",
			"genre":         "Action/Sci-Fi",
			"duration_mins": 115,
			"poster_url":    "https://m.media-amazon.com/images/M/MV5BMTY0N2MzODctY2ExYy00OWYxLTkyNDItMTVhZGIxZjliZjU5XkEyXkFqcGc@._V1_.jpg",
			"synopsis":      "Two ancient titans, Godzilla and Kong, clash in an epic battle.",
			"backdrop_url":  "https://4kwallpapers.com/images/wallpapers/godzilla-x-kong-the-5120x2880-15314.jpg",
			"rating":        6.5,
		},
		{
			"_id":           "6a2be959e5986753bf0f48df",
			"poster_url":    "https://i0.wp.com/www.patsonic.com/wp-content/uploads/2024/04/gdh-lahn-mah-lao-poster.jpg?ssl=1",
			"synopsis":      "A young man quits his job to care for his dying grandmother, motivated by her fortune.",
			"title":         "How to Make Millions Before Grandma Dies (หลานม่า)",
			"genre":         "Drama",
			"duration_mins": 125,
			"backdrop_url":  "https://images.squarespace-cdn.com/content/v1/556facf5e4b0b7fae0a87634/1739992771052-CAC8UZEMQCLVTLKB8WEM/HowToMakeMillionsBeforeGrandmaDiesAPUC.jpg",
			"rating":        7.7,
		},
		{
			"_id":           "6a2be959e5986753bf0f48e0",
			"title":         "Furiosa: A Mad Max Saga",
			"genre":         "Action/Adventure",
			"duration_mins": 148,
			"poster_url":    "https://cdn.posteritati.com/posters/000/000/072/696/furiosa-a-mad-max-saga-md-web.jpg",
			"synopsis":      "The origin story of renegade warrior Furiosa.",
			"backdrop_url":  "https://4kwallpapers.com/images/wallpapers/furiosa-a-mad-max-7680x4320-16776.jpg",
			"rating":        5.5,
		},
		{
			"_id":           "6a2be959e5986753bf0f48e1",
			"title":         "Kingdom of the Planet of the Apes",
			"genre":         "Sci-Fi/Action",
			"duration_mins": 145,
			"poster_url":    "https://lumiere-a.akamaihd.net/v1/images/th-_kopa_chinaart_poster_th_1sht_tagline-s_0347e5be.jpeg",
			"synopsis":      "Many years after the reign of Caesar, a young ape goes on a journey.",
			"backdrop_url":  "https://images5.alphacoders.com/140/thumb-1920-1404695.jpg",
			"rating":        6.8,
		},
		{
			"_id":           "6a2be959e5986753bf0f48e2",
			"duration_mins": 95,
			"poster_url":    "https://image.tmdb.org/t/p/original/2a2x54KigX2TchLiyfFe6SJHDgb.jpg",
			"synopsis":      "Gru, Lucy, Margo, Edith, and Agnes welcome a new member to the family.",
			"title":         "Despicable Me 4",
			"genre":         "Animation/Comedy",
			"backdrop_url":  "https://4kwallpapers.com/images/wallpapers/despicable-me-4-3440x1440-17255.jpeg",
			"rating":        8.9,
		},
		{
			"_id":           "6a2be959e5986753bf0f48e3",
			"title":         "Oppenheimer",
			"genre":         "Drama/Biography",
			"duration_mins": 180,
			"poster_url":    "https://upload.wikimedia.org/wikipedia/en/4/4a/Oppenheimer_%28film%29.jpg",
			"synopsis":      "The story of American scientist J. Robert Oppenheimer and his role in the development of the atomic bomb.",
			"backdrop_url":  "https://images5.alphacoders.com/125/thumb-1920-1257951.jpeg",
			"rating":        7.8,
		},
		{
			"_id":           "6a2be959e5986753bf0f48e4",
			"title":         "Spider-Man: Across the Spider-Verse",
			"genre":         "Animation/Action",
			"duration_mins": 140,
			"poster_url":    "https://media.themoviedb.org/t/p/w220_and_h330_face/4CwKj1fw33BXYzxvrpM3GlAhK4L.jpg",
			"synopsis":      "Miles Morales catapults across the Multiverse.",
			"backdrop_url":  "https://4kwallpapers.com/images/wallpapers/spider-man-across-3840x2160-10140.jpg",
			"rating":        8.3,
		},
		{
			"_id":           "6a2be959e5986753bf0f48e5",
			"title":         "The Batman",
			"genre":         "Action/Crime",
			"duration_mins": 176,
			"poster_url":    "https://m.media-amazon.com/images/I/91ezOOQjE3L.jpg",
			"synopsis":      "When a sadistic serial killer begins murdering key political figures in Gotham, Batman is forced to investigate.",
			"backdrop_url":  "https://w0.peakpx.com/wallpaper/307/244/HD-wallpaper-batman-the-batman.jpg",
			"rating":        7.6,
		},
		{
			"_id":           "6a2be959e5986753bf0f48e6",
			"duration_mins": 192,
			"poster_url":    "https://lumiere-a.akamaihd.net/v1/images/avatar-wayofwater-en_a56a9e95_1d38e163.jpeg",
			"synopsis":      "Jake Sully lives with his newfound family formed on the extrasolar moon Pandora.",
			"title":         "Avatar: The Way of Water",
			"genre":         "Sci-Fi/Action",
			"backdrop_url":  "https://images.alphacoders.com/129/1292804.jpg",
			"rating":        9.4,
		},
	}

	movieIDs := make([]primitive.ObjectID, 0, len(moviesData))
	moviesDocs := make([]interface{}, 0, len(moviesData))
	movieDurations := make(map[primitive.ObjectID]int)

	for _, m := range moviesData {
		hexID := m["_id"].(string)
		id, _ := primitive.ObjectIDFromHex(hexID)
		m["_id"] = id
		movieIDs = append(movieIDs, id)
		moviesDocs = append(moviesDocs, m)
		
		// Store duration for showtime end_time calculation
		durationVal := m["duration_mins"]
		var duration int
		switch v := durationVal.(type) {
		case int:
			duration = v
		case float64:
			duration = int(v)
		default:
			duration = 120 // Default fallback
		}
		movieDurations[id] = duration
	}
	db.Collection("movies").InsertMany(ctx, moviesDocs)
	log.Printf("Inserted %d movies.\n", len(moviesDocs))

	log.Println("--- 3. สร้างข้อมูล Cinemas (2 สาขา) ---")
	cinemasData := []map[string]string{
		{"name": "Golden Stage Siam", "type": "IMAX"},
		{"name": "Golden Stage Ladprao", "type": "Standard 2D"},
	}
	var cinemaIDs []primitive.ObjectID

	for _, c := range cinemasData {
		cID := primitive.NewObjectID()
		cinemaIDs = append(cinemaIDs, cID)
		db.Collection("cinemas").InsertOne(ctx, bson.M{
			"_id":         cID,
			"name":        c["name"],
			"cinema_type": c["type"],
		})
	}
	log.Printf("Inserted %d cinemas.\n", len(cinemaIDs))

	log.Println("--- 4. สร้างข้อมูล Halls (สาขาละ 4 โรง) ---")
	var hallIDs []primitive.ObjectID
	for _, cID := range cinemaIDs {
		for i := 1; i <= 4; i++ {
			hID := primitive.NewObjectID()
			hallIDs = append(hallIDs, hID)
			db.Collection("halls").InsertOne(ctx, bson.M{
				"_id":       hID,
				"cinema_id": cID,
				"name":      fmt.Sprintf("Hall %d", i),
				"capacity":  80,
			})
		}
	}
	log.Printf("Inserted %d halls.\n", len(hallIDs))

	log.Println("--- 5. สร้างข้อมูล Showtimes (7 วัน) ---")
	loc, _ := time.LoadLocation("Asia/Bangkok")
	now := time.Now().In(loc)
	midnight := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, loc)
	minutesOptions := []int{0, 15, 30, 45}

	var showtimeIDs []primitive.ObjectID
	showtimesDocs := make([]interface{}, 0)

	// ลูป 7 วัน
	for day := 0; day < 7; day++ {
		currentDate := midnight.AddDate(0, 0, day)

		// ลูป 8 โรง
		for i, hID := range hallIDs {
			cID := cinemaIDs[i/4] // 4 โรงแรกของสาขา 1, 4 โรงหลังของสาขา 2

			// สุ่มรอบฉาย 5 รอบต่อวันต่อโรง
			for j := 0; j < 5; j++ {
				sID := primitive.NewObjectID()
				showtimeIDs = append(showtimeIDs, sID)
				movieID := movieIDs[rand.Intn(len(movieIDs))] // สุ่มหนัง

				// สุ่มชั่วโมง (10-22) และนาที (0, 15, 30, 45)
				hour := 10 + rand.Intn(13) // 10 + (0 to 12) = 10 to 22
				minute := minutesOptions[rand.Intn(len(minutesOptions))]

				startTime := time.Date(currentDate.Year(), currentDate.Month(), currentDate.Day(), hour, minute, 0, 0, loc)
				duration := movieDurations[movieID]
				endTime := startTime.Add(time.Duration(duration) * time.Minute)

				showtimesDocs = append(showtimesDocs, bson.M{
					"_id":        sID,
					"movie_id":   movieID,
					"cinema_id":  cID,
					"hall_id":    hID,
					"start_time": startTime,
					"end_time":   endTime,
				})
			}
		}
	}
	db.Collection("showtimes").InsertMany(ctx, showtimesDocs)
	log.Printf("Inserted %d showtimes.\n", len(showtimesDocs))

	log.Println("--- 6. สร้างข้อมูล Seats (รอบฉายละ 80 ที่นั่ง) ---")
	rows := []string{"A", "B", "C", "D", "E", "F", "G", "H"}
	seatsDocs := make([]interface{}, 0)

	for _, sID := range showtimeIDs {
		for _, row := range rows {
			seatType := "Normal"
			price := 150.0
			if row == "A" || row == "B" {
				seatType = "Premium"
				price = 300.0
			}

			for num := 1; num <= 10; num++ {
				seatsDocs = append(seatsDocs, bson.M{
					"_id":         primitive.NewObjectID(),
					"showtime_id": sID,
					"seat_number": fmt.Sprintf("%s%d", row, num),
					"type":        seatType,
					"status":      "AVAILABLE",
					"price":       price,
				})
			}
		}

		// Insert ทุกๆ 5,000 เรคคอร์ดเพื่อไม่ให้เปลือง Memory เกินไป
		if len(seatsDocs) >= 5000 {
			db.Collection("showtime_seats").InsertMany(ctx, seatsDocs)
			seatsDocs = seatsDocs[:0] // Clear slice
		}
	}

	// Insert ส่วนที่เหลือ
	if len(seatsDocs) > 0 {
		db.Collection("showtime_seats").InsertMany(ctx, seatsDocs)
	}
	log.Printf("Inserted %d showtime_seats.\n", len(showtimeIDs)*80)

	log.Println("--- Seeding Complete! ---")
}
