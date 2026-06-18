package command

import (
	"be-lms/config"
	"be-lms/database/db"
	redisperm "be-lms/redis"
	"fmt"

	"github.com/joho/godotenv"
)

func RefreshPermissionsCacheCommand() {
	_ = godotenv.Load()
	cfg := config.LoadConfig()
	if err := db.ConnectPostgres(cfg); err != nil {
		panic(err)
	}
	_ = db.ConnectRedis(cfg)

	for _, id := range []int{1, 2, 3} {
		if err := redisperm.NewRoleRedis(id).ClearRolePermissionsCache(); err != nil {
			fmt.Printf("role %d: %v\n", id, err)
			continue
		}
		fmt.Printf("refreshed permissions cache for role %d\n", id)
	}
}
