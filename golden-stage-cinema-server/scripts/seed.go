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
		{"title": "Dune: Part Two", "genre": "Sci-Fi/Adventure", "duration_mins": 166, "poster_url": "https://example.com/dune2.jpg", "synopsis": "Paul Atreides unites with Chani and the Fremen while on a warpath of revenge."},
		{"title": "Deadpool & Wolverine", "genre": "Action/Comedy", "duration_mins": 127, "poster_url": "https://example.com/deadpool3.jpg", "synopsis": "Wolverine is recovering from his injuries when he crosses paths with the loudmouth, Deadpool."},
		{"title": "Inside Out 2", "genre": "Animation/Family", "duration_mins": 96, "poster_url": "https://example.com/insideout2.jpg", "synopsis": "Follows Riley, in her teenage years, encountering new emotions."},
		{"title": "Godzilla x Kong: The New Empire", "genre": "Action/Sci-Fi", "duration_mins": 115, "poster_url": "https://example.com/gxk.jpg", "synopsis": "Two ancient titans, Godzilla and Kong, clash in an epic battle."},
		{"title": "How to Make Millions Before Grandma Dies (หลานม่า)", "genre": "Drama", "duration_mins": 125, "poster_url": "https://example.com/lanma.jpg", "synopsis": "A young man quits his job to care for his dying grandmother, motivated by her fortune."},
		{"title": "Furiosa: A Mad Max Saga", "genre": "Action/Adventure", "duration_mins": 148, "poster_url": "https://example.com/furiosa.jpg", "synopsis": "The origin story of renegade warrior Furiosa."},
		{"title": "Kingdom of the Planet of the Apes", "genre": "Sci-Fi/Action", "duration_mins": 145, "poster_url": "https://example.com/apes.jpg", "synopsis": "Many years after the reign of Caesar, a young ape goes on a journey."},
		{"title": "Despicable Me 4", "genre": "Animation/Comedy", "duration_mins": 95, "poster_url": "https://example.com/dm4.jpg", "synopsis": "Gru, Lucy, Margo, Edith, and Agnes welcome a new member to the family."},
		{"title": "Oppenheimer", "genre": "Drama/Biography", "duration_mins": 180, "poster_url": "https://example.com/oppenheimer.jpg", "synopsis": "The story of American scientist J. Robert Oppenheimer and his role in the development of the atomic bomb."},
		{"title": "Spider-Man: Across the Spider-Verse", "genre": "Animation/Action", "duration_mins": 140, "poster_url": "https://example.com/spiderverse.jpg", "synopsis": "Miles Morales catapults across the Multiverse."},
		{"title": "The Batman", "genre": "Action/Crime", "duration_mins": 176, "poster_url": "https://example.com/batman.jpg", "synopsis": "When a sadistic serial killer begins murdering key political figures in Gotham, Batman is forced to investigate."},
		{"title": "Avatar: The Way of Water", "genre": "Sci-Fi/Action", "duration_mins": 192, "poster_url": "https://example.com/avatar2.jpg", "synopsis": "Jake Sully lives with his newfound family formed on the extrasolar moon Pandora."},
	}

	movieIDs := make([]primitive.ObjectID, 0, len(moviesData))
	moviesDocs := make([]interface{}, 0, len(moviesData))
	for _, m := range moviesData {
		id := primitive.NewObjectID()
		movieIDs = append(movieIDs, id)
		m["_id"] = id
		moviesDocs = append(moviesDocs, m)
	}
	db.Collection("movies").InsertMany(ctx, moviesDocs)
	log.Printf("Inserted %d movies.\n", len(moviesDocs))

	log.Println("--- 3. สร้างข้อมูล Cinemas (2 สาขา) ---")
	cinemasData := []string{"Golden Stage Siam", "Golden Stage Ladprao"}
	var cinemaIDs []primitive.ObjectID

	for _, name := range cinemasData {
		cID := primitive.NewObjectID()
		cinemaIDs = append(cinemaIDs, cID)
		db.Collection("cinemas").InsertOne(ctx, bson.M{
			"_id":  cID,
			"name": name,
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
	now := time.Now().Truncate(24 * time.Hour) // เริ่มที่เที่ยงคืนวันนี้
	times := []int{10, 12, 14, 16, 18, 20, 22} // ชั่วโมงฉาย

	var showtimeIDs []primitive.ObjectID
	showtimesDocs := make([]interface{}, 0)

	// ลูป 7 วัน
	for day := 0; day < 7; day++ {
		currentDate := now.AddDate(0, 0, day)

		// ลูป 8 โรง
		for i, hID := range hallIDs {
			cID := cinemaIDs[i/4] // 4 โรงแรกของสาขา 1, 4 โรงหลังของสาขา 2

			// ลูป 7 รอบต่อวัน
			for _, hour := range times {
				sID := primitive.NewObjectID()
				showtimeIDs = append(showtimeIDs, sID)
				movieID := movieIDs[rand.Intn(len(movieIDs))] // สุ่มหนัง

				startTime := currentDate.Add(time.Duration(hour) * time.Hour)
				endTime := startTime.Add(2 * time.Hour)

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
