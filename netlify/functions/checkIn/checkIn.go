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
	"github.com/twilio/twilio-go"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func handler(ctx context.Context, request events.APIGatewayProxyRequest) (*events.APIGatewayProxyResponse, error) {
	//verify token via phone and magic link
	//get user
	//set checkin Time to now + hours planning to stay
	//find user based requirements

	//add table to get history of all matches

	//phone best match
	//on their response, both users will get locked in and matched and will not be returned.

	lc, ok := lambdacontext.FromContext(ctx)
	if !ok {
		return &events.APIGatewayProxyResponse{
			StatusCode: 503,
			Body:       "Something went wrong :(",
		}, nil
	}

	cc := lc.ClientContext

	return &events.APIGatewayProxyResponse{
		StatusCode: 200,
		Body:       "Hello, " + cc.Client.AppTitle,
	}, nil
}

func main() {
	lambda.Start(handler)
}

type accessRequest struct {
	phone               string
	AuthenticationToken string
}
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

func respond(body string) (string, error) {

	target := accessRequest{}
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
	err = coll.FindOne(context.TODO(), bson.D{{"phone", target.phone}, {"token", target.AuthenticationToken}}).Decode(&result2)
	if result2 == "" {
		panic("user does not exist")
	}

	//find a user that best matches the current user that is active
	//insert algorithm here

	//add to a history table

	//text the match to wait for a responce.

	params := &openapi.CreateMessageParams{}
	params.SetTo("+1" + target.phone)
	params.SetFrom(os.Getenv("TWILIO_FROM_PHONE"))

	//magic link will be the access token for now. and it will be forever for now.
	//TODO: reset this magic key and handle proper authentication
	magicKey := result2.AuthenticationToken
	//this will redirect to our client and this will save the magic key
	magicLink := "https://" + os.Getenv("myClientSite") + "/id?=" + magicKey
	params.SetBody("Your phone number has been registered with two player belayer. Please open this link to complete registration. " + magicLink)

	resp, err := client.Api.CreateMessage(params)
	if err != nil {
		fmt.Println(err.Error())
		err = nil
	} else {
		fmt.Println("Message Sid: " + *resp.Sid)
	}
	return "Authentication String", result2.AuthenticationToken
}
