package main

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/UzStack/paycue/internal/config"
	"github.com/UzStack/paycue/internal/domain"
	"github.com/UzStack/paycue/internal/http/routes"
	"github.com/UzStack/paycue/internal/infra"
	"github.com/UzStack/paycue/internal/repository"
	"github.com/UzStack/paycue/internal/usecase"
	"github.com/joho/godotenv"
	_ "github.com/mattn/go-sqlite3"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

var (
	VERSION = "1.0.11"
)

func author() {
	fmt.Println("Fullname      Azamov Samandar")
	fmt.Println("Telegram      https://t.me/Azamov_Samandar")
	fmt.Println("Github        https://github.com/JscorpTech")
	fmt.Println("Phone         +998(88)-811-23-09")
	os.Exit(0)
}

func version() {
	fmt.Printf("Version: %s\n", VERSION)
	os.Exit(0)
}

func printHelp() {
	fmt.Println("Usage paycue [options]")
	fmt.Println("Options:")
	fmt.Println("  --help       -h       Show this help message")
	fmt.Println("  --version    -v       Show version")
	fmt.Println("  --telegram   -t       Telegram accountni ulash")
	fmt.Println("  --author     -a       Author of the project")
	os.Exit(0)
}

func main() {
	if err := godotenv.Load(".env"); err != nil {
		panic(".env file not loaded: " + err.Error())
	}

	var log *zap.Logger
	if os.Getenv("DEBUG") == "true" {
		logCfg := zap.NewDevelopmentConfig()
		logCfg.Level = zap.NewAtomicLevelAt(zap.InfoLevel)
		log, _ = logCfg.Build()
	} else {
		writer := zapcore.AddSync(&lumberjack.Logger{
			Filename:   "logs/app.log",
			MaxSize:    10,   // MB da (fayl 10 MB bo‘lsa rotate qiladi)
			MaxBackups: 5,    // necha eski faylni saqlash
			MaxAge:     30,   // kunlarda saqlash muddati
			Compress:   true, // eski fayllarni .gz qiladi
		})
		encoderCfg := zap.NewProductionEncoderConfig()
		encoderCfg.EncodeTime = zapcore.ISO8601TimeEncoder

		core := zapcore.NewCore(
			zapcore.NewJSONEncoder(encoderCfg),
			writer,
			zapcore.InfoLevel,
		)
		log = zap.New(core)
	}
	defer log.Sync()
	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "--author", "-a":
			author()
		case "--version", "-v":
			version()
		case "--help", "-h":
			printHelp()
		case "--telegram", "-t":
			if err := infra.Mtproto(context.Background(), nil, zap.NewNop(), 0, false, nil); err != nil {
				panic(err)
			}
			fmt.Println("")
			fmt.Println("Account qo'shildi")
			os.Exit(0)
		default:
			fmt.Println("Invalid options:", os.Args[1])
			os.Exit(0)
		}
	}

	cfg, err := config.NewConfig()
	if err != nil {
		panic(err)
	}

	db, err := sql.Open("sqlite3", "./db.sqlite3")
	if err != nil {
		panic(err)
	}
	defer db.Close()
	repository.InitTables(db)
	tasks := make(chan domain.Task, 10)

	mux := http.NewServeMux()
	routes.InitRoutes(mux, db, log, tasks, cfg)

	srv := &http.Server{
		Addr:    ":" + cfg.Port,
		Handler: mux,
	}
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	if err := usecase.InitWorker(ctx, log, tasks, cfg, db); err != nil {
		log.Error("worker init failed", zap.Any("error", err.Error()))
	}
	defer close(tasks)

	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal("server failed", zap.Error(err))
		}
	}()
	go func() {
		if err := infra.Mtproto(ctx, db, log, cfg.WatchID, true, tasks); err != nil {
			log.Fatal("server failed", zap.Error(err))
		}
	}()

	log.Info("server started", zap.String("addr", srv.Addr))

	<-ctx.Done()
	log.Info("shutting down...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Error("server shutdown failed", zap.Error(err))
	} else {
		log.Info("server exited properly")
	}
}
