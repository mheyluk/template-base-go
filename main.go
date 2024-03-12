package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"template-base-go/src/core"
	"template-base-go/src/handlers"
	"template-base-go/src/repositories"
	"template-base-go/src/services"
	"template-base-go/src/utils"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	gorillaHandlers "github.com/gorilla/handlers"
	"go.mongodb.org/mongo-driver/mongo"
)

var handlersContainer *handlers.Container

func loadEnvs() {
	if err := utils.ParseEnvironmentFile(""); err != nil {
		log.Fatal(err)
	}
}

func main() {
	loadEnvs()

	utilsContainer := utils.NewContainer(utils.NewEnvironment())

	isLambda := utilsContainer.Environment.IsLambda()

	fmt.Println("IsLambda?: ", isLambda)

	servicesContainer := services.NewContainer(&utils.Logger{}, &utils.Environment{}, &services.ExampleService{})
	dbConnection := getDbConnection(utilsContainer)
	repositoriesContainer := repositories.NewContainer(repositories.NewExampleRepository(dbConnection))
	handlersContainer = handlers.NewContainer(handlers.NewExampleHandler(servicesContainer.Logger, servicesContainer.ExampleServices, repositoriesContainer.ExampleRepository))

	if isLambda {
		lambda.Start(LambdaHandler)
	} else {
		httpServer(handlersContainer, utilsContainer)
	}

	logServerInfo(utilsContainer, servicesContainer)

}

func httpServer(handlersContainer *handlers.Container, utilsContainer *utils.Container) {

	api := core.NewApi(handlersContainer, &utils.Logger{})

	port := getServerPort(utilsContainer)
	log.Fatal(http.ListenAndServe(port, setupCORS(api.Router())))
}

func setupCORS(handler http.Handler) http.Handler {
	headers := gorillaHandlers.AllowedHeaders([]string{"Content-Type", "Access-Control-Allow-Headers", "Authorization", "X-Requested-With"})
	methods := gorillaHandlers.AllowedMethods([]string{"DELETE", "POST", "GET", "OPTIONS", "PUT", "PATCH"})
	origins := gorillaHandlers.AllowedOrigins([]string{"*"})
	return gorillaHandlers.CORS(headers, methods, origins)(handler)
}

func logServerInfo(utilsContainer *utils.Container, servicesContainer *services.Container) {
	port := getServerPort(utilsContainer)
	serverURL := fmt.Sprintf("Server url: http://localhost%s", port)
	swaggerURL := fmt.Sprintf("%s/api-docs/index.html", serverURL)
	servicesContainer.Logger.Info("Initializing server")
	servicesContainer.Logger.Info(fmt.Sprintf("Environment: %s", utilsContainer.Environment.GetEnvVar("ENV")))
	servicesContainer.Logger.Info(serverURL)
	servicesContainer.Logger.Info(fmt.Sprintf("Swagger url: %s", swaggerURL))
}

func getServerPort(utilsContainer *utils.Container) string {
	port := fmt.Sprintf(":%v", utilsContainer.Environment.GetEnvVar("PORT"))
	if port == ":" {
		port = ":8080"
	}
	return port
}

func getDbConnection(utilsContainer *utils.Container) *mongo.Database {

	var (
		uri     = utilsContainer.Environment.GetEnvVar("MONGO_URI")
		db_name = utilsContainer.Environment.GetEnvVar("DB_NAME")
	)

	return repositories.Connect(uri, db_name)
}

func LambdaHandler(ctx context.Context, request events.LambdaFunctionURLRequest) (events.APIGatewayProxyResponse, error) {
	api := core.NewApi(handlersContainer, &utils.Logger{})

	req, err := utils.LambdaEventToHttpRequest(request)
	if err != nil {
		return events.APIGatewayProxyResponse{StatusCode: http.StatusInternalServerError, Body: err.Error()}, nil
	}

	recorder := httptest.NewRecorder()
	api.Router().ServeHTTP(recorder, req)

	response := events.APIGatewayProxyResponse{
		StatusCode: recorder.Code,
		Body:       recorder.Body.String(),
		Headers:    map[string]string{"Content-Type": "application/json"},
	}

	return response, nil
}
