package utils

import (
	"fmt"
	"math/rand"
	"time"
)

func GenerateLoanCode() string {
	now := time.Now()
	timePart := now.Format("20060102")
	randomPart := rand.Intn(900) + 100

	return fmt.Sprintf("LOA-%s-%d", timePart, randomPart)
}

func GeneratePaymentCode() string {
	now := time.Now()
	timePart := now.Format("20060102")
	randomPart := rand.Intn(900) + 100

	return fmt.Sprintf("PAY-%s-%d", timePart, randomPart)
}
