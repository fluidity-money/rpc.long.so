package main

import (
	"database/sql"
	"math/big"
	"net/http"

	"github.com/gorilla/rpc"
	"github.com/gorilla/rpc/json"

	"github.com/aws/aws-lambda-go/lambda"

	"github.com/awslabs/aws-lambda-go-api-proxy/httpadapter"
)

type service struct {
	db *sql.DB
}

type (
	PoolsArgs struct{}

	Pool struct {
		Address, TokenName string
		Liq0Str, Liq1Str   string
		Liq0Big, Liq1Big   *big.Int
	}

	PoolsResp struct {
		Pools []Pool
	}
)

func (s service) Pools(r *http.Request, args *PoolsArgs, reply *PoolsResp) error {
	reply.Pools = []Pool{{
		Address: "Alex",
	}}
	return nil
}

func main() {
	s := service{}
	r := rpc.NewServer()
	r.RegisterService(&s, "")
	r.RegisterCodec(json.NewCodec(), "application/json")
	http.Handle("/", r)
	lambda.Start(httpadapter.NewV2(http.DefaultServeMux).ProxyWithContext)
}
