package db

import (
	"correlator/config"
	"correlator/logger"
	"fmt"
	"math/rand"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var DB *gorm.DB

func Connect(cfg *config.DatabaseConfig) {
	var (
		err      error
		host     string = cfg.Host
		user     string = cfg.User
		password string = cfg.Password
		dbname   string = cfg.DBName
		port     int    = cfg.Port
		sslmode  string = cfg.SSLMode
	)

	//dsn := "host=localhost user=cor_admin password=2002 dbname=maindb port=5432 sslmode=disable"
	dsn := fmt.Sprintf("host=%v user=%v password=%v dbname=%v port=%v sslmode=%v",
		host, user, password, dbname, port, sslmode)

	DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		logger.ErrorLogger.Fatalf("Error in connecting to database: %v", err)
	}
}

func AutoMigrateAndStartDB(cfg *config.DatabaseConfig) {
	Connect(cfg)
	err := DB.AutoMigrate(User{})
	if err != nil {
		logger.ErrorLogger.Fatalf("Migration error: %v", err)
	}
	err = DB.AutoMigrate(Group{})
	if err != nil {
		logger.ErrorLogger.Fatalf("Migration error: %v", err)
	}
	err = DB.AutoMigrate(Rules{})
	if err != nil {
		logger.ErrorLogger.Fatalf("Migration error: %v", err)
	}
	err = DB.AutoMigrate(List{})
	if err != nil {
		logger.ErrorLogger.Fatalf("Migration error: %v", err)
	}

	if err = InitDefaultUser(); err != nil {
		logger.ErrorLogger.Printf("Failed to create default instances: %v", err)
	}
}

func InitDefaultUser() error {
	var users int64 = 0
	var groups int64 = 0
	DB.Model(&User{}).Count(&users)
	DB.Model(&Group{}).Count(&groups)
	if users == 0 || groups == 0 {
		perm := map[string]interface{}{"admin": true, "list": "rw", "rule": "rw"}
		g := Group{GroupName: "Default Admins", Description: "Default group for admins", IsActive: true, Permissions: perm}
		if err := DB.Create(&g).Error; err != nil {
			logger.ErrorLogger.Printf("Failed to create default admin group\n")
			return err
		}

		randompass := RandomPassword(16)
		u := User{Username: "DefCorAdm", Password: randompass, IsActive: true, ChangePass: false, Group: g}
		if err := u.HashPassword(); err != nil {
			logger.ErrorLogger.Printf("Failed to create default admin user\n")
			return err
		}
		if err := DB.Create(&u).Error; err != nil {
			logger.ErrorLogger.Printf("Failed to create default admin user\n")
			return err
		} else {
			fmt.Printf("Your default admin credentials are:\nUsername: %s\tPassword: %s\n", u.Username, randompass)
		}
	}
	return nil
}

func RandomPassword(length uint) string {
	seed := rand.NewSource(time.Now().UnixNano())
	rng := rand.New(seed)

	digits := "0123456789"
	specials := "~=+%^*/()[]{}/!@#$?|"
	all := "ABCDEFGHIJKLMNOPQRSTUVWXYZ" +
		"abcdefghijklmnopqrstuvwxyz" +
		digits + specials
	buf := make([]byte, length)
	buf[0] = digits[rng.Intn(len(digits))]
	buf[1] = specials[rng.Intn(len(specials))]
	for i := 2; i < int(length); i++ {
		buf[i] = all[rng.Intn(len(all))]
	}
	rng.Shuffle(len(buf), func(i, j int) {
		buf[i], buf[j] = buf[j], buf[i]
	})
	return string(buf)
}
