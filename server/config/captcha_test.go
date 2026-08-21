package config

import "testing"

func TestCaptchaRequiredForEveryLogin(t *testing.T) {
	captcha := Captcha{OpenCaptcha: 0}
	if !captcha.Required(0) || !captcha.Required(99) {
		t.Fatal("open-captcha=0 must require a captcha for every login")
	}
}

func TestCaptchaRequiredAfterFailureThreshold(t *testing.T) {
	captcha := Captcha{OpenCaptcha: 3}
	if captcha.Required(3) {
		t.Fatal("captcha was required before the configured threshold was exceeded")
	}
	if !captcha.Required(4) {
		t.Fatal("captcha was not required after the configured threshold was exceeded")
	}
}

func TestCaptchaValidateRejectsUnsafeValues(t *testing.T) {
	tests := []Captcha{
		{KeyLong: 3, ImgWidth: 240, ImgHeight: 80, OpenCaptcha: 0, OpenCaptchaTimeOut: 3600},
		{KeyLong: 4, ImgWidth: 79, ImgHeight: 80, OpenCaptcha: 0, OpenCaptchaTimeOut: 3600},
		{KeyLong: 4, ImgWidth: 240, ImgHeight: 39, OpenCaptcha: 0, OpenCaptchaTimeOut: 3600},
		{KeyLong: 4, ImgWidth: 240, ImgHeight: 80, OpenCaptcha: -1, OpenCaptchaTimeOut: 3600},
		{KeyLong: 4, ImgWidth: 240, ImgHeight: 80, OpenCaptcha: 11, OpenCaptchaTimeOut: 3600},
		{KeyLong: 4, ImgWidth: 240, ImgHeight: 80, OpenCaptcha: 1, OpenCaptchaTimeOut: 0},
	}
	for _, captcha := range tests {
		if err := captcha.Validate(); err == nil {
			t.Fatalf("unsafe captcha configuration was accepted: %#v", captcha)
		}
	}
}

func TestCaptchaValidateAcceptsSupportedModes(t *testing.T) {
	for _, captcha := range []Captcha{
		{KeyLong: 4, ImgWidth: 240, ImgHeight: 80, OpenCaptcha: 0, OpenCaptchaTimeOut: 3600},
		{KeyLong: 6, ImgWidth: 320, ImgHeight: 100, OpenCaptcha: 3, OpenCaptchaTimeOut: 600},
	} {
		if err := captcha.Validate(); err != nil {
			t.Fatalf("valid captcha configuration was rejected: %v", err)
		}
	}
}
