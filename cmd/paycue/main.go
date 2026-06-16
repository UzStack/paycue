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
	"github.com/UzStack/paycue/internal/http/middleware"
	"github.com/UzStack/paycue/internal/http/routes"
	"github.com/UzStack/paycue/internal/repository"
	"github.com/UzStack/paycue/internal/telegram"
	"github.com/UzStack/paycue/internal/usecase"
	"github.com/joho/godotenv"
	_ "github.com/mattn/go-sqlite3"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

var VERSION = "dev" // build vaqtida ldflags orqali tag'dan to'ldiriladi

func author() {
	fmt.Println("Fullname      Azamov Samandar")
	fmt.Println("Telegram      https://t.me/Azamov_Samandar")
	fmt.Println("Github        https://github.com/UzStack")
	os.Exit(0)
}

func printHelp() {
	fmt.Println("Usage: paycue [options]")
	fmt.Println("Options:")
	fmt.Println("  --help     -h   Yordam")
	fmt.Println("  --version  -v   Versiya")
	fmt.Println("  --author   -a   Muallif")
	fmt.Println("")
	fmt.Println("Telegram account, carta va webhook'lar API orqali sozlanadi (paycue-cli ga qarang).")
	os.Exit(0)
}

func newLogger(debug bool) *zap.Logger {
	if debug {
		cfg := zap.NewDevelopmentConfig()
		cfg.Level = zap.NewAtomicLevelAt(zap.DebugLevel)
		l, _ := cfg.Build()
		return l
	}
	writer := zapcore.AddSync(&lumberjack.Logger{
		Filename:   "logs/app.log",
		MaxSize:    10,
		MaxBackups: 5,
		MaxAge:     30,
		Compress:   true,
	})
	encoderCfg := zap.NewProductionEncoderConfig()
	encoderCfg.EncodeTime = zapcore.ISO8601TimeEncoder
	core := zapcore.NewCore(zapcore.NewJSONEncoder(encoderCfg), writer, zapcore.InfoLevel)
	return zap.New(core)
}

func main() {
	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "--author", "-a":
			author()
		case "--version", "-v":
			fmt.Printf("Version: %s\n", VERSION)
			os.Exit(0)
		case "--help", "-h":
			printHelp()
		}
	}

	if err := godotenv.Load(".env"); err != nil {
		panic(".env file not loaded: " + err.Error())
	}

	cfg, err := config.NewConfig()
	if err != nil {
		panic(err)
	}

	log := newLogger(cfg.Debug)
	defer log.Sync()

	db, err := sql.Open("sqlite3", cfg.DBPath)
	if err != nil {
		panic(err)
	}
	defer db.Close()
	repository.InitTables(db)

	tasks := make(chan domain.Task, 100)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	if err := usecase.InitWorker(ctx, log, tasks, cfg, db); err != nil {
		log.Error("worker init failed", zap.Error(err))
	}
	defer close(tasks)

	tgManager := telegram.NewManager(cfg, db, log, tasks)
	tgManager.RestoreWatchers(ctx)

	mux := http.NewServeMux()
	routes.InitRoutes(mux, db, log, cfg, tgManager)

	srv := &http.Server{Addr: ":" + cfg.Port, Handler: middleware.CORS(cfg.CORSOrigin, mux)}
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
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
