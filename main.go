package main

import (
	"fmt"
	"github.com/foolin/gin-template"
	"github.com/gin-contrib/cors"
	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
	"github.com/utrack/gin-csrf"
	"github.com/willf/bitset"
	"net/http"
	"time"

	"log"
)

const DEFAULT_SIZE = 2 << 24

var seeds = []uint{7, 11, 13, 31, 37, 61}

type SimpleHash struct {
	cap  uint
	seed uint
}

func (s SimpleHash) Hash(value string) uint {
	var result uint = 0
	for i := 0; i < len(value); i++ {
		result = result*s.seed + uint(value[i])
	}
	return (s.cap - 1) & result
}

type BloomFilter struct {
	b   *bitset.BitSet
	fns [6]SimpleHash
}

func New(size uint) *BloomFilter {
	bf := &BloomFilter{}
	bf.b = bitset.New(DEFAULT_SIZE)
	for i := 0; i < len(seeds); i++ {
		bf.fns[i] = SimpleHash{
			cap:  DEFAULT_SIZE,
			seed: seeds[i],
		}
	}
	return bf
}

func (bf *BloomFilter) Add(value string) {
	if value == "" {
		return
	}

	for _, fn := range bf.fns {
		bf.b.Set(fn.Hash(value))
	}
}

func (bf *BloomFilter) Contains(value string) bool {
	if value == "" {
		return false
	}

	ret := true
	for _, fn := range bf.fns {
		ret = bf.b.Test(fn.Hash(value))
	}
	return ret
}

func (bf *BloomFilter) Load(filename string) bool {
	if filename == "" {
		return false
	}

	return true
}

var bf *BloomFilter

func main() {
	bf := New(DEFAULT_SIZE)
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

	engine.Run(":12580")
}
