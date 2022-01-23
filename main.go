package main

import (
	"fmt"
	"log"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var db *gorm.DB

func init() {
	var err error
	db, err = gorm.Open(postgres.Open("host=localhost user=postgres password=1234567890 dbname=postgres sslmode=disable port=5432"), &gorm.Config{})
	if err != nil {
		log.Fatalf("connect: %s", err.Error())
		return
	}
	sqlDB, err := db.DB()
	if err != nil {
		log.Fatalf("get db: %s", err.Error())
		return
	}

	// 連線最大數量
	sqlDB.SetMaxOpenConns(1)

	// if err := SetLockTimeout(5 * time.Second); err != nil {
	// 	log.Fatalf("set lock timeout: %s", err.Error())
	// 	return
	// }

	// // 連線並操作完成後(commit/rollback)，可以持續佔有連線的數量 => 避免我們的服務連線釋放後，無用的連線馬上被其他服務搶光
	sqlDB.SetMaxIdleConns(2)
	// // 連線並操作完成後(commit/rollback)，還可以維持連線多久。如果設定 SetMaxIdleConns(0)，那 SetConnMaxLifetime 設定就沒用 => 避免我們的服務佔用無用的連線太久
	sqlDB.SetConnMaxLifetime(5 * time.Minute)
}

func main() {
	tx := db.Begin()
	defer tx.Rollback()
	fmt.Printf("tx start\n")

	time.Sleep(5 * time.Second)

	fmt.Printf("tx set effective time\n")
	if err := SetTxEffectiveTime(tx, 1*time.Second); err != nil {
		log.Fatalf("SetTxEffectiveTime: %s", err.Error())
		return
	}

	time.Sleep(5 * time.Second)
	fmt.Printf("tx wake\n")

	if err := tx.Commit().Error; err != nil {
		fmt.Printf("commit: %s\n", err.Error())
	}
	fmt.Printf("tx end\n")

	// ==============

	tx2 := db.Begin()
	defer tx2.Rollback()

	fmt.Printf("tx2 start\n")
	if err := SetTxEffectiveTime(tx2, 1*time.Second); err != nil {
		log.Fatalf("SetTxEffectiveTime: %s", err.Error())
		return
	}

	time.Sleep(3 * time.Second)
	fmt.Printf("tx2 wake\n")

	if err := tx2.Commit().Error; err != nil {
		fmt.Printf("commit: %s\n", err.Error())
	}
	fmt.Printf("tx2 end\n")

	// ==============

	tx3 := db.Begin()
	defer tx3.Rollback()

	fmt.Printf("tx3 start\n")
	time.Sleep(3 * time.Second)
	fmt.Printf("tx3 wake\n")

	if err := tx3.Commit().Error; err != nil {
		fmt.Printf("commit: %s\n", err.Error())
	}
	fmt.Printf("tx3 end\n")

	time.Sleep(2 * time.Minute)

	// var wg sync.WaitGroup
	// for i := 1; i <= 50; i++ {
	// 	wg.Add(1)
	// 	go func(i int) {
	// 		defer wg.Done()
	// 		tx := db.Begin()
	// 		if err := tx.Error; err != nil {
	// 			fmt.Printf("%d: begin err: %s\n", i, err.Error())
	// 		}

	// 		fmt.Printf("%d start\n", i)
	// 		time.Sleep(3 * time.Second)
	// 		fmt.Printf("%d wake\n", i)

	// 		if err := tx.Commit().Error; err != nil {
	// 			fmt.Printf("%d: commit err: %s\n", i, err.Error())
	// 		}
	// 		fmt.Printf("%d end\n", i)
	// 	}(i)
	// }
	// wg.Wait()

	// fmt.Println("all dona")
	// time.Sleep(1 * time.Minute)
}

// SetTxEffectiveTime set idle_in_transaction_session_timeout
func SetTxEffectiveTime(tx *gorm.DB, d time.Duration) error {
	sql := fmt.Sprintf("SET LOCAL idle_in_transaction_session_timeout = %d;", d.Milliseconds())
	return tx.Exec(sql).Error
}

// SetLockTimeout set lock_timeout
func SetLockTimeout(tx *gorm.DB, d time.Duration) error {
	sql := fmt.Sprintf("SET LOCAL lock_timeout = %d;", d.Milliseconds())
	return tx.Exec(sql).Error
}
