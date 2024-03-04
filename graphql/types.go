package graph

import (
	"context"
	"fmt"
	"io"
	"strconv"

	"github.com/akshaybt001/proto_files/pb"
	"github.com/graphql-go/graphql"
)

var (
	secret       []byte
	ProductsConn pb.ProductServiceClient
)

func RetrieveSercet(secretString string) {
	secret = []byte(secretString)
}

func Initialize(prodConn pb.ProductServiceClient) {
	ProductsConn = prodConn
}

var ProductType = graphql.NewObject(
	graphql.ObjectConfig{
		Name: "product",
		Fields: graphql.Fields{
			"id": &graphql.Field{
				Type: graphql.Int,
			},
			"name": &graphql.Field{
				Type: graphql.String,
			},
			"price": &graphql.Field{
				Type: graphql.Int,
			},
			"quantity": &graphql.Field{
				Type: graphql.Int,
			},
		},
	},
)

var RootQuery = graphql.NewObject(
	graphql.ObjectConfig{
		Name: "RootQuery",
		Fields: graphql.Fields{
			"product": &graphql.Field{
				Type: ProductType,
				Args: graphql.FieldConfigArgument{
					"id": &graphql.ArgumentConfig{
						Type: graphql.NewNonNull(graphql.Int),
					},
				},
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					return ProductsConn.GetProduct(context.Background(), &pb.GetProductByID{
						Id: uint32(p.Args["id"].(int)),
					})
				},
			},
			"products": &graphql.Field{
				Type: graphql.NewList(ProductType),
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {

					var res []*pb.ProductResponse

					products, err := ProductsConn.GetAllProduct(context.Background(), &pb.NoParam{})
					if err != nil {
						fmt.Println(err.Error())
					}

					for {
						prod, err := products.Recv()
						if err == io.EOF {
							break
						}
						if err != nil {
							fmt.Println(err)
						}
						res = append(res, prod)
					}
					return res, err
				},
			},
		},
	},
)

var Mutation = graphql.NewObject(
	graphql.ObjectConfig{
		Name: "Mutation",
		Fields: graphql.Fields{
			"AddProduct": &graphql.Field{
				Type: ProductType,
				Args: graphql.FieldConfigArgument{
					"name": &graphql.ArgumentConfig{
						Type: graphql.NewNonNull(graphql.String),
					},
					"price": &graphql.ArgumentConfig{
						Type: graphql.NewNonNull(graphql.Int),
					},
					"quantity": &graphql.ArgumentConfig{
						Type: graphql.NewNonNull(graphql.Int),
					},
				},
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					products, err := ProductsConn.AddProduct(context.Background(), &pb.AddProductRequest{
						Name:     p.Args["name"].(string),
						Price:    int32(p.Args["price"].(int)),
						Quantity: int32(p.Args["quantity"].(int)),
					})
					if err != nil {
						fmt.Println(err.Error())
						return nil, err
					}
					return products, nil
				},
			},
			"UpdateStock": &graphql.Field{
				Type: ProductType,
				Args: graphql.FieldConfigArgument{
					"id": &graphql.ArgumentConfig{
						Type: graphql.NewNonNull(graphql.ID),
					},
					"stock": &graphql.ArgumentConfig{
						Type: graphql.NewNonNull(graphql.Int),
					},
					"increase": &graphql.ArgumentConfig{
						Type: graphql.NewNonNull(graphql.Boolean),
					},
				},
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					id, _ := strconv.Atoi(p.Args["id"].(string))
					return ProductsConn.UpdateStock(context.Background(), &pb.UpdateStockRequest{
						Id:       uint32(id),
						Quantity: int32(p.Args["stock"].(int)),
						Increase: p.Args["increase"].(bool),
					})
				},
			},
		},
	},
)

var Schema, _ = graphql.NewSchema(graphql.SchemaConfig{
	Query:    RootQuery,
	Mutation: Mutation,
})
