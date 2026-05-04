package notifications

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/sesv2"
	"github.com/aws/aws-sdk-go-v2/service/sesv2/types"
	"github.com/mi-michi/backend/internal/db"
	"github.com/mi-michi/backend/internal/vaccines"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// RunDailyVaccineReminders busca vacunas próximas (≤7 días) y envía emails via SES.
func RunDailyVaccineReminders(ctx context.Context) {
	upcoming, err := vaccines.ListUpcoming(ctx, 7)
	if err != nil {
		log.Printf("notifications: error listando vacunas: %v", err)
		return
	}
	if len(upcoming) == 0 {
		log.Println("notifications: no hay vacunas próximas")
		return
	}

	for _, v := range upcoming {
		if v.NextDueDate == nil {
			continue
		}
		email, name := getUserEmailAndName(ctx, v.UserID)
		if email == "" {
			continue
		}
		catName := getCatName(ctx, v.CatID)
		if err := sendVaccineEmailSES(ctx, email, name, catName, v.Name, *v.NextDueDate); err != nil {
			log.Printf("notifications: error enviando email a %s: %v", email, err)
		} else {
			log.Printf("notifications: email SES enviado a %s para vacuna %s", email, v.Name)
		}
	}
}

func getUserEmailAndName(ctx context.Context, userID string) (string, string) {
	oid, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return "", ""
	}
	var result struct {
		Email       string `bson:"email"`
		DisplayName string `bson:"display_name"`
	}
	if err := db.Col("users").FindOne(ctx, bson.M{"_id": oid}).Decode(&result); err != nil {
		return "", ""
	}
	return result.Email, result.DisplayName
}

func getCatName(ctx context.Context, catID string) string {
	oid, err := primitive.ObjectIDFromHex(catID)
	if err != nil {
		return "tu gato"
	}
	var result struct {
		Name string `bson:"name"`
	}
	if err := db.Col("cats").FindOne(ctx, bson.M{"_id": oid}).Decode(&result); err != nil {
		return "tu gato"
	}
	return result.Name
}

// sendVaccineEmailSES envía el email usando AWS SES v2.
func sendVaccineEmailSES(ctx context.Context, to, userName, catName, vaccineName, dueDate string) error {
	from := os.Getenv("SES_FROM_EMAIL")
	region := os.Getenv("SES_REGION")
	accessKey := os.Getenv("S3_ACCESS_KEY") // reutilizamos las mismas credenciales IAM
	secretKey := os.Getenv("S3_SECRET_KEY")

	if from == "" {
		return fmt.Errorf("SES_FROM_EMAIL no configurado")
	}
	if region == "" {
		region = "us-east-1"
	}

	cfg, err := config.LoadDefaultConfig(ctx,
		config.WithRegion(region),
		config.WithCredentialsProvider(
			credentials.NewStaticCredentialsProvider(accessKey, secretKey, ""),
		),
	)
	if err != nil {
		return fmt.Errorf("error configurando AWS SES: %w", err)
	}

	client := sesv2.NewFromConfig(cfg)

	subject := fmt.Sprintf("🐾 Recordatorio: vacuna de %s", catName)
	htmlBody := fmt.Sprintf(`
<!DOCTYPE html>
<html>
<body style="font-family:sans-serif;max-width:480px;margin:0 auto;padding:24px">
  <div style="background:linear-gradient(135deg,#E879A0,#A78BFA);border-radius:16px;padding:24px;text-align:center;margin-bottom:24px">
    <h1 style="color:white;margin:0;font-size:28px">🐾 Mi Michi</h1>
  </div>
  <h2>Hola %s 👋</h2>
  <p>Te recordamos que <strong>%s</strong> tiene pendiente la vacuna:</p>
  <div style="background:#F3F4F6;border-radius:12px;padding:16px;margin:16px 0">
    <p style="margin:0;font-size:18px;font-weight:bold">💉 %s</p>
    <p style="margin:4px 0 0;color:#6B7280">Fecha: %s</p>
  </div>
  <p>¡No olvides agendar la cita con tu veterinario! 🩺</p>
  <p style="color:#9CA3AF;font-size:12px;margin-top:32px">— El equipo de Mi Michi 🐱</p>
</body>
</html>
`, userName, catName, vaccineName, dueDate)

	textBody := fmt.Sprintf(
		"Hola %s, recordatorio: %s tiene la vacuna %s el %s. ¡Agenda la cita con tu vet!",
		userName, catName, vaccineName, dueDate,
	)

	_, err = client.SendEmail(ctx, &sesv2.SendEmailInput{
		FromEmailAddress: aws.String(from),
		Destination: &types.Destination{
			ToAddresses: []string{to},
		},
		Content: &types.EmailContent{
			Simple: &types.Message{
				Subject: &types.Content{
					Data:    aws.String(subject),
					Charset: aws.String("UTF-8"),
				},
				Body: &types.Body{
					Html: &types.Content{
						Data:    aws.String(htmlBody),
						Charset: aws.String("UTF-8"),
					},
					Text: &types.Content{
						Data:    aws.String(textBody),
						Charset: aws.String("UTF-8"),
					},
				},
			},
		},
	})
	return err
}

// StartDailyJob lanza el job de notificaciones cada 24 horas.
func StartDailyJob() {
	go func() {
		time.Sleep(10 * time.Second)
		RunDailyVaccineReminders(context.Background())

		ticker := time.NewTicker(24 * time.Hour)
		defer ticker.Stop()
		for range ticker.C {
			RunDailyVaccineReminders(context.Background())
		}
	}()
	log.Println("📬 Job de notificaciones SES iniciado (cada 24h)")
}
