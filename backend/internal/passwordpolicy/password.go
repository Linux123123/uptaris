package passwordpolicy

import "github.com/go-passwd/validator"

const (
	MinLength = 8
	MaxLength = 128
)

type Error struct{ message string }

func (e *Error) Error() string { return e.message }

func (e *Error) ValidationMessage() string { return e.message }

// Regex rules reject passwords missing the required character category.
var policy = validator.New(
	validator.MinLength(MinLength, &Error{message: "Use 8-128 characters for your password"}),
	validator.MaxLength(MaxLength, &Error{message: "Use 8-128 characters for your password"}),
	validator.Regex(`^[^\p{Ll}]*$`, &Error{message: "Include a lowercase letter"}),
	validator.Regex(`^[^\p{Lu}]*$`, &Error{message: "Include an uppercase letter"}),
	validator.Regex(`^[^\p{N}]*$`, &Error{message: "Include a number"}),
	validator.Regex(`^[^\p{P}\p{S}]*$`, &Error{message: "Include a symbol"}),
)

func Validate(value string) error { return policy.Validate(value) }
