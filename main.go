package main

import (
	"blm/bloomFilter"
	"fmt"
	"github.com/foolin/gin-template"
	"github.com/gin-contrib/cors"
	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
	"github.com/lw000/gocommon/db/rdsex"
	"github.com/lw000/gocommon/utils"
	"github.com/utrack/gin-csrf"

	"net/http"
	"time"

	log "github.com/sirupsen/logrus"
)

var bf *bloomf.BloomFilter

func main() {
	log.SetFormatter(&log.TextFormatter{TimestampFormat: "2006-01-02 15:04:05"})

	bf := bloomf.New(bloomf.DEFAULT_SIZE)
	bf.Add("456")
	var s = "456"
	if bf.Contains(s) {
		log.Printf("%s is exists", s)
	}

	engine := gin.Default()
	engine.HTMLRender = gintemplate.Default()

	store := cookie.NewStore([]byte("secret"))
	engine.Use(sessions.Sessions("mysession", store))
	engine.Use(csrf.Middleware(csrf.Options{
		Secret:        "secret123",
		IgnoreMethods: nil,
		ErrorFunc: func(c *gin.Context) {
			c.String(http.StatusNotFound, "CSRF token mismatch")
			c.Abort()
		},
		TokenGetter: nil,
	}))

	engine.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"https://foo.com"},
		AllowMethods:     []string{"GET", "POST"},
		AllowHeaders:     []string{"Origin"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		AllowOriginFunc: func(origin string) bool {
			return origin == "https://github.com"
		},
		MaxAge: 12 * time.Hour,
	}))

	engine.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"c": 0, "m": "布隆过滤", "d": gin.H{}})
	})

	engine.GET("/blm", func(c *gin.Context) {
		value := c.DefaultQuery("value", "")
		if value == "" {
			c.JSON(http.StatusOK, gin.H{"c": 0, "m": "value is empty"})
			return
		}

		if bf.Contains(value) {
			c.JSON(http.StatusOK, gin.H{"c": 0, "m": "exists", "d": gin.H{
				"value": value,
			}})
			return
		}

		bf.Add(value)
		c.JSON(http.StatusOK, gin.H{"c": 0, "m": fmt.Sprintf("%s add success", value)})
	})

	engine.GET("/protected", func(c *gin.Context) {
		c.String(200, csrf.GetToken(c))
	})

	TestRedis()

	engine.Run(":12580")
}

func TestRedis() {
	cfg, err := tyrdsex.LoadJsonConfig("conf/redis.json")
	if err != nil {
		log.Panic(err)
	}
	log.Println(cfg)

	rds := &tyrdsex.RdsServer{}
	err = rds.OpenWithJsonConfig(cfg)
	if err != nil {
		log.Print(err)
		return
	}

	defer func() {
		_ = rds.Close()
	}()

	for i := 0; i < 10; i++ {
		_, _ = rds.Set(fmt.Sprintf("user:name%d", i), "levi", -1)
	}

	r, err := rds.Get("user:name3")
	if err != nil {
		log.Panic(err)
	}

	log.Println(r)

	for i := 0; i < 10; i++ {
		token := tyutils.UUID()
		_, _ = rds.Set("tokens:"+token, "1111", time.Second*time.Duration(300))
	}
}
