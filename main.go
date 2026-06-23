package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"github.com/joho/godotenv"
)

// 定义存储对象结构
type MediaItem struct {
	ID           string    `json:"id"`
	Type         string    `json:"type"` // article, photo, document
	Title        string    `json:"title"`
	Content      string    `json:"content"`       // 存储的文件名
	OriginalName string    `json:"original_name"` // ✅ 新增：原始文件名
	CreatedAt    time.Time `json:"created_at"`
}

// 初始化Redis客户端
func initRedis() *redis.Client {
	// redisURL := os.Getenv("UPSTASH_REDIS_REST_URL")
	redisURL := os.Getenv("UPSTASH_REDIS_REST_URL")
	if redisURL == "" {
		log.Fatal("UPSTASH_REDIS_REST_URL环境变量未设置")
	}

	opt, err := redis.ParseURL(redisURL)
	if err != nil {
		log.Fatalf("解析Redis URL失败: %v", err)
	}

	return redis.NewClient(opt)
}

func main() {
	// 初始化环境变量godotenv .env文件
	err := godotenv.Load()
	if err != nil {
		log.Fatal("加载.env环境变量失败")
	}
	// os.Setenv("UPSTASH_REDIS_REST_URL", "rediss://@.upstash.io:6379")
	// 设置Gin模式
	if os.Getenv("ENV") == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	// 初始化Redis客户端
	rdb := initRedis()

	// 创建文件上传目录
	if _, err := os.Stat("uploads"); os.IsNotExist(err) {
		os.Mkdir("uploads", 0755)
	}

	// 初始化Gin引擎
	r := gin.Default()

	// 静态文件服务
	r.Static("/static", "./static")
	r.Static("/uploads", "./uploads") // Uploads目录读取权限开关
	r.LoadHTMLGlob("templates/*")

	// 首页

	// 尝试使用.env设置中文app标题遇到乱码
	app_title := os.Getenv("APP_TITLE")
	fmt.Println(app_title)
	if app_title == "" {
		app_title = "Jin"
	}

	r.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "index.tmpl", gin.H{
			"title": app_title + "多媒体分享平台",
			//"title": app_title,
		})
	})

	// API路由组
	api := r.Group("/api")
	{
		// 获取所有媒体项
		api.GET("/items", func(c *gin.Context) {
			keys, err := rdb.Keys(c.Request.Context(), "item:*").Result()
			fmt.Println("rd1-read-key")
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "获取媒体项失败"})
				return
			}

			if len(keys) == 0 {
				c.JSON(http.StatusOK, []MediaItem{})
				return
			}

			// 使用 MGet 一次性获取多个键的值
			results, err := rdb.MGet(c.Request.Context(), keys...).Result()
			fmt.Println("rd2-read-Mget")
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "获取媒体项失败"})
				return
			}

			var items []MediaItem
			for _, result := range results {
				if result == nil {
					continue
				}
				data, ok := result.(string)
				if !ok {
					continue
				}

				var item MediaItem
				if err := json.Unmarshal([]byte(data), &item); err != nil {
					continue
				}

				items = append(items, item)
			}

			// 反转列表顺序，使最新的文章在最上面
			for i, j := 0, len(items)-1; i < j; i, j = i+1, j-1 {
				items[i], items[j] = items[j], items[i]
			}

			c.JSON(http.StatusOK, items)
		})

		// 上传文章
		api.POST("/articles", func(c *gin.Context) {
			title := c.PostForm("title")
			content := c.PostForm("content")

			if title == "" || content == "" {
				c.JSON(http.StatusBadRequest, gin.H{"error": "标题和内容不能为空"})
				return
			}

			item := MediaItem{
				ID:        fmt.Sprintf("article-%d", time.Now().UnixNano()),
				Type:      "article",
				Title:     title,
				Content:   content,
				CreatedAt: time.Now(),
			}

			itemJSON, _ := json.Marshal(item)

			exp, _ := strconv.Atoi(os.Getenv("EXPIRATION_TIME"))
			//数据10天过期的写法
			//err := rdb.Set(c.Request.Context(), fmt.Sprintf("item:%s", item.ID), itemJSON, 10*2400*time.Minute).Err()
			err := rdb.Set(c.Request.Context(), fmt.Sprintf("item:%s", item.ID), itemJSON, time.Duration(exp)*time.Hour).Err()
			fmt.Println("rd3-set-up")
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "保存文章失败"})
				return
			}

			c.JSON(http.StatusCreated, item)
		})

		// 上传文件(照片/文档)
		api.POST("/files", func(c *gin.Context) {
			fileType := c.PostForm("type")
			if fileType != "photo" && fileType != "document" {
				c.JSON(http.StatusBadRequest, gin.H{"error": "文件类型必须是photo或document"})
				return
			}

			file, fileHeader, err := c.Request.FormFile("file") // ✅ 获取文件头（原始名）
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "未上传文件"})
				return
			}
			defer file.Close()

			originalName := fileHeader.Filename // ✅ 原始文件名（含后缀）
			ext := filepath.Ext(originalName)   // ✅ 获取后缀 .pdf/.doc/.ppt

			// 生成唯一文件名
			filename := fmt.Sprintf("%s-%d%s", fileType, time.Now().UnixNano(), ext)
			filepath := fmt.Sprintf("uploads/%s", filename)

			// 保存文件
			outFile, err := os.Create(filepath)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "保存文件失败"})
				return
			}
			defer outFile.Close()
			io.Copy(outFile, file)

			// 创建媒体项 👇 重点：保存 OriginalName
			item := MediaItem{
				ID:           fmt.Sprintf("%s-%d", fileType, time.Now().UnixNano()),
				Type:         fileType,
				Title:        originalName,
				Content:      filename,
				OriginalName: originalName, // ✅ 存储原始文件名
				CreatedAt:    time.Now(),
			}

			itemJSON, _ := json.Marshal(item)
			err = rdb.Set(c.Request.Context(), fmt.Sprintf("item:%s", item.ID), itemJSON, 0).Err()
			if err != nil {
				os.Remove(filepath)
				c.JSON(http.StatusInternalServerError, gin.H{"error": "保存媒体项失败"})
				return
			}

			c.JSON(http.StatusCreated, item)

		})

		// 下载文件
		api.GET("/files/:id", func(c *gin.Context) {
			id := c.Param("id")
			key := fmt.Sprintf("item:%s", id)

			data, err := rdb.Get(c.Request.Context(), key).Result()
			if err != nil {
				c.JSON(http.StatusNotFound, gin.H{"error": "文件不存在"})
				return
			}

			var item MediaItem
			if err := json.Unmarshal([]byte(data), &item); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "解析文件信息失败"})
				return
			}

			if item.Type == "article" {
				c.JSON(http.StatusBadRequest, gin.H{"error": "文章不能通过此接口下载"})
				return
			}

			filepath := fmt.Sprintf("uploads/%s", item.Content)
			if _, err := os.Stat(filepath); os.IsNotExist(err) {
				c.JSON(http.StatusNotFound, gin.H{"error": "物理文件不存在"})
				return
			}

			// ✅ 核心修复：指定下载时的原始文件名
			c.Header("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, item.OriginalName))
			c.File(filepath)
		})

	}

	// 启动服务器
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("服务器启动在端口 %s", port)
	log.Fatal(r.Run(":" + port))
}
