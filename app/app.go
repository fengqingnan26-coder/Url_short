package app

import (
	"context"
	"fmt"
	"log"

	// "net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	// "github.com/facebookgo/grace/gracehttp"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/jekyulll/url_shortener/config"
	"github.com/jekyulll/url_shortener/database"
	"github.com/jekyulll/url_shortener/internal/api"
	"github.com/jekyulll/url_shortener/internal/cache"
	"github.com/jekyulll/url_shortener/internal/repository"
	"github.com/jekyulll/url_shortener/internal/service"
	"github.com/jekyulll/url_shortener/pkg/email"
	"github.com/jekyulll/url_shortener/pkg/filter"
	"github.com/jekyulll/url_shortener/pkg/hasher"
	"github.com/jekyulll/url_shortener/pkg/jwt"
	"github.com/jekyulll/url_shortener/pkg/randnum"
	"github.com/jekyulll/url_shortener/pkg/shortcode"
	"gorm.io/gorm"
)

type Application struct {
	r           *gin.Engine
	db          *gorm.DB
	redisCache  *cache.RedisCache
	jwt         *jwt.JWT
	cfg         *config.Config
	urlService  *service.URLService
	urlHandler  *api.URLHandler
	userHandler *api.UserHandler
}

func New() *Application {
	return &Application{}
}

func (a *Application) Init(configPath string) error {
	// config
	cfg, err := config.NewFromFile(configPath)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}
	a.cfg = cfg
	log.Println("init: config loaded")

	// db
	a.db, err = database.NewDB(cfg.Database)
	if err != nil {
		return err
	}
	log.Println("init: db connected")

	// redis
	redisCache, err := cache.NewRedisCache(cfg.Redis)
	if err != nil {
		return err
	}
	a.redisCache = redisCache
	log.Println("init: redis connected")

	// pkg
	emailSender, err := email.NewEmailSend(cfg.Email)
	if err != nil {
		return err
	}
	log.Println("init: email initialized")

	passwordHash := hasher.NewPassworkHash()

	a.jwt = jwt.NewJWT(cfg.JWT)

	randNum := randnum.NewRandNum(cfg.RandNum)

	generator := shortcode.NewShortCodeGeneratorImpl(cfg.ShortCode.Length)

	// cuntomValidator := validator.NewCustomValidator()

	filter := filter.New(a.cfg.Filter.Capacity, a.cfg.Filter.ErrorRate)

	urlRepo := repository.NewURLRepo(a.db)
	userRepo := repository.NewUserRepo(a.db)

	a.urlService = service.NewURLService(urlRepo, filter, generator, redisCache, cfg.App)
	userService := service.NewUserService(userRepo, passwordHash, a.jwt, redisCache, emailSender, randNum)

	a.urlHandler = api.NewURLHandler(a.urlService)
	a.userHandler = api.NewUserHandler(userService)

	// TODO
	// TimeOut未设置
	// Log中间件未设置
	// CORS未设置
	// Recover未设置

	// r := gin.Default()
	// // r.Use()
	// r.POST("/api/url", a.urlHandler.CreateURL)
	// r.GET(":code", a.urlHandler.RedirectURL)
	// r.GET("/", a.urlHandler.DefaultURL)
	// a.r = r

	a.r = gin.Default()
	// 允许所有跨域请求
	a.r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:3000"}, // 前端地址
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))
	a.initRouter()

	return nil
}

func (a *Application) Run() {
	go a.start()
	go a.tickSyncViewsToDB()
	go a.tickCleanUp()

	a.shutDown()
}

func (a *Application) start() {
	if err := a.r.Run(a.cfg.Server.Addr); err != nil {
		log.Println(err)
	}
	// 开机时清理一次过期url
	go func() {
		if err := a.urlService.DeleteAllExpired(context.Background()); err != nil {
			log.Println(err)
		}
	}()
	// 用 gracehttp 代替 gin 的 Run, 自动优雅退出
	// ! gracehttp 只能在 Linux 上用

	// if err := gracehttp.Serve(&http.Server{
	// 	Addr:         a.cfg.Server.Addr,
	// 	Handler:      a.e,
	// 	WriteTimeout: a.cfg.Server.WriteTimeout,
	// 	ReadTimeout:  a.cfg.Server.ReadTimeout,
	// }); err != nil {
	// 	fmt.Println(err)
	// }
}

func (a *Application) tickCleanUp() {
	ticker := time.NewTicker(a.cfg.App.CleanupInterval)
	defer ticker.Stop()
	for range ticker.C {
		func() {
			// 5min 超时取消
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
			defer cancel()

			// 加分布式锁
			lockKey := "lock:cleanup"
			lockValue, ok, err := a.redisCache.AcquireLock(ctx, lockKey, 5*time.Minute)
			if err != nil || !ok {
				log.Println("cleanup skipped: lock not acquired")
				return
			}
			defer a.redisCache.ReleaseLock(ctx, lockKey, lockValue)

			if err := a.urlService.DeleteAllExpired(ctx); err != nil {
				log.Println(err)
			}
		}()
	}
}

func (a *Application) tickSyncViewsToDB() {
	ticker := time.NewTicker(a.cfg.App.SyncViewDuration)
	defer ticker.Stop()

	for range ticker.C {
		func() {
			// 5min 超时取消
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
			defer cancel()

			// 加分布式锁
			lockKey := "lock:sync_views"
			lockValue, ok, err := a.redisCache.AcquireLock(ctx, lockKey, 5*time.Minute)
			if err != nil || !ok {
				log.Println("sync_views skipped: lock not acquired")
				return
			}
			defer a.redisCache.ReleaseLock(ctx, lockKey, lockValue)

			if err := a.urlService.SyncViewsToDB(ctx); err != nil {
				log.Printf("failed to SyncViewsToDB: %v", err.Error())
			}
		}()
	}
}

// func (a *Application) tickRebuildFilter() {
// 	ticker := time.NewTicker(24 * time.Hour)
// 	defer ticker.Stop()
// 	for range ticker.C {
// 		if err := a.repo.RebuildBloomFilter(context.Background()); err != nil {
// 			// TODO 层级设计错误
// 		}
// 	}
// }

func (a *Application) shutDown() {
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	defer func() {
		sqlDB, err := a.db.DB()
		if err != nil {
			log.Println(err)
		}
		if err := sqlDB.Close(); err != nil {
			log.Println(err)
		}
	}()

	defer func() {
		if err := a.redisCache.Close(); err != nil {
			log.Println(err)
		}
	}()

	// TODO
	// gin 的优雅退出

	// 5s时间退出
	_, cancle := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancle()
}
