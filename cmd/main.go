package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/Gopher0727/RTMP/config"
	"github.com/gin-gonic/gin"
)

func main() {
	config, err := config.LoadConfig("config.toml")
	if err != nil {
		log.Fatal(err)
	}

	// todo
	fmt.Printf("AppName: %s, Env: %s, MySQLHost: %s\n", config.AppName, config.Env, config.MySQL.Host)

	r := gin.New()

	r.Use(gin.Logger())

	r.GET("/ping", func(ctx *gin.Context) {
		ctx.JSON(http.StatusOK, gin.H{
			"message": "pong",
		})
	})

	r.Run(":8080")
}
