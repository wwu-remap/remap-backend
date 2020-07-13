package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/gridfs"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

func auth(w http.ResponseWriter, r *http.Request, apiKey string, authColl *mongo.Collection) (string, bool) {
	if r.Header.Get("x-api-key") != apiKey {
		log.Println("Bad x-api-key from", r.RemoteAddr)
		w.WriteHeader(http.StatusForbidden)
		fmt.Fprintf(w, "Wrong API key")
		return "", false
	}
	username, password, ok := r.BasicAuth()
	if !ok {
		log.Println("Could not get auth header from", r.RemoteAddr)
		w.Header().Set("WWW-Authenticate", "Basic realm=\"ReMAP\"")
		w.WriteHeader(http.StatusUnauthorized)
		fmt.Fprintf(w, "Unauthorized")
		return "", false
	}

	var result struct {
		Password string `bson:"password"`
	}
	err := authColl.FindOne(r.Context(), bson.M{"subjectId": username}).Decode(&result)
	if err != nil {
		log.Println("User", username, "not found from", r.RemoteAddr, ":", err)
		w.WriteHeader(http.StatusForbidden)
		fmt.Fprintf(w, "Unknown username")
		return "", false
	}
	if result.Password != password {
		log.Println("Wrong password for user", username, "from", r.RemoteAddr)
		w.WriteHeader(http.StatusForbidden)
		fmt.Fprintf(w, "Wrong password")
		return "", false
	}

	return username, true
}

func main() {
	if len(os.Args[1:]) != 4 {
		fmt.Printf("Usage: %s LISTEN_ADDR MONGODB_ADDR API_KEY\n", os.Args[0])
		return
	}

	listenAddr := os.Args[1]
	mongodbAddr := os.Args[2]
	mongodbName := os.Args[3]
	apiKey := os.Args[4]

	log.Println("Connecting to MongoDB at", mongodbAddr)
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(mongodbAddr))
	if err != nil {
		log.Printf("Could not connect to mongodb:", err)
		return
	}
	err = client.Ping(ctx, readpref.Primary())
	if err != nil {
		log.Printf("Could not connect to mongodb:", err)
		return
	}
	Db := client.Database(mongodbName)
	authColl := Db.Collection("auth")
	eventsColl := Db.Collection("events")

	http.HandleFunc("/events", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			log.Println("Bad request from", r.RemoteAddr)
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprintf(w, "Wrong method")
			return
		}

		username, ok := auth(w, r, apiKey, authColl)
		if !ok {
			return
		}

		if r.Header.Get("Content-Type") != "application/json" {
			log.Println("Wrong content-type for user", username, ":", r.Header.Get("Content-Type"))
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		bytes, err := ioutil.ReadAll(r.Body)
		if err != nil {
			log.Println("Could not read data for user", username, ":", err)
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintf(w, "Could not read data:", err)
			return
		}

		data := bson.D{}
		err = bson.UnmarshalExtJSON(bytes, true, &data)
		if err != nil {
			log.Println("Could not parse event for user", username, ":", err)
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprintf(w, "Could not parse data:", err)
			return
		}

		_, err = eventsColl.InsertOne(r.Context(), bson.M{"subjectId": username, "createdDate": time.Now(), "data": data})
		if err != nil {
			log.Println("Could not insert event for user", username, ":", err)
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintf(w, "Could not insert event:", err)
			return
		}

		fmt.Fprintf(w, "Success")
	})

	http.HandleFunc("/upload", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			log.Println("Bad request from", r.RemoteAddr)
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprintf(w, "Wrong method")
			return
		}

		username, ok := auth(w, r, apiKey, authColl)
		if !ok {
			return
		}

		if r.ContentLength > 16e6 {
			log.Println("Content too large for user", username, " from", r.RemoteAddr, ":", r.ContentLength)
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprintf(w, "Content too large")
			return
		}

		bucket, err := gridfs.NewBucket(Db)
		if err != nil {
			log.Println("Could not create bucket for user", username, ":", err)
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintf(w, "Could not create bucket:", err)
			return
		}

		metadata := options.GridFSUpload().SetMetadata(bson.M{"subjectId": username, "createdDate": time.Now()})
		_, err = bucket.UploadFromStream("upload", r.Body, metadata)
		if err != nil {
			log.Println("Could not upload stream for user", username, ":", err)
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintf(w, "Could not upload stream:", err)
			return
		}

		fmt.Fprintf(w, "Success")
	})

	http.HandleFunc("/tasks", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			log.Println("Bad request from", r.RemoteAddr)
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprintf(w, "Wrong method")
			return
		}

		_, ok := auth(w, r, apiKey, authColl)
		if !ok {
			return
		}

		body, err := loadTasks()
		if err != nil {
			w.WriteHeader(http.StatusNotFound)
			fmt.Fprint(w, "[]")
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write(body)

	})

	log.Println("Listening on", listenAddr, "...")
	log.Fatal(http.ListenAndServe(listenAddr, nil))
}

func loadTasks() ([]byte, error) {
	filename := "example/tasks.json"

	if _, err := os.Stat(filename); os.IsNotExist(err) {
		// filename does not exist
		return nil, err
	}

	body, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	return body, nil
}
