package main

import (
	"context"

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

func handler(ctx context.Context, request events.APIGatewayProxyRequest) (*events.APIGatewayProxyResponse, error) {
	testCode()


	lc, ok := lambdacontext.FromContext(ctx)
	if !ok {
		return &events.APIGatewayProxyResponse{
			StatusCode: 503,
			Body:       "Something went wrong :(",
		}, nil
	}

	cc := lc.ClientContext
	//return key
	return &events.APIGatewayProxyResponse{
		StatusCode: 200,
		Body:       "Hello, " + cc.Client.AppTitle,
	}, nil
}

func main() {
	lambda.Start(handler)

}

func testCode() {

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

	//check phone number for duplicates; throw error if true
	//Insert new record;
	//return authentication/ text magic link

	coll := mongoClient.Database("twoPlayerBelayer").Collection("users")

	doc := bson.D{
		{"preferedName", ""},
		{"IAmOver18", ""},
		{"phone", ""},
		{"belayCertified", ""},

		{"pronouns", ""},
		{"apeSpan", ""},
		{"height", ""},
		{"weight", ""}, // ignore weight if null
		{"gender", ""},
		{"experience", ""},
		{"favoriteEmoji/ profileEmoji", ""},
		{"afterHoursMatching", ""},

		{"preferedRangeMaxWeight", ""},
		{"preferedGenderMaybe?", ""},
		{"preferedExperience", ""},
		{"disregardFlag", ""},//ignore this and set the above as any

		//automagically set
		{"lastCheckIn", ""},
		{"availableUntil", ""},
		{"AuthenticationToken", ""},
		{"userMatchExpiration", ""},//we dont want to match people to more than 1 person
		{"reportedCount", ""},//anyone over X amount of reports will be shadow banned, array of complaints
	}
	coll.InsertOne(context.TODO(), doc)
	// if err != nil {
	// 	panic(err)
	// }
	//TEXT
	client := twilio.NewRestClientWithParams(twilio.ClientParams{
		Username: os.Getenv("TWILIO_ACCOUNT_SID"),
		Password: os.Getenv("TWILIO_AUTH_TOKEN"),
	})

	params := &openapi.CreateMessageParams{}
	params.SetTo(os.Getenv("TWILIO_TO_PHONE"))
	params.SetFrom(os.Getenv("TWILIO_FROM_PHONE"))
	params.SetBody("test")

	resp, err := client.Api.CreateMessage(params)
	if err != nil {
		fmt.Println(err.Error())
		err = nil
	} else {
		fmt.Println("Message Sid: " + *resp.Sid)
	}
}
