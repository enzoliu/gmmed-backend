package validator

import "github.com/go-playground/validator/v10"

func Struct(s interface{}) error {
	return DefaultValidator.Struct(s)
}

func Register(register ...*ValidationRegister) {
	DefaultValidator.Register(register...)
}

var DefaultValidator = New()

type Validator struct {
	*validator.Validate
}

func New() *Validator {
	v := &Validator{
		Validate: validator.New(validator.WithRequiredStructEnabled()),
	}

	v.Register(
		CustomUUIDValidation(),
		NullableUUIDValidation(),
		NullableMaxValidation(),
		NullableMinValidation(),
		NullableLteValidation(),
		NullableGteValidation(),
		NullableHttpUrlValidation(),
	)

	return v
}

func (v *Validator) Register(registers ...*ValidationRegister) {
	for _, r := range registers {
		v.RegisterValidation(r.Tag, r.Func, r.CallValidationEvenIfNull)
	}
}

type ValidationRegister struct {
	Tag                      string
	Func                     validator.Func
	CallValidationEvenIfNull bool
}
