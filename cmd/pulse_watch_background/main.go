package main

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	trmRedis "github.com/avito-tech/go-transaction-manager/drivers/goredis8/v2"
	trmSqlx "github.com/avito-tech/go-transaction-manager/drivers/sqlx/v2"
	"github.com/avito-tech/go-transaction-manager/trm/v2"
	trmContext "github.com/avito-tech/go-transaction-manager/trm/v2/context"
	"github.com/avito-tech/go-transaction-manager/trm/v2/manager"
	"github.com/avito-tech/go-transaction-manager/trm/v2/settings"
	"github.com/elastic/go-elasticsearch/v9"
	"github.com/go-redis/redis/v8"
	"github.com/hibiken/asynq"
	"github.com/jmoiron/sqlx"
	"github.com/joho/godotenv"
	asynq_server "github.com/vekshinnikita/pulse_watch/internal/asynq"
	"github.com/vekshinnikita/pulse_watch/internal/config"
	"github.com/vekshinnikita/pulse_watch/internal/constants"
	es "github.com/vekshinnikita/pulse_watch/internal/dbs/elasticseach"
	"github.com/vekshinnikita/pulse_watch/internal/dbs/postgres"
	redis_db "github.com/vekshinnikita/pulse_watch/internal/dbs/redis"
	background_tasks_handler "github.com/vekshinnikita/pulse_watch/internal/handlers/background_tasks"
	"github.com/vekshinnikita/pulse_watch/internal/logger"
	"github.com/vekshinnikita/pulse_watch/internal/repository"
	kafka_repository "github.com/vekshinnikita/pulse_watch/internal/repository/kafka"
	"github.com/vekshinnikita/pulse_watch/internal/service"
	gin_validators "github.com/vekshinnikita/pulse_watch/internal/validators/gin"
)

func initConfig() {
	err := config.LoadConfig("configs/config.yaml")
	if err != nil {
		log.Fatalf("error while loading config: %v", err.Error())
	}

	if err := godotenv.Load(); err != nil {
		log.Fatalf("error while loading environment variables: %v", err.Error())
	}
}

func initLogger() {
	err := logger.InitLogger()
	if err != nil {
		log.Fatalf("error while init logger: %v", err.Error())
	}
}

func initDB() *sqlx.DB {
	slog.Debug("Initializing db")
	db, err := postgres.NewDB()
	if err != nil {
		msg := fmt.Sprintf("error occurred while initializing db: %s", err.Error())
		slog.Error(msg)
		log.Fatal(msg)
	}

	return db
}

func initElasticClient() *elasticsearch.TypedClient {
	slog.Debug("Initializing elasticsearch client")
	esClient, err := es.NewESClient()
	if err != nil {
		msg := fmt.Sprintf("error occurred while initializing elasticsearch client: %s", err.Error())
		slog.Error(msg)
		log.Fatal(msg)
	}

	return esClient
}

func initRedisClient() *redis.Client {
	slog.Debug("Initializing elasticsearch client")
	client, err := redis_db.NewClient()
	if err != nil {
		msg := fmt.Sprintf("error occurred while initializing redis client: %s", err.Error())
		slog.Error(msg)
		log.Fatal(msg)
	}

	return client
}

func initProducer() repository.MessageProducer {
	slog.Info("Init kafka producer")
	producer, err := kafka_repository.NewKafkaProducer()
	if err != nil {
		msg := fmt.Sprintf("init repository: %s", err.Error())
		slog.Error(msg)
		log.Fatal(msg)
	}
	slog.Info("Kafka producer is initialized")

	return producer
}

func initValidators() {
	err := gin_validators.InitValidators()
	if err != nil {
		msg := fmt.Sprintf("error occurred while initializing validators: %s", err.Error())
		slog.Error(msg)
		log.Fatal(msg)
	}
}

func getAsynqClientOpt() *asynq.RedisClientOpt {
	redisCfg := redis_db.GetConfig()

	return &asynq.RedisClientOpt{
		Addr:     fmt.Sprintf("%v:%v", redisCfg.Host, redisCfg.Port),
		DB:       1,
		Password: redisCfg.Password,
	}
}

func main() {
	initConfig()
	initLogger()
	initValidators()

	slog.Info("Starting background tasks consumers...")

	db := initDB()
	defer func() {
		slog.Info("Closing DB connection...")
		db.Close()
	}()

	producer := initProducer()
	defer producer.Close()

	esClient := initElasticClient()
	defer func() {
		slog.Info("Closing elastic search connection...")
		esClient.Close(context.Background())
	}()

	// Init Transaction Manager

	redisClient := initRedisClient()
	defer func() {
		slog.Info("Closing redis connection...")
		redisClient.Close()
	}()

	// Asyncq Option
	opt := getAsynqClientOpt()
	asynqClient := asynq.NewClient(opt)

	// Init Transaction Manager
	sqlxMng := manager.Must(trmSqlx.NewDefaultFactory(db))
	redisMng := manager.Must(
		trmRedis.NewDefaultFactory(redisClient),
		manager.WithSettings(trmRedis.MustSettings(
			settings.Must(
				settings.WithCtxKey(constants.RedisTrCtxKey),
			),
			trmRedis.WithMulti(false),
		)),
	)
	trManager := manager.MustChained([]trm.Manager{sqlxMng, redisMng})

	repos := repository.NewRepository(&repository.RepositoryParams{
		Database:     db,
		SQLTRMGetter: trmSqlx.DefaultCtxGetter,

		Producer: producer,
		ESClient: esClient,

		RedisClient:    redisClient,
		RedisTRMGetter: trmRedis.NewCtxGetter(trmContext.New(constants.RedisTrCtxKey)),
	})
	services := service.NewServices(repos, trManager)
	handler := background_tasks_handler.NewHandler(asynqClient, services)

	srv := asynq_server.NewServer(opt, asynqClient)
	go func() {
		if err := srv.Run(handler.InitTasks()); err != nil {
			msg := fmt.Sprintf("run asynq server: %s", err.Error())
			slog.Error(msg)
			log.Fatalf(msg)
		}
	}()
	defer srv.Stop()

	// Обработка сигнала завершения
	stopChan := make(chan os.Signal, 2)
	signal.Notify(stopChan, os.Interrupt, syscall.SIGTERM)

	// Ждем сигнала завершения
	<-stopChan
	slog.Info("Shutting down...")
}
