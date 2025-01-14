package main

import (
	"fmt"
	"log/slog"
	"strings"
	"math/big"
	"net/http"
	"os"

	ethAbi "github.com/ethereum/go-ethereum/accounts/abi"
	ethAbiBind "github.com/ethereum/go-ethereum/accounts/abi/bind"
	ethCommon "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"

	"github.com/gorilla/rpc"
	"github.com/gorilla/rpc/json"

	"github.com/aws/aws-lambda-go/lambda"

	"github.com/awslabs/aws-lambda-go-api-proxy/httpadapter"
)

var abi, _ = ethAbi.JSON(strings.NewReader(`[{"type":"function","name":"tokenReservesFFCCDB8F","inputs":[{"name":"pool","type":"address","internalType":"address"}],"outputs":[{"name":"","type":"uint256","internalType":"uint256"},{"name":"","type":"uint256","internalType":"uint256"}],"stateMutability":"nonpayable"}]`))

type Service struct {
	c *ethclient.Client
	a ethCommon.Address
}

type (
	Args struct {
		Pool string
	}

	Resp struct {
		Address          string
		Liq0Str, Liq1Str string
		Liq0Big, Liq1Big *big.Int
	}
)

func (s Service) Reserves(r *http.Request, args *Args, reply *Resp) error {
	c := ethAbiBind.NewBoundContract(s.a, abi, s.c, s.c, s.c)
	var a []any
	if !ethCommon.IsHexAddress(args.Pool) {
		return fmt.Errorf("not address")
	}
	err := c.Call(
		&ethAbiBind.CallOpts{
			Context: r.Context(),
		},
		&a,
		"tokenReservesFFCCDB8F",
		ethCommon.HexToAddress(args.Pool),
	)
	if err != nil {
		slog.Error("get pool reserves", "err", err, "addr", s.a)
		return fmt.Errorf("requesting reserves")
	}
	amt0, ok := a[0].(*big.Int)
	if !ok {
		slog.Error("convert amt 0", "a", a, "a0", a[0], "addr", s.a)
		return fmt.Errorf("decoding reserves0")
	}
	amt1, ok := a[1].(*big.Int)
	if !ok {
		slog.Error("convert amt 1", "a", a, "a1", a[1], "addr", s.a)
		return fmt.Errorf("decoding reserves1")
	}
	reply.Liq0Str = amt0.String()
	reply.Liq1Str = amt1.String()
	reply.Liq0Big = amt0
	reply.Liq1Big = amt1
	return nil
}

func main() {
	c, err := ethclient.Dial(os.Getenv("SPN_SUPERPOSITION_URL"))
	if err != nil {
		panic(err)
	}
	defer c.Close()
	a := ethCommon.HexToAddress(os.Getenv("SPN_LONGTAIL_ADDR"))
	s := Service{c, a}
	r := rpc.NewServer()
	r.RegisterService(&s, "")
	r.RegisterCodec(json.NewCodec(), "application/json")
	http.Handle("/", r)
	lambda.Start(httpadapter.NewV2(http.DefaultServeMux).ProxyWithContext)
}
