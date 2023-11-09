package main

import (
	"context"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Define your MongoDB connection string
const uri = "mongodb+srv://thanhvt:vuthithanh@cluster0.9vgulaq.mongodb.net/?retryWrites=true&w=majority"

// Create a global variable to hold our MongoDB connection
var mongoClient *mongo.Client

// This function runs before we call our main function and connects to our MongoDB database. If it cannot connect, the application stops.
func init() {
	if err := connect_to_mongodb(); err != nil {
		log.Fatal("Could not connect to MongoDB: ", err.Error())
	}
}

// Our entry point into our application
func main() {
	// The simplest way to start a Gin application using the frameworks defaults
	router := gin.Default()

	// Our route definitions
	router.GET("/movies", getMovies)
	router.GET("/movies/:id", getMovieByID)
	router.POST("/movies/aggregations", aggregateMovies)

	// The Run() method starts our Gin server
	router.Run("localhost:8080")
}

func connect_to_mongodb() error {
	serverAPI := options.ServerAPI(options.ServerAPIVersion1)
	opts := options.Client().ApplyURI(uri).SetServerAPIOptions(serverAPI)

	client, err := mongo.Connect(context.TODO(), opts)
	if err != nil {
		panic(err)
	}
	err = client.Ping(context.TODO(), nil)
	mongoClient = client
	return err
}

// Implemention of the /movies route that returns all of the movies from our movies collection.
func getMovies(c *gin.Context) {
	cursor, err := mongoClient.Database("sample_mflix").Collection("movies").Find(context.TODO(), bson.D{{}})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Map results
	var movies []bson.M
	if err = cursor.All(context.TODO(), &movies); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})

		return
	}

	// Return movies
	c.IndentedJSON(http.StatusOK, movies)
}

func getMovieByID(c *gin.Context) {
	idStr := c.Param("id")

	// Convert id string to ObjectId
	id, err := primitive.ObjectIDFromHex(idStr)
	if err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})

		return
	}

	// Find movie by ObjectId
	var movie bson.M
	err = mongoClient.Database("sample_mflix").Collection("movies").FindOne(context.TODO(), bson.D{
		{"_id", id},
	}).Decode(&movie)

	if err != nil {
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Return movie
	c.IndentedJSON(http.StatusOK, movie)
}

// The implementation of our /movies/aggregations endpoint that allows a user to pass in an aggregation to run our the movies collection.
func aggregateMovies(c *gin.Context) {
	// Get aggregation pipeline from request body
	var pipeline interface{}
	if err := c.ShouldBindJSON(&pipeline); err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Run aggregations
	cursor, err := mongoClient.Database("sample_mflix").Collection("movies").Aggregate(context.TODO(), pipeline)
	if err != nil {
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Map results
	var result []bson.M
	if err = cursor.All(context.TODO(), &result); err != nil {
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Return result
	c.IndentedJSON(http.StatusOK, result)
}
