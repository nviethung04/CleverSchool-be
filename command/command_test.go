package command

import (
	"be-lms/config"
	"be-lms/database/db"
	"be-lms/redis"
	"flag"
	"fmt"
	"log"
	"testing"

	"github.com/joho/godotenv"
)

var functionName string
var name string

// go test ./command -v --function="create_user" --name="Linh"

func init() {
	flag.StringVar(&functionName, "function", "", "Function to test")
	flag.StringVar(&name, "name", "", "Test name (optional)")
}

func TestCommand(t *testing.T) {
	_ = godotenv.Load("../.env")

	cfg := config.LoadConfig()

	config.InitDisks()
	config.InitLogger()

	flag.Parse()

	if functionName == "" {
		t.Errorf("❌ Function name is required!")
		return
	}

	fmt.Printf("Function: %s\n", functionName)
	if name != "" {
		fmt.Printf("Name: %s\n", name)
	} else {
		fmt.Println("Name is not provided")
	}

	switch functionName {
	case "create_user":
		t.Run("CreateUserCommand", func(t *testing.T) {
			RunCreateUserCommand(t, name)
		})
	case "delete_user":
		t.Run("RunDeleteUserCommand", func(t *testing.T) {
			RunDeleteUserCommand(t, 1)
		})
	case "clear_cache":
		if err := db.ConnectRedis(cfg); err != nil {
			log.Fatal("❌ Redis connection failed:", err)
		}

		if name != "" {
			redis.ClearCacheByPrefix(name)
			// permissions:role
		} else {
			redis.ClearAllCache()
		}
	default:
		t.Errorf("❌ Unknown function: %s", functionName)
	}
}

func RunCreateUserCommand(t *testing.T, name string) {
	if name == "" {
		name = "Default User"
	}
	createCmd := CreateUserCommand{Title: name}
	result := createCmd.Execute()

	if !result {
		t.Errorf("❌ User creation failed, expected true, got: %v", result)
	} else {
		t.Logf("✅ Test Passed: User %s was successfully created", name)
	}
}

func RunDeleteUserCommand(t *testing.T, id int) {
	createCmd := DeleteUserCommand{ID: id}
	result := createCmd.Execute()

	if !result {
		t.Errorf("❌ User delete failed, expected true, got: %v", result)
	} else {
		t.Logf("✅ Test Passed: User %s was successfully deleted", name)
	}
}
