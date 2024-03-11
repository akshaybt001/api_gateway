package graph

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strconv"

	"github.com/akshaybt001/api_gateway/authorize"
	"github.com/akshaybt001/api_gateway/middleware"
	"github.com/akshaybt001/proto_files/pb"
	"github.com/graphql-go/graphql"
)

var (
	Secret       []byte
	ProductsConn pb.ProductServiceClient
	UsersConn    pb.UserServiceClient
	CartConn     pb.CartServiceClient
)

func RetrieveSercet(secretString string) {
	Secret = []byte(secretString)
}

func Initialize(prodConn pb.ProductServiceClient, userConn pb.UserServiceClient, cartConn pb.CartServiceClient) {
	ProductsConn = prodConn
	UsersConn = userConn
	CartConn = cartConn
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

var UserType = graphql.NewObject(
	graphql.ObjectConfig{
		Name: "user",
		Fields: graphql.Fields{
			"id": &graphql.Field{
				Type: graphql.Int,
			},
			"name": &graphql.Field{
				Type: graphql.String,
			},
			"email": &graphql.Field{
				Type: graphql.String,
			},
			"password": &graphql.Field{
				Type: graphql.String,
			},
		},
	},
)

var CartType = graphql.NewObject(
	graphql.ObjectConfig{
		Name: "cart",
		Fields: graphql.Fields{
			"id": &graphql.Field{
				Type: graphql.Int,
			},
			"userId": &graphql.Field{
				Type: graphql.Int,
			},
			"productId": &graphql.Field{
				Type: graphql.Int,
			},
			"quantity": &graphql.Field{
				Type: graphql.Int,
			},
			"total": &graphql.Field{
				Type: graphql.Float,
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
						fmt.Println(prod)
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
			"userlogin": &graphql.Field{
				Type: UserType,
				Args: graphql.FieldConfigArgument{
					"email": &graphql.ArgumentConfig{
						Type: graphql.NewNonNull(graphql.String),
					},
					"password": &graphql.ArgumentConfig{
						Type: graphql.NewNonNull(graphql.String),
					},
				},
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					res, err := UsersConn.UserLogin(context.Background(), &pb.LoginRequest{
						Email:    p.Args["email"].(string),
						Password: p.Args["password"].(string),
					})
					if err != nil {
						return nil, err
					}
					token, err := authorize.GenerateJwt(uint(res.Id), false, false, Secret)
					if err != nil {
						fmt.Println("error here:", err)
						return nil, err
					}

					w := p.Context.Value("httpResponseWriter").(http.ResponseWriter)
					http.SetCookie(w, &http.Cookie{
						Name:  "jwtToken",
						Value: token,
						Path:  "/",
					})
					return res, nil
				},
			},
			"adminlogin": &graphql.Field{
				Type: UserType,
				Args: graphql.FieldConfigArgument{
					"email": &graphql.ArgumentConfig{
						Type: graphql.NewNonNull(graphql.String),
					},
					"password": &graphql.ArgumentConfig{
						Type: graphql.String,
					},
				},
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					res, err := UsersConn.AdminLogin(context.Background(), &pb.LoginRequest{
						Email:    p.Args["email"].(string),
						Password: p.Args["password"].(string),
					})
					if err != nil {
						return nil, err
					}
					token, err := authorize.GenerateJwt(uint(res.Id), true, false, Secret)
					if err != nil {
						return nil, err
					}
					w := p.Context.Value("httpResponseWriter").(http.ResponseWriter)

					http.SetCookie(w, &http.Cookie{
						Name:  "jwtToken",
						Value: token,
						Path:  "/",
					})
					return res, nil
				},
			},
			"supadminlogin": &graphql.Field{
				Type: UserType,
				Args: graphql.FieldConfigArgument{
					"email": &graphql.ArgumentConfig{
						Type: graphql.NewNonNull(graphql.String),
					},
					"password": &graphql.ArgumentConfig{
						Type: graphql.NewNonNull(graphql.String),
					},
				},
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					res, err := UsersConn.SupAdminLogin(context.Background(), &pb.LoginRequest{
						Email:    p.Args["email"].(string),
						Password: p.Args["password"].(string),
					})
					if err != nil {
						return nil, err
					}
					token, err := authorize.GenerateJwt(uint(res.Id), true, true, Secret)
					if err != nil {
						return nil, err
					}
					w := p.Context.Value("httpResponseWriter").(http.ResponseWriter)

					http.SetCookie(w, &http.Cookie{
						Name:  "jwtToken",
						Value: token,
						Path:  "/",
					})
					return res, nil
				},
			},
			"GetAllAdmins": &graphql.Field{
				Type: graphql.NewList(UserType),
				Resolve: middleware.SupAdminMiddleware(func(p graphql.ResolveParams) (interface{}, error) {
					admins, err := UsersConn.GetAllAdmins(context.Background(), &pb.NoPara{})
					if err != nil {
						return nil, err
					}
					var res []*pb.UserResponse
					for {
						admin, err := admins.Recv()
						if err == io.EOF {
							break
						}
						if err != nil {
							fmt.Println(err.Error())
						}
						res = append(res, admin)
					}
					return res, nil
				}),
			},
			"GetAllUser": &graphql.Field{
				Type: graphql.NewList(UserType),
				Resolve: middleware.AdminMiddleware(func(p graphql.ResolveParams) (interface{}, error) {
					users, err := UsersConn.GetAllUsers(context.Background(), &pb.NoPara{})
					if err != nil {
						return nil, err
					}
					var res []*pb.UserResponse
					for {
						user, err := users.Recv()
						if err == io.EOF {
							break
						}
						if err != nil {
							fmt.Println(err.Error())
						}
						res = append(res, user)
					}
					return res, nil
				}),
			},
			"GetAllCartItems": &graphql.Field{
				Type: graphql.NewList(CartType),
				Resolve: middleware.UserMiddleware(func(p graphql.ResolveParams) (interface{}, error) {
					userId := p.Context.Value("userId").(uint)
					cartItems, err := CartConn.GetAllCart(context.Background(), &pb.CartCreate{
						UserId: uint32(userId),
					})
					if err != nil {
						return nil, err
					}
					var res []*pb.GetAllCartResponse
					for {
						item, err := cartItems.Recv()
						if err == io.EOF {
							break
						}
						if err != nil {
							fmt.Println(err.Error())
						}
						res = append(res, item)
					}
					return res, nil
				}),
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
					res, err := ProductsConn.UpdateStock(context.Background(), &pb.UpdateStockRequest{
						Id:       uint32(id),
						Quantity: int32(p.Args["stock"].(int)),
						Increase: p.Args["increase"].(bool),
					})
					if err != nil {
						return nil, err
					}
					fmt.Println(res)
					return res, nil
				},
			},
			"UserSignUp": &graphql.Field{
				Type: UserType,
				Args: graphql.FieldConfigArgument{
					"name": &graphql.ArgumentConfig{
						Type: graphql.NewNonNull(graphql.String),
					},
					"email": &graphql.ArgumentConfig{
						Type: graphql.NewNonNull(graphql.String),
					},
					"password": &graphql.ArgumentConfig{
						Type: graphql.NewNonNull(graphql.String),
					},
				},
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					name, _ := p.Args["name"].(string)
					email, _ := p.Args["email"].(string)
					password, _ := p.Args["password"].(string)

					if name == "" || email == "" || password == "" {
						return nil, fmt.Errorf("name,email,and password are required")
					}
					res, err := UsersConn.UserSignUp(context.Background(), &pb.UserSignUpRequest{
						Name:     p.Args["name"].(string),
						Email:    p.Args["email"].(string),
						Password: p.Args["password"].(string),
					})
					if err != nil {
						return nil, err
					}
					fmt.Println("before cart")
					cart, err := CartConn.CreateCart(context.Background(), &pb.CartCreate{
						UserId: res.Id,
					})
					if err != nil {
						return nil, err
					}
					if cart.UserId == 0 {
						return nil, fmt.Errorf("error while creating cart")
					}
					fmt.Println("cart user id ", cart.UserId, cart.CartId)
					response := &pb.UserResponse{
						Id:    res.Id,
						Name:  res.Name,
						Email: res.Email,
					}
					return response, nil
				},
			},
			"addAdmin": &graphql.Field{
				Type: UserType,
				Args: graphql.FieldConfigArgument{
					"name": &graphql.ArgumentConfig{
						Type: graphql.NewNonNull(graphql.String),
					},
					"email": &graphql.ArgumentConfig{
						Type: graphql.NewNonNull(graphql.String),
					},
					"password": &graphql.ArgumentConfig{
						Type: graphql.NewNonNull(graphql.String),
					},
				},
				Resolve: middleware.SupAdminMiddleware(func(p graphql.ResolveParams) (interface{}, error) {

					admin, err := UsersConn.AddAdmin(context.Background(), &pb.UserSignUpRequest{
						Name:     p.Args["name"].(string),
						Email:    p.Args["email"].(string),
						Password: p.Args["password"].(string),
					})
					if err != nil {
						fmt.Println(err.Error())
						return nil, err
					}
					return admin, nil
				}),
			},
			"AddToCart": &graphql.Field{
				Type: CartType,
				Args: graphql.FieldConfigArgument{
					"productId": &graphql.ArgumentConfig{
						Type: graphql.NewNonNull(graphql.Int),
					},
					"quantity": &graphql.ArgumentConfig{
						Type: graphql.NewNonNull(graphql.Int),
					},
				},
				Resolve: middleware.UserMiddleware(func(p graphql.ResolveParams) (interface{}, error) {
					userIDval := p.Context.Value("userId").(uint)
					res, err := CartConn.AddToCart(context.Background(), &pb.AddToCartRequst{
						UserId:   uint32(userIDval),
						ProdId:   uint32(p.Args["productId"].(int)),
						Quantity: int32(p.Args["quantity"].(int)),
					})
					if err != nil {
						return nil, err
					}
					return res, nil

				}),
			},
			"RemoveFromCart": &graphql.Field{
				Type: CartType,
				Args: graphql.FieldConfigArgument{
					"productId": &graphql.ArgumentConfig{
						Type: graphql.NewNonNull(graphql.Int),
					},
				},
				Resolve: middleware.UserMiddleware(func(p graphql.ResolveParams) (interface{}, error) {
					userId := p.Context.Value("userId").(uint)
					return CartConn.RemoveCart(context.Background(), &pb.RemoveCartRequest{
						UserId: uint32(userId),
						ProdId: uint32(p.Args["productId"].(int)),
					})
				}),
			},
		},
	},
)

var Schema, _ = graphql.NewSchema(graphql.SchemaConfig{
	Query:    RootQuery,
	Mutation: Mutation,
})
