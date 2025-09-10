package main

import "fmt"

// SecurityCode is a struct that holds the security code
type SecurityCode struct {
	code int
}

// newSecurityCode is a constructor for SecurityCode
func newSecurityCode(code int) *SecurityCode {
	return &SecurityCode{
		code: code,
	}
}

// checkCode checks if the incoming code matches the security code
func (s *SecurityCode) checkCode(incomingCode int) error {
	if s.code != incomingCode {
		return fmt.Errorf("security Code is incorrect")
	}
	fmt.Println("SecurityCode Verified")
	return nil
}
