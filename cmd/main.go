package main

import (
	"context"
	"log"
	"net/http"
	"os"

	graph "github.com/akshaybt001/api_gateway/graphql"
	"github.com/akshaybt001/proto_files/pb"
	"github.com/graphql-go/handler"
	"github.com/joho/godotenv"
	"google.golang.org/grpc"
)

func main() {
	productConn, err := grpc.Dial("localhost:8080", grpc.WithInsecure())
	if err != nil {
		log.Println(err.Error())
	}

	defer func() {
		productConn.Close()
	}()
	productRes := pb.NewProductServiceClient(productConn)

	if err := godotenv.Load("../.env"); err != nil {
		log.Fatalf(err.Error())
	}
	secretString := os.Getenv("SECRET")

	graph.Initialize(productRes)
	graph.RetrieveSercet(secretString)

	h := handler.New(&handler.Config{
		Schema: &graph.Schema,
		Pretty: true,
	})

	http.HandleFunc("/graphql", func(w http.ResponseWriter, r *http.Request) {
		// Add the http.ResponseWriter to the context.
		ctx := context.WithValue(r.Context(), "httpResponseWriter", w)
		ctx = context.WithValue(ctx, "request", r)

		//Update the request's context.
		r = r.WithContext(ctx)

		//call the Graphql handler.
		h.ContextHandler(ctx, w, r)

	})

	log.Println("listening on port : 8080 of api gateway")

	http.ListenAndServe(":8080", nil)

}
