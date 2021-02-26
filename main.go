package main

import (
	"context"
	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"log"
	"net/http"
)

type App struct {
	Client          *mongo.Client
	Context         context.Context
	DB              *mongo.Database
	Router          *mux.Router
	UsersCollection *mongo.Collection
	CancelCallback  context.CancelFunc
}

func (a *App) PutInitialDataIntoDatabase() {
	//Insertion of user's data beforehand. Here, we are using upsert to avoid any duplicates
	upsert := true
	updateOpts := options.UpdateOptions{
		Upsert: &upsert,
	}
	_, err := a.UsersCollection.UpdateOne(a.Context, bson.M{"guid": "37e3f55c-7c34-439c-ab6d-60644d23cc7f"},
		bson.M{"$set": bson.M{"refresh_token": "", "expires_at": 0, "checksum": ""}},
		&updateOpts)

	if err != nil {
		log.Fatal(err)
	}
}

func (a *App) Initialize() {
	var err error
	var cancel context.CancelFunc
	a.Context, cancel = context.WithCancel(context.Background())
	a.CancelCallback = cancel

	a.Router = mux.NewRouter()

	//Mere database connection setup
	a.Client, err = mongo.Connect(a.Context, options.Client().ApplyURI("mongodb://localhost:27017"))
	if err != nil {
		log.Fatal(err)
	}
	//Access the database
	a.DB = a.Client.Database("jwt")

	//Get access to users collection
	a.UsersCollection = a.DB.Collection("users")

	//Fill database with initial data
	a.PutInitialDataIntoDatabase()
}

func main() {
	a := App{}
	addr := "127.0.0.1:4560"
	a.Initialize()
	defer a.CancelCallback()
	defer a.Client.Disconnect(a.Context)

	//Setup routes
	a.Router.Path("/signin").HandlerFunc(a.ObtainTokenPairApi).Methods("POST")
	a.Router.Path("/refresh").HandlerFunc(a.Refresh).Methods("POST")

	log.Printf("Running server on %s\n", addr)
	log.Fatal(http.ListenAndServe(addr, a.Router))

}
