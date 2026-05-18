package i18n

import (
	"be-Clever School/config"
	_ "embed"
	"encoding/json"
	"reflect"
	"strings"
	"unicode"

	basei18n "github.com/nicksnyder/go-i18n/v2/i18n"
	"golang.org/x/text/language"
)

//go:embed languages/vi/translations.vi.json
var viTranslations []byte

//go:embed languages/en/translations.en.json
var enTranslations []byte

var Bundle *basei18n.Bundle
var currentLocale = "en"

type LocaleProvider interface {
	CurrentLocale() string
	DefaultLocale() string
}

type GlobalLocaleProvider struct{}

func (GlobalLocaleProvider) CurrentLocale() string { return GetCurrentLocale() }
func (GlobalLocaleProvider) DefaultLocale() string { return GetDefaultLocale() }

func Init(locale string, files ...string) error {
	if len(files) == 0 {
		return InitEmbedded(locale)
	}

	langTag := language.Make(locale)
	Bundle = basei18n.NewBundle(langTag)
	Bundle.RegisterUnmarshalFunc("json", json.Unmarshal)

	for _, file := range files {
		if _, err := Bundle.LoadMessageFile(file); err != nil {
			return err
		}
	}

	currentLocale = locale
	return nil
}

func InitEmbedded(locale string) error {
	Bundle = basei18n.NewBundle(language.Vietnamese)
	Bundle.RegisterUnmarshalFunc("json", json.Unmarshal)

	if _, err := Bundle.ParseMessageFileBytes(viTranslations, "translations.vi.json"); err != nil {
		return err
	}

	if _, err := Bundle.ParseMessageFileBytes(enTranslations, "translations.en.json"); err != nil {
		return err
	}

	config.Log.Info("Init embedded - Bundle locale")
	return nil
}

func GetDefaultLocale() string {
	return "vi"
}

func GetCurrentLocale() string {
	return currentLocale
}

func SetCurrentLocale(locale string) {
	currentLocale = locale
}

func Localize(messageID string, data ...map[string]interface{}) string {
	localizer := basei18n.NewLocalizer(Bundle, currentLocale)

	config := &basei18n.LocalizeConfig{
		MessageID: messageID,
	}

	if len(data) > 0 && data[0] != nil {
		config.TemplateData = data[0]
	}

	msg, err := localizer.Localize(config)
	if err != nil {
		parts := strings.Split(messageID, ".")
		key := parts[len(parts)-1]
		readable := strings.ReplaceAll(key, "_", " ")
		readable = strings.Title(readable)
		return readable
	}

	return msg
}

// PickField selects a string field from a struct by i18nKey and current locale.
// Example: with fields tagged `i18nKey:"name" locale:"vi"` and `i18nKey:"name" locale:"en"`.
// defaultLocale optional, e.g., "vi"
func PickField(model interface{}, key string, defaultLocale ...string) string {
	loc := strings.ToLower(GetCurrentLocale())
	return PickFieldByLocale(model, key, loc, defaultLocale...)
}

// PickFieldByLocale is like PickField but allows passing a specific locale
func PickFieldByLocale(model interface{}, key string, locale string, defaultLocale ...string) string {
	v := reflect.Indirect(reflect.ValueOf(model))
	if v.Kind() != reflect.Struct {
		return ""
	}

	t := v.Type()

	var fallbackVal string
	var hasFallback bool

	locale = strings.ToLower(locale)
	var preferredFallback string
	if len(defaultLocale) > 0 {
		preferredFallback = strings.ToLower(defaultLocale[0])
	} else {
		preferredFallback = "vi"
	}

	// First pass: exact locale match and collect fallback
	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)
		if f.Tag.Get("i18nKey") != key {
			continue
		}
		fieldLocale := strings.ToLower(f.Tag.Get("locale"))
		val := v.Field(i)
		if val.Kind() != reflect.String {
			continue
		}
		s := val.String()
		if fieldLocale == locale && s != "" {
			return s
		}
		if !hasFallback && fieldLocale == preferredFallback && s != "" {
			fallbackVal = s
			hasFallback = true
		}
	}

	if hasFallback {
		return fallbackVal
	}

	// Second pass: return first non-empty value of the key
	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)
		if f.Tag.Get("i18nKey") != key {
			continue
		}
		val := v.Field(i)
		if val.Kind() == reflect.String && val.String() != "" {
			return val.String()
		}
	}
	return ""
}

// FillTranslation populates dest struct's string fields by matching their snake_case names
// to i18nKey on the model. Example: field Name -> key "name", FullName -> "full_name".
// dest must be a pointer to struct with string fields.
func FillTranslation(model interface{}, dest interface{}, locale string, defaultLocale ...string) error {
	dv := reflect.ValueOf(dest)
	if dv.Kind() != reflect.Ptr || dv.IsNil() {
		return nil
	}
	sv := dv.Elem()
	if sv.Kind() != reflect.Struct {
		return nil
	}

	for i := 0; i < sv.NumField(); i++ {
		field := sv.Field(i)
		if !field.CanSet() || field.Kind() != reflect.String {
			continue
		}
		key := toSnakeCase(sv.Type().Field(i).Name)
		val := PickFieldByLocale(model, key, locale, defaultLocale...)
		field.SetString(val)
	}
	return nil
}

// FillTranslationAuto uses global current and default locale
func FillTranslationAuto(model interface{}, dest interface{}) error {
	return FillTranslation(model, dest, GetCurrentLocale(), GetDefaultLocale())
}

// FillTranslationWithProvider uses a provided LocaleProvider
func FillTranslationWithProvider(model interface{}, dest interface{}, provider LocaleProvider) error {
	if provider == nil {
		provider = GlobalLocaleProvider{}
	}
	return FillTranslation(model, dest, provider.CurrentLocale(), provider.DefaultLocale())
}

func toSnakeCase(s string) string {
	var b strings.Builder
	for i, r := range s {
		if unicode.IsUpper(r) {
			if i > 0 {
				b.WriteByte('_')
			}
			b.WriteRune(unicode.ToLower(r))
		} else {
			b.WriteRune(r)
		}
	}
	return b.String()
}

// ====== WRITE HELPERS (generic, reusable across models) ======

// SetFieldsByLocale sets i18n-tagged fields on modelPtr for a given locale using the provided map.
// modelPtr must be a pointer to struct. values map key is the i18nKey (snake_case).
// Returns number of fields successfully set.
func SetFieldsByLocale(modelPtr interface{}, values map[string]string, locale string) int {
	mv := reflect.ValueOf(modelPtr)
	if mv.Kind() != reflect.Ptr || mv.IsNil() {
		return 0
	}
	sv := mv.Elem()
	if sv.Kind() != reflect.Struct {
		return 0
	}

	t := sv.Type()
	locale = strings.ToLower(locale)
	updated := 0

	for key, val := range values {
		// find field with matching i18nKey and locale
		for i := 0; i < t.NumField(); i++ {
			f := t.Field(i)
			if f.Tag.Get("i18nKey") != key {
				continue
			}
			if strings.ToLower(f.Tag.Get("locale")) != locale {
				continue
			}
			field := sv.Field(i)
			if field.CanSet() && field.Kind() == reflect.String {
				field.SetString(val)
				updated++
				break
			}
		}
	}
	return updated
}

// SetFieldsWithProvider sets fields for provider.CurrentLocale().
func SetFieldsWithProvider(modelPtr interface{}, values map[string]string, provider LocaleProvider) int {
	if provider == nil {
		provider = GlobalLocaleProvider{}
	}
	return SetFieldsByLocale(modelPtr, values, provider.CurrentLocale())
}

// SetFieldsAuto sets fields for the global current locale.
func SetFieldsAuto(modelPtr interface{}, values map[string]string) int {
	return SetFieldsByLocale(modelPtr, values, GetCurrentLocale())
}

// SetFieldsWithProviderFallback sets fields for current locale and ensures default locale has a value if empty.
// Only fills default locale if the existing string field is empty to avoid overwriting user-provided data.
func SetFieldsWithProviderFallback(modelPtr interface{}, values map[string]string, provider LocaleProvider) int {
	if provider == nil {
		provider = GlobalLocaleProvider{}
	}
	updated := SetFieldsByLocale(modelPtr, values, provider.CurrentLocale())

	// Fill default locale if empty
	mv := reflect.ValueOf(modelPtr)
	if mv.Kind() != reflect.Ptr || mv.IsNil() {
		return updated
	}
	sv := mv.Elem()
	if sv.Kind() != reflect.Struct {
		return updated
	}
	t := sv.Type()
	defaultLocale := strings.ToLower(provider.DefaultLocale())

	for key, val := range values {
		// locate default-locale field for this key
		var defIdx = -1
		var curIdx = -1
		for i := 0; i < t.NumField(); i++ {
			f := t.Field(i)
			if f.Tag.Get("i18nKey") != key {
				continue
			}
			tagLoc := strings.ToLower(f.Tag.Get("locale"))
			if tagLoc == defaultLocale {
				defIdx = i
			}
			if tagLoc == strings.ToLower(provider.CurrentLocale()) {
				curIdx = i
			}
		}
		if defIdx >= 0 {
			field := sv.Field(defIdx)
			if field.CanSet() && field.Kind() == reflect.String && field.Len() == 0 {
				// only set if empty
				// prefer current-locale value if set; else use incoming val
				var v = val
				if curIdx >= 0 {
					curField := sv.Field(curIdx)
					if curField.Kind() == reflect.String && curField.Len() > 0 {
						v = curField.String()
					}
				}
				field.SetString(v)
				updated++
			}
		}
	}
	return updated
}
