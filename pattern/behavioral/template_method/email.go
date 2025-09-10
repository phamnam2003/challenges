package main

import "fmt"

// Email is a concrete implementation of OTP for email
type Email struct {
	Otp
}

// genRandomOTP generates a random OTP
func (s *Email) genRandomOTP(len int) string {
	randomOTP := "1234"
	fmt.Printf("EMAIL: generating random otp %s\n", randomOTP)
	return randomOTP
}

// saveOTPCache saves the OTP to cache
func (s *Email) saveOTPCache(otp string) {
	fmt.Printf("EMAIL: saving otp: %s to cache\n", otp)
}

// getMessage returns the message to be sent
func (s *Email) getMessage(otp string) string {
	return "EMAIL OTP for login is " + otp
}

// sendNotification sends the notification
func (s *Email) sendNotification(message string) error {
	fmt.Printf("EMAIL: sending email: %s\n", message)
	return nil
}
