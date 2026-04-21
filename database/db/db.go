package db

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"be-lms/config"
	"be-lms/observer"

	"github.com/redis/go-redis/v9"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var (
	MasterDB    *gorm.DB
	ReplicaDB   *gorm.DB
	RedisClient *redis.Client
	Ctx         = context.Background()
)

func ConnectPostgres(cfg config.Config) error {
	var err error

	// Connect Postgres Master
	MasterDB, err = gorm.Open(postgres.Open(cfg.DBMasterURL), &gorm.Config{})
	if err != nil {
		log.Println("Failed to connect to master database:", err)
		return err
	}
	masterSql, err := MasterDB.DB()
	if err != nil {
		log.Println("Failed to get master database:", err)
		return err
	}
	masterSql.SetMaxOpenConns(20)
	masterSql.SetMaxIdleConns(5)
	masterSql.SetConnMaxLifetime(time.Minute * 10)
	log.Println("Connected to master database")

	// Connect Postgres Replica nếu có
	if cfg.DBReplicaURL != "" {
		ReplicaDB, err = gorm.Open(postgres.Open(cfg.DBReplicaURL), &gorm.Config{})
		if err != nil {
			log.Println("⚠️ Failed to connect to replica database, fallback to use master database for replica:", err)
			ReplicaDB = MasterDB
			return nil
		}
		replicaSql, err := ReplicaDB.DB()
		if err != nil {
			log.Println("⚠️ Failed to get replica database sql.DB, fallback to use master database for replica:", err)
			ReplicaDB = MasterDB
			return nil
		}
		replicaSql.SetMaxOpenConns(20)
		replicaSql.SetMaxIdleConns(5)
		replicaSql.SetConnMaxLifetime(time.Minute * 10)
		log.Println("Connected to replica database")
	}

	return nil
}

func ConnectRedis(cfg config.Config) error {
	if !cfg.RedisEnabled {
		log.Println("⚠️ Redis is disabled, skipping Redis connection")
		return nil
	}

	RedisClient = redis.NewClient(&redis.Options{
		Addr:         fmt.Sprintf("%s:%s", cfg.RedisHost, cfg.RedisPort),
		Password:     cfg.RedisPassword,
		DB:           cfg.RedisDB,
		PoolSize:     100,
		MinIdleConns: 10,
		PoolTimeout:  60 * time.Second,
		DialTimeout:  20 * time.Second,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
	})

	_, err := RedisClient.Ping(Ctx).Result()
	if err != nil {
		log.Println("❌ Failed to connect to Redis:", err)
		return err
	}

	log.Println("✅ Connected to Redis")
	return nil
}

func CloseRedis() {
	if RedisClient != nil {
		if err := RedisClient.Close(); err != nil {
			log.Println("⚠️ Error closing Redis connection:", err)
		} else {
			log.Println("Redis connection closed")
		}
	}
}

var registeredCallbacks = make(map[string]bool)
var mu sync.Mutex

func registerCallbacksForModel(db *gorm.DB, modelName string) {
	mu.Lock()
	defer mu.Unlock()

	if registeredCallbacks[modelName] {
		return
	}
	registeredCallbacks[modelName] = true

	var obsModel observer.Observer

	switch modelName {
	case "user":
		obsModel = &observer.UserObserver{}
	case "question":
		obsModel = &observer.QuestionObserver{}
	case "media":
		obsModel = &observer.MediaObserver{}
	case "role":
		obsModel = &observer.RoleObserver{}
	default:
		obsModel = nil
	}

	if obsModel != nil {
		db.Callback().Create().Before("gorm:create").Register(modelName+"_before_create", func(tx *gorm.DB) {
			err := obsModel.BeforeCreate(tx.Statement.Dest, tx)
			if err != nil {
				tx.AddError(err)
			}
		})

		db.Callback().Create().After("gorm:create").Register(modelName+"_after_create", func(tx *gorm.DB) {
			err := obsModel.AfterCreate(tx.Statement.Dest, tx)
			if err != nil {
				tx.AddError(err)
			}
		})

		db.Callback().Update().Before("gorm:update").Register(modelName+"_before_update", func(tx *gorm.DB) {
			err := obsModel.BeforeUpdate(tx.Statement.Dest, tx)
			if err != nil {
				tx.AddError(err)
			}
		})
		db.Callback().Update().After("gorm:update").Register(modelName+"_after_update", func(tx *gorm.DB) {
			err := obsModel.AfterUpdate(tx.Statement.Dest, tx)
			if err != nil {
				tx.AddError(err)
			}
		})
		db.Callback().Delete().Before("gorm:delete").Register(modelName+"_before_delete", func(tx *gorm.DB) {
			err := obsModel.BeforeDelete(tx.Statement.Dest, tx)
			if err != nil {
				tx.AddError(err)
			}
		})
		db.Callback().Delete().After("gorm:delete").Register(modelName+"_after_delete", func(tx *gorm.DB) {
			err := obsModel.AfterDelete(tx.Statement.Dest, tx)
			if err != nil {
				tx.AddError(err)
			}
		})
	}
}

var registerCallbacksOnce sync.Once

func RegisterGormCallbacks(db *gorm.DB) {
	// Todo add role
	registerCallbacksOnce.Do(func() {
		registerCallbacksForModel(db, "user")
		registerCallbacksForModel(db, "question")
		registerCallbacksForModel(db, "media")
		registerCallbacksForModel(db, "role")
	})
}

func TestPostgresConnection() error {
	var count int64

	// Test MasterDB
	if MasterDB == nil {
		config.Log.Error("❌ MasterDB is nil")
		return fmt.Errorf("master database not initialized")
	}

	err := MasterDB.Table("users").Count(&count).Error
	if err != nil {
		config.Log.Error("❌ Failed to count users on MasterDB:", err)
		return err
	}
	config.Log.Infof("✅ MasterDB: users table has %d rows", count)

	// Test ReplicaDB (nếu có)
	if ReplicaDB != nil {
		count = 0
		err := ReplicaDB.Table("users").Count(&count).Error
		if err != nil {
			config.Log.Error("❌ Failed to count users on ReplicaDB:", err)
			return err
		}
		config.Log.Infof("✅ ReplicaDB: users table has %d rows", count)
	} else {
		config.Log.Warn("⚠️ ReplicaDB is nil, skipping replica check")
	}

	return nil
}
