package main

import (
	"context"
	"encoding/json"

	"fmt"
	"log"
	"os"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-lambda-go/lambdacontext"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/twilio/twilio-go"
	openapi "github.com/twilio/twilio-go/rest/api/v2010"
)

type user struct {
	phone              string
	preferedName       string
	iAmOver18          bool
	belayCertified     bool
	pronouns           string
	apeSpan            int
	height             int
	weight             int
	experience         string
	favoriteEmoji      string
	afterHoursMatching bool

	preferedRangeMaxWeight int
	preferedBelayCertified bool
	preferedExperience     string
	disregardFlag          bool

	//automagically set
	lastCheckIn         string //time.Date
	availableUntil      string //time.Date
	AuthenticationToken string
	userMatchExpiration string //time.Date
	reportedCount       int
}

func handler(ctx context.Context, request events.APIGatewayProxyRequest) (*events.APIGatewayProxyResponse, error) {
	authenticationResult, err := register(request.Body)

	_, ok := lambdacontext.FromContext(ctx)

	if !ok {
		return &events.APIGatewayProxyResponse{
			StatusCode: 503,
			Body:       err.Error(),
		}, nil
	}

	//cc := lc.ClientContext

	//return key
	return &events.APIGatewayProxyResponse{
		StatusCode: 200,
		Body:       authenticationResult,
	}, nil
}

func main() {
	lambda.Start(handler)

}

func register(body string) (string, error) {

	target := user{}
	input := []byte(body)
	json.Unmarshal(input, &target)
	//	json.NewDecoder(body).Decode(target)

	uri := os.Getenv("CONNECTION_STRING")
	if uri == "" {
		log.Fatal("You must set your 'MONGODB_URI' environmental variable. See\n\t https://www.mongodb.com/docs/drivers/go/current/usage-examples/#environment-variable")
	}
	mongoClient, err := mongo.Connect(context.TODO(), options.Client().ApplyURI(uri))
	if err != nil {
		panic(err)
	}

	defer func() {
		if err := mongoClient.Disconnect(context.TODO()); err != nil {
			panic(err)
		}
	}()

	coll := mongoClient.Database("twoPlayerBelayer").Collection("users")

	//TEXT
	client := twilio.NewRestClientWithParams(twilio.ClientParams{
		Username: os.Getenv("TWILIO_ACCOUNT_SID"),
		Password: os.Getenv("TWILIO_AUTH_TOKEN"),
	})
	result2 := ""
	//check phone number for duplicates; throw error if true
	err = coll.FindOne(context.TODO(), bson.D{{"phone", 5}}).Decode(&result2)
	if result2 != "" {
		panic("user exists")
	}

	//Insert new record;
	//return authentication/ text magic link

	coll.InsertOne(context.TODO(), target)
	// if err != nil {
	// 	panic(err)
	// }

	params := &openapi.CreateMessageParams{}
	params.SetTo("+1" + target.phone)
	params.SetFrom(os.Getenv("TWILIO_FROM_PHONE"))
	params.SetBody("test")

	resp, err := client.Api.CreateMessage(params)
	if err != nil {
		fmt.Println(err.Error())
		err = nil
	} else {
		fmt.Println("Message Sid: " + *resp.Sid)
	}
	return "Authentication String", nil
}
