package main

import "fmt"

// Sms is a concrete implementation of Otp for sending OTP via SMS
type Sms struct {
	Otp
}

// genRandomOTP generates a random OTP of given length
func (s *Sms) genRandomOTP(len int) string {
	randomOTP := "1234"
	fmt.Printf("SMS: generating random otp %s\n", randomOTP)
	return randomOTP
}

// saveOTPCache saves the OTP to cache (simulated here with a print statement)
func (s *Sms) saveOTPCache(otp string) {
	fmt.Printf("SMS: saving otp: %s to cache\n", otp)
}

// getMessage constructs the message to be sent with the OTP
func (s *Sms) getMessage(otp string) string {
	return "SMS OTP for login is " + otp
}

// sendNotification sends the OTP message via SMS (simulated here with a print statement)
func (s *Sms) sendNotification(message string) error {
	fmt.Printf("SMS: sending sms: %s\n", message)
	return nil
}
