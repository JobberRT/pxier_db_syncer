package main

import (
	"fmt"
	nFormatter "github.com/antonfisher/nested-logrus-formatter"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	gormLogger "gorm.io/gorm/logger"
	"os"
	"runtime"
	"strings"
	"time"
)

type proxy struct {
	Id        int    `gorm:"primaryKey; autoIncrement" json:"id"`
	Address   string `json:"address"`
	Provider  string `json:"provider"`
	CreatedAt int64  `json:"-"`
	UpdatedAt int64  `json:"-"`
	ErrTimes  int    `json:"-"`
	DialType  string `json:"dial_type"`
}

func (p *proxy) TableName() string {
	return "proxy"
}

func init() {
	viper.AutomaticEnv()
	viper.AllowEmptyEnv(true)
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")

	err := viper.ReadInConfig()
	if err != nil {
		logrus.Panic(err)
		os.Exit(-1)
	}

	logrus.SetLevel(logrus.InfoLevel)
	logrus.SetOutput(os.Stdout)
	logrus.SetReportCaller(true)
	logrus.SetFormatter(&nFormatter.Formatter{
		NoColors:        false,
		HideKeys:        false,
		TimestampFormat: time.Stamp,
		CallerFirst:     true,
		CustomCallerFormatter: func(frame *runtime.Frame) string {
			filename := ""
			slash := strings.LastIndex(frame.File, "/")
			if slash >= 0 {
				filename = frame.File[slash+1:]
			}
			return fmt.Sprintf("「%s:%d」", filename, frame.Line)
		},
	})
}

func main() {
	masterUrl := viper.GetString("write_url")
	if len(masterUrl) == 0 {
		logrus.Panic("missing write_url")
	}
	slaveUrl := viper.GetString("read_url")
	if len(slaveUrl) == 0 {
		logrus.Panic("missing read_url")
	}

	masterDB := newMasterDB(masterUrl)
	slaveDB := newSlaveDB(slaveUrl)
	interval := viper.GetInt64("sync_interval")
	if interval == 0 {
		interval = 5
	}
	ticker := time.NewTicker(time.Duration(interval) * time.Second)
	defer ticker.Stop()
	for {
		logrus.Info("sync")
		temp := make([]*proxy, 0)
		if err := masterDB.Find(&temp).Error; err != nil {
			logrus.WithError(err).Error("failed to query master mysql")
			continue
		}
		slaveDB.Where("id > ?", "0").Delete(&proxy{})
		for _, each := range temp {
			slaveDB.Save(&each)
		}
		<-ticker.C
	}
}

func newMasterDB(url string) *gorm.DB {
	logrus.Info("start master mysql")
	db, err := gorm.Open(mysql.Open(url), &gorm.Config{
		Logger: gormLogger.Default.LogMode(gormLogger.Silent),
	})
	if err != nil {
		logrus.WithError(err).Panic("failed to create db")
	}
	if err := db.AutoMigrate(&proxy{}); err != nil {
		logrus.WithError(err).Panic("failed to migrate model")
	}
	d, _ := db.DB()
	d.SetMaxIdleConns(10)
	d.SetMaxOpenConns(100)
	d.SetConnMaxLifetime(time.Hour)
	return db
}

func newSlaveDB(url string) *gorm.DB {
	logrus.Info("start slave mysql")
	if len(url) == 0 {
		logrus.Panic("mysql url is empty")
	}
	db, err := gorm.Open(mysql.Open(url), &gorm.Config{
		Logger: gormLogger.Default.LogMode(gormLogger.Silent),
	})
	if err != nil {
		logrus.WithError(err).Panic("failed to create db")
	}
	if err := db.AutoMigrate(&proxy{}); err != nil {
		logrus.WithError(err).Panic("failed to migrate model")
	}
	d, _ := db.DB()
	d.SetMaxIdleConns(10)
	d.SetMaxOpenConns(100)
	d.SetConnMaxLifetime(time.Hour)
	return db
}
