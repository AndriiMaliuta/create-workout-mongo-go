package gcp_mongo_sless

import (
	"context"
	"encoding/json"
	"fmt"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/GoogleCloudPlatform/functions-framework-go/functions"
)

type Workout struct {
	Id           string    `json:"_id" bson:"_id,omitempty"`
	Record       int64     `json:"record"`
	Sets         int       `json:"sets"`
	Comments     string    `json:"comments"`
	CreationDate time.Time `json:"creation_date" bson:"creation_date"`
	WorkoutDate  string    `json:"workout_date" bson:"workout_date"`
	Day          string    `json:"day"`
	Week         int       `json:"week"`
	WorkoutType  string    `json:"workout_type" bson:"workout_type"`
	Month        string    `json:"month"`
	Year         int       `json:"year"`
}

func init() {
	functions.HTTP("createWorkout", createWorkout)
}

func createWorkout(wr http.ResponseWriter, req *http.Request) {
	var workout Workout
	if err := json.NewDecoder(req.Body).Decode(&workout); err != nil {
		fmt.Fprint(wr, err)
		log.Panicln(err)
	}
	uri := os.Getenv("MONGODB_URI")
	client, err := mongo.Connect(context.TODO(), options.Client().ApplyURI(uri))
	if err != nil {
		panic(err)
	}
	defer func() {
		if err := client.Disconnect(context.TODO()); err != nil {
			panic(err)
		}
	}()
	coll := client.Database("workouts").Collection("workouts")
	if req.Method == "GET" {
		fmt.Fprint(wr, "TEST")
	} else if req.Method == "PUT" {
		wDate := req.URL.Query().Get("workout_date")
		wType := req.URL.Query().Get("workout_type")
		// filter
		var existingWrkt Workout
		filter := bson.D{{"workout_date", wDate}, {"workout_type", wType}}
		coll.FindOne(context.TODO(), filter).Decode(&existingWrkt)
		// data
		var sets, week int
		var comments, workDate, workType, day, month string
		if workout.WorkoutDate != "" {
			workDate = workout.WorkoutDate
		} else {
			workDate = existingWrkt.WorkoutDate
		}
		if workout.Sets != 0 {
			sets = workout.Sets
		} else {
			sets = existingWrkt.Sets
		}
		if workout.Comments != "" {
			comments = workout.Comments
		} else {
			comments = existingWrkt.Comments
		}
		if workout.WorkoutType != "" {
			workType = workout.WorkoutType
		} else {
			workType = existingWrkt.WorkoutType
		}
		if workout.Day != "" {
			day = workout.Day
		} else {
			day = existingWrkt.Day
		}
		if workout.Month != "" {
			month = workout.Month
		} else {
			month = existingWrkt.Month
		}
		if workout.Week != 0 {
			week = workout.Week
		} else {
			week = existingWrkt.Week
		}
		updatedWorkut := bson.D{{"$set", bson.D{
			{"workout_date", workDate},
			{"sets", sets}, {"comments", comments}, {"workout_type", workType},
			{"day", day}, {"month", month}, {"week", week}}}}
		uddateRes, err2 := coll.UpdateOne(context.TODO(), filter, updatedWorkut)
		if err2 != nil {
			log.Panicln(err2)
		}
		fmt.Fprintf(wr, "Updated workout %s", uddateRes.UpsertedID)
	} else if req.Method == "POST" && req.Header.Get("Content-Type") != "" {
		doc := bson.D{
			{"record", time.Now().Unix()},
			{"sets", workout.Sets},
			{"workout_date", workout.WorkoutDate},
			{"creation_date", time.Now().String()},
			{"workout_type", workout.WorkoutType},
			{"comments", workout.Comments},
			{"day", workout.Day},
			{"month", workout.Month},               // todo - take from workout_date
			{"week", workout.Week},                 // todo-take from date
			{"month", time.Now().Month().String()}, // todo-take from date
			{"year", time.Now().Year()},            // take from date
		}
		res, err := coll.InsertOne(context.TODO(), doc)
		//bodyJson, err := ioutil.ReadAll(req.Body)
		//fmt.Fprintf(wr, string(bodyJson))
		wr.Header().Set("Content-Type", "application/json")
		if err != nil {
			return
		}
		fmt.Printf("Created workout with _id: %v\n", res.InsertedID)
	}
}
