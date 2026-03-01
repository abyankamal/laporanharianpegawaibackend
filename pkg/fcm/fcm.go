package fcm

import (
	"context"
	"errors"
	"log"
	"os"

	firebase "firebase.google.com/go/v4"
	"firebase.google.com/go/v4/messaging"
	"google.golang.org/api/option"
)

var (
	app    *firebase.App
	client *messaging.Client
)

// InitFirebase menginisialisasi Firebase Admin SDK.
// Pastikan file serviceAccountKey.json ada di root directory.
/*
Cara Mendapatkan serviceAccountKey.json:
1. Buka Firebase Console (https://console.firebase.google.com/).
2. Pilih proyek Anda.
3. Klik icon gear (Project Settings) -> tab Service accounts.
4. Pilih Firebase Admin SDK dan pastikan Go terpilih.
5. Klik "Generate new private key".
6. Ganti nama file JSON yang terunduh menjadi serviceAccountKey.json dan pindahkan ke root project.
*/
func InitFirebase() error {
	serviceAccountPath := "serviceAccountKey.json"

	// Cek apakah file credentials ada
	if _, err := os.Stat(serviceAccountPath); os.IsNotExist(err) {
		log.Printf("⚠️  Warning: File %s tidak ditemukan. Fitur Push Notification FCM belum aktif.", serviceAccountPath)
		return nil // Return nil agar aplikasi tidak crash jika file belum disiapkan
	}

	opt := option.WithCredentialsFile(serviceAccountPath)

	var err error
	app, err = firebase.NewApp(context.Background(), nil, opt)
	if err != nil {
		log.Printf("❌ Error inisialisasi Firebase App: %v", err)
		return err
	}

	client, err = app.Messaging(context.Background())
	if err != nil {
		log.Printf("❌ Error inisialisasi Firebase Messaging Client: %v", err)
		return err
	}

	log.Println("✅ Firebase Admin SDK berhasil diinisialisasi")
	return nil
}

// SendPushNotification mengirimkan push notification ke device menggunakan FCM Token.
func SendPushNotification(fcmToken string, title string, body string) error {
	if client == nil {
		return errors.New("firebase messaging client belum diinisialisasi")
	}

	if fcmToken == "" {
		return errors.New("fcm token kosong")
	}

	message := &messaging.Message{
		Notification: &messaging.Notification{
			Title: title,
			Body:  body,
		},
		Token: fcmToken,
	}

	// Mengirim pesan ke perangkat yang sesuai dengan token
	response, err := client.Send(context.Background(), message)
	if err != nil {
		log.Printf("❌ Gagal mengirim notifikasi (Token: %s): %v", fcmToken, err)
		return err
	}

	log.Printf("✅ Notifikasi FCM terkirim (Response: %v)", response)
	return nil
}
