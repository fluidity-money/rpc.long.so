package main

import (
	"database/sql"
	"fmt"
	"log/slog"
	"math/big"
	"net/http"
	"os"

	longTypes "github.com/fluidity-money/long.so/lib/types"

	_ "github.com/lib/pq"

	"github.com/gorilla/rpc"
	"github.com/gorilla/rpc/json"

	"github.com/aws/aws-lambda-go/lambda"

	"github.com/awslabs/aws-lambda-go-api-proxy/httpadapter"
)

type Service struct {
	db *sql.DB
}

type (
	PoolsArgs struct{}

	Pool struct {
		Address          string
		Decimals         int
		Liq0Str, Liq1Str string
		Liq0Big, Liq1Big *big.Int
	}

	PoolsResp struct {
		Pools []Pool
	}
)

func (s Service) Pools(r *http.Request, args *PoolsArgs, reply *PoolsResp) error {
	rows, err := s.db.Query(`
SELECT pool, decimals, cumulative_amount0, cumulative_amount1
FROM snapshot_positions_latest_decimals_grouped_1`,
	)
	if err != nil {
		slog.Error("Failed to search pools using snapshot_positions_latest_decimals_grouped_1",
			"error", err,
		)
		return fmt.Errorf("search snapshots")
	}
	defer rows.Close()
	for rows.Next() {
		var (
			pool       longTypes.Address
			decimals   int
			liq0, liq1 longTypes.Number
		)
		if err := rows.Scan(&pool, &decimals, &liq0, &liq1); err != nil {
			slog.Error("Error scanning pools",
				"error", err,
			)
			return fmt.Errorf("scanning pools")
		}
		reply.Pools = append(reply.Pools, Pool{
			Address:  pool.String(),
			Decimals: decimals,
			Liq0Str:  liq0.String(),
			Liq1Str:  liq1.String(),
			Liq0Big:  liq0.Int,
			Liq1Big:  liq1.Int,
		})
	}
	return nil
}

func main() {
	db, err := sql.Open("postgres", os.Getenv("SPN_TIMESCALE"))
	if err != nil {
		panic(err)
	}
	defer db.Close()
	s := Service{db}
	r := rpc.NewServer()
	r.RegisterService(&s, "")
	r.RegisterCodec(json.NewCodec(), "application/json")
	http.Handle("/", r)
	lambda.Start(httpadapter.NewV2(http.DefaultServeMux).ProxyWithContext)
}
