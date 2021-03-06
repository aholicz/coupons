package main

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/spf13/viper"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

type Coupon struct {
	Coupon string `gorm:"column:coupon"`
}

var start time.Time

func init() {
	start = time.Now()
}

func init() {
	viper.SetConfigName("config")
	viper.AddConfigPath(".")

	if err := viper.ReadInConfig(); err != nil {
		log.Panicf("fatal error config file: %s", err)
	}

	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
}

func main() {
	var (
		cap         = 1700000
		coupons     = make([]*Coupon, 0, cap)
		couponCache = make(map[string]struct{})

		batchLimit = 3000

		db = newGormDB()
	)
	defer closeDB(db)

	for i := 0; i < cap; i++ {
		id := strings.Replace(uuid.New().String(), "-", "", -1)[:10]
		if v, ok := couponCache[id]; ok {
			fmt.Println("duplicated coupon code: ", v)
			cap++
			continue
		}

		couponCache[id] = struct{}{}
		coupons = append(coupons, &Coupon{Coupon: id})
	}

	if len(coupons) != 0 {
		for i := 0; i < len(coupons); i += batchLimit {
			couponChuck := coupons[i:min(i+batchLimit, len(coupons))]
			if err := db.Table(viper.GetString("mysql.table.coupon")).CreateInBatches(couponChuck, cap).Error; err != nil {
				panic(err)
			}
		}
	}

	fmt.Printf("insert %d coupons time used: %s", len(coupons), time.Since(start).String())
}

func min(a, b int) int {
	if a <= b {
		return a
	}
	return b
}

func newGormDB() *gorm.DB {
	dsn := fmt.Sprintf(
		"%s:%s@tcp(%s)/%s?checkConnLiveness=false&loc=Local&parseTime=true&readTimeout=%s&timeout=%s&writeTimeout=%s&maxAllowedPacket=0",
		viper.GetString("mysql.username"),
		viper.GetString("mysql.password"),
		viper.GetString("mysql.host")+":"+viper.GetString("mysql.port"),
		viper.GetString("mysql.database"),
		viper.GetString("mysql.timeout"),
		viper.GetString("mysql.timeout"),
		viper.GetString("mysql.timeout"),
	)

	db, err := gorm.Open(mysql.Open(dsn), nil)
	if err != nil {
		log.Panicf("cannot open mysql connection:%s", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		log.Panic(err)
	}

	sqlDB.SetMaxIdleConns(viper.GetInt("mysql.maxconns"))
	sqlDB.SetMaxOpenConns(viper.GetInt("mysql.maxconns"))
	sqlDB.SetConnMaxLifetime(viper.GetDuration("mysql.maxlifetime"))

	return db
}

func closeDB(db *gorm.DB) {
	sqlDB, _ := db.DB()
	sqlDB.Close()
}
