package handler

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v3"
)

// ValidationError merepresentasikan kumpulan error validasi per-field.
type ValidationError struct {
	Errors map[string]string
}

func (v *ValidationError) Error() string {
	if len(v.Errors) == 0 {
		return "validasi gagal"
	}
	var messages []string
	for field, msg := range v.Errors {
		messages = append(messages, fmt.Sprintf("%s: %s", field, msg))
	}
	return strings.Join(messages, "; ")
}

// BindAndValidate mem-binding payload (JSON atau Form) dan melakukan validasi tag struct boundary.
// Tag yang didukung:
// - `validate:"required"`: field tidak boleh kosong (string non-empty, int != 0, slice non-empty, pointer != nil)
// - `validate:"min=X"`: panjang string / slice minimal X, atau nilai numerik minimal X
// - `validate:"max=X"`: panjang string / slice maksimal X, atau nilai numerik maksimal X
// - `validate:"oneof=A B C"`: nilai string harus salah satu dari pilihan yang diberikan
//
// Return value:
// - Mengembalikan true jika validasi berhasil (data valid).
// - Mengembalikan false jika binding atau validasi gagal, dan response HTTP (400 atau 422) sudah otomatis dikirim ke client via SendError / SendValidationError.
//
// Cara penggunaan di handler:
//
//	var req CreateRequest
//	if ok := BindAndValidate(c, &req); !ok {
//	    return nil
//	}
//	// Lanjutkan proses service...
func BindAndValidate(c fiber.Ctx, out interface{}) bool {
	// 1. Binding body
	contentType := string(c.Get("Content-Type"))
	if strings.Contains(contentType, "application/json") {
		if err := c.Bind().JSON(out); err != nil {
			_ = SendError(c, fiber.StatusBadRequest, "Format request tidak valid", "Invalid JSON payload")
			return false
		}
	} else {
		// Default fallback atau form / body binding
		if err := c.Bind().Body(out); err != nil {
			_ = SendError(c, fiber.StatusBadRequest, "Format request tidak valid", "Invalid body payload")
			return false
		}
	}

	// 2. Validasi struct boundary
	errs := ValidateStruct(out)
	if len(errs) > 0 {
		_ = SendValidationError(c, errs)
		return false
	}

	return true
}

// SendValidationError mengirim HTTP 422 Unprocessable Entity dengan rincian field error di property details.
func SendValidationError(c fiber.Ctx, fieldErrors map[string]string) error {
	return SendError(c, fiber.StatusUnprocessableEntity, "Validasi input gagal", fieldErrors)
}

// ValidateStruct memvalidasi pointer struct berdasarkan tag `validate` dan `json`.
func ValidateStruct(s interface{}) map[string]string {
	errs := make(map[string]string)
	val := reflect.ValueOf(s)

	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}

	if val.Kind() != reflect.Struct {
		return errs
	}

	typ := val.Type()
	for i := 0; i < val.NumField(); i++ {
		fieldVal := val.Field(i)
		fieldType := typ.Field(i)

		// Jangan periksa unexported field
		if !fieldType.IsExported() {
			continue
		}

		validateTag := fieldType.Tag.Get("validate")
		if validateTag == "" {
			continue
		}

		jsonTag := fieldType.Tag.Get("json")
		fieldName := strings.Split(jsonTag, ",")[0]
		if fieldName == "" || fieldName == "-" {
			fieldName = strings.ToLower(fieldType.Name)
		}

		rules := strings.Split(validateTag, ",")
		for _, rule := range rules {
			rule = strings.TrimSpace(rule)
			if rule == "" {
				continue
			}

			if rule == "required" {
				if isZero(fieldVal) {
					errs[fieldName] = fmt.Sprintf("%s wajib diisi", fieldName)
					break // Jika required gagal, hentikan pengecekan rule lain untuk field ini
				}
			} else if strings.HasPrefix(rule, "min=") {
				minStr := strings.TrimPrefix(rule, "min=")
				minVal, _ := strconv.Atoi(minStr)
				if !checkMin(fieldVal, minVal) {
					switch fieldVal.Kind() {
					case reflect.String:
						errs[fieldName] = fmt.Sprintf("%s minimal %d karakter", fieldName, minVal)
					case reflect.Slice, reflect.Array, reflect.Map:
						errs[fieldName] = fmt.Sprintf("%s minimal %d item", fieldName, minVal)
					case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
						errs[fieldName] = fmt.Sprintf("%s minimal bernilai %d", fieldName, minVal)
					}
				}
			} else if strings.HasPrefix(rule, "max=") {
				maxStr := strings.TrimPrefix(rule, "max=")
				maxVal, _ := strconv.Atoi(maxStr)
				if !checkMax(fieldVal, maxVal) {
					switch fieldVal.Kind() {
					case reflect.String:
						errs[fieldName] = fmt.Sprintf("%s maksimal %d karakter", fieldName, maxVal)
					case reflect.Slice, reflect.Array, reflect.Map:
						errs[fieldName] = fmt.Sprintf("%s maksimal %d item", fieldName, maxVal)
					case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
						errs[fieldName] = fmt.Sprintf("%s maksimal bernilai %d", fieldName, maxVal)
					}
				}
			} else if strings.HasPrefix(rule, "oneof=") {
				optionsStr := strings.TrimPrefix(rule, "oneof=")
				options := strings.Fields(optionsStr)
				if fieldVal.Kind() == reflect.String {
					strVal := fieldVal.String()
					if strVal != "" {
						matched := false
						for _, opt := range options {
							if strings.EqualFold(strVal, opt) {
								matched = true
								break
							}
						}
						if !matched {
							errs[fieldName] = fmt.Sprintf("%s harus salah satu dari: %s", fieldName, strings.Join(options, ", "))
						}
					}
				}
			}
		}
	}

	return errs
}

func isZero(v reflect.Value) bool {
	switch v.Kind() {
	case reflect.String:
		return strings.TrimSpace(v.String()) == ""
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return v.Int() == 0
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return v.Uint() == 0
	case reflect.Float32, reflect.Float64:
		return v.Float() == 0
	case reflect.Slice, reflect.Map, reflect.Array:
		return v.Len() == 0
	case reflect.Ptr, reflect.Interface:
		return v.IsNil()
	case reflect.Bool:
		return false // Bool false bukan berarti zero-value yang invalid secara default
	default:
		return v.IsZero()
	}
}

func checkMin(v reflect.Value, min int) bool {
	switch v.Kind() {
	case reflect.String:
		if v.String() == "" {
			return true // Jika kosong dan bukan required, lewatkan pengecekan min
		}
		return len(strings.TrimSpace(v.String())) >= min
	case reflect.Slice, reflect.Array, reflect.Map:
		return v.Len() >= min
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return v.Int() >= int64(min)
	default:
		return true
	}
}

func checkMax(v reflect.Value, max int) bool {
	switch v.Kind() {
	case reflect.String:
		return len(v.String()) <= max
	case reflect.Slice, reflect.Array, reflect.Map:
		return v.Len() <= max
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return v.Int() <= int64(max)
	default:
		return true
	}
}
