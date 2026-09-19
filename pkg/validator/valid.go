package validator

import (
	"time"

	"github.com/go-playground/validator/v10"
)

type CustomValidator struct {
	validator *validator.Validate
}

func NewCustomValidator() *CustomValidator {
	v := validator.New()
	// 验证"after"标签对应的时间是否晚于当前时间
	v.RegisterValidation("after", func(fl validator.FieldLevel) bool {
		startTime, ok := fl.Field().Interface().(time.Time)
		if !ok {
			return false
		}
		return startTime.After(time.Now())
	})
	return &CustomValidator{
		validator: v,
	}
}

func (c *CustomValidator) Validate(i interface{}) error {
	return c.validator.Struct(i)
}
