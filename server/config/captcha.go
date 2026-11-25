package config

import "fmt"

const (
	DefaultCaptchaKeyLong            = 6
	DefaultCaptchaImgWidth           = 240
	DefaultCaptchaImgHeight          = 80
	DefaultCaptchaOpenCaptchaTimeout = 3600
)

type Captcha struct {
	KeyLong            int `mapstructure:"key-long" json:"key-long" yaml:"key-long"`                                     // 验证码长度
	ImgWidth           int `mapstructure:"img-width" json:"img-width" yaml:"img-width"`                                  // 验证码宽度
	ImgHeight          int `mapstructure:"img-height" json:"img-height" yaml:"img-height"`                               // 验证码高度
	OpenCaptcha        int `mapstructure:"open-captcha" json:"open-captcha" yaml:"open-captcha"`                         // 防爆破验证码开启此数，0代表每次登录都需要验证码，其他数字代表错误密码次数，如3代表错误三次后出现验证码
	OpenCaptchaTimeOut int `mapstructure:"open-captcha-timeout" json:"open-captcha-timeout" yaml:"open-captcha-timeout"` // 防爆破验证码超时时间，单位：s(秒)
}

// Normalize fills omitted values while preserving an explicitly configured
// failure threshold. OpenCaptcha == 0 means every login requires a captcha.
func (c *Captcha) Normalize() {
	if c.KeyLong == 0 {
		c.KeyLong = DefaultCaptchaKeyLong
	}
	if c.ImgWidth == 0 {
		c.ImgWidth = DefaultCaptchaImgWidth
	}
	if c.ImgHeight == 0 {
		c.ImgHeight = DefaultCaptchaImgHeight
	}
	if c.OpenCaptchaTimeOut == 0 {
		c.OpenCaptchaTimeOut = DefaultCaptchaOpenCaptchaTimeout
	}
}

func (c Captcha) Validate() error {
	if c.KeyLong < 4 || c.KeyLong > 6 {
		return fmt.Errorf("验证码长度必须在 4 到 6 之间")
	}
	if c.ImgWidth < 80 || c.ImgWidth > 1000 {
		return fmt.Errorf("验证码图片宽度必须在 80 到 1000 之间")
	}
	if c.ImgHeight < 40 || c.ImgHeight > 400 {
		return fmt.Errorf("验证码图片高度必须在 40 到 400 之间")
	}
	if c.OpenCaptcha < 0 || c.OpenCaptcha > 10 {
		return fmt.Errorf("验证码触发次数必须在 0 到 10 之间")
	}
	if c.OpenCaptchaTimeOut < 1 || c.OpenCaptchaTimeOut > 86400 {
		return fmt.Errorf("验证码触发窗口必须在 1 到 86400 秒之间")
	}
	return nil
}

// Required reports whether the current login attempt must provide a captcha.
// The counter includes the initial request marker, so a positive threshold is
// intentionally crossed only after the configured number of failures.
func (c Captcha) Required(failures int) bool {
	return c.OpenCaptcha == 0 || failures > c.OpenCaptcha
}
