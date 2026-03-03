package main

import (
	"laporanharianapi/pkg/fcm"
	"log"
)

func main() {
	err := fcm.InitFirebase()
	if err != nil {
		log.Fatalf("❌ Error: %v", err)
	}
	log.Println("✅ Success testing FCM init")
}
