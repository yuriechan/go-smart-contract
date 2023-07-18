package main

import (
	"context"
	"crypto/ecdsa"
	"encoding/json"
	"fmt"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"go-smart-contract/api"
	"math/big"
	"net/http"
	"strconv"
)

const ethAddress = ""

func main() {
	client, err := ethclient.Dial("http://127.0.0.1:7545")
	if err != nil {
		panic(err)
	}

	auth := getAccountAuth(client, ethAddress)
	deployedContractAddress, tx, instance, err := api.DeployApi(auth, client)
	if err != nil {
		panic(err)
	}

	fmt.Println(deployedContractAddress.Hex())
	fmt.Println("instance->", instance)
	fmt.Println("tx->", tx.Hash().Hex())

	e := echo.New()
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	conn, err := api.NewApi(common.HexToAddress(deployedContractAddress.Hex()), client)
	if err != nil {
		panic(err)
	}
	e.GET("/balance", func(c echo.Context) error {
		reply, err := conn.Balance(&bind.CallOpts{})
		if err != nil {
			return err
		}
		return c.JSON(http.StatusOK, reply)
	})
	e.POST("/deposit/:amount", func(c echo.Context) error {
		amount := c.Param("amount")
		amt, _ := strconv.Atoi(amount)

		var v map[string]interface{}
		err := json.NewDecoder(c.Request().Body).Decode(&v)
		if err != nil {
			panic(err)
		}
		fmt.Println("v", v)
		auth := getAccountAuth(client, v["accountPublicAccess"].(string))

		reply, err := conn.Deposite(auth, big.NewInt(int64(amt)))
		if err != nil {
			return err
		}
		return c.JSON(http.StatusOK, reply)
	})

	e.Logger.Fatal(e.Start(":1323"))
}

func getAccountAuth(client *ethclient.Client, accountAddress string) *bind.TransactOpts {
	privateKey, err := crypto.HexToECDSA(accountAddress)
	if err != nil {
		panic(err)
	}

	publicKey := privateKey.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		panic("invalid key")
	}

	fromAddress := crypto.PubkeyToAddress(*publicKeyECDSA)

	nonce, err := client.PendingNonceAt(context.Background(), fromAddress)
	if err != nil {
		panic(err)
	}

	chainID, err := client.ChainID(context.Background())
	if err != nil {
		panic(err)
	}

	auth, err := bind.NewKeyedTransactorWithChainID(privateKey, chainID)
	if err != nil {
		panic(err)
	}
	auth.Nonce = big.NewInt(int64(nonce))
	auth.Value = big.NewInt(0)
	auth.GasLimit = uint64(3000000)
	auth.GasPrice = big.NewInt(1000000)

	return auth
}
