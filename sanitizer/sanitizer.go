// Traverses an arbitrary struct and sanitizes strings that match expected name or value patterns.
// Largely copied from https://gist.github.com/hvoecking/10772475 (MIT License)

package sanitizer

import (
	"net/mail"
	"net/url"
	"reflect"
	"regexp"
	"strings"
	"unsafe"
)

const DefaultSanitizePlaceholder = "----"

type SanitizeFunc func(name string, value string, customSanitizer SanitizeFunc) (handled bool, sanitizedValue string)

func Sanitize(obj interface{}, customSanitizer SanitizeFunc) interface{} {
	if obj == nil {
		return nil
	}

	// Wrap the original in a reflect.Value
	original := reflect.ValueOf(obj)

	copy := reflect.New(original.Type()).Elem()
	sanitizeRecursive("", copy, original, customSanitizer)

	// Remove the reflection wrapper
	return copy.Interface()
}

//note: these patterns should all be lowercased
var sanitizedNamePatterns = []*regexp.Regexp{
	regexp.MustCompile(`^.*password.*$`),
	regexp.MustCompile(`^secret.*`),
	regexp.MustCompile(`^token$`),
	regexp.MustCompile(`^pwd$`),
	regexp.MustCompile(`^pass$`),
	regexp.MustCompile(`^p$`),
	regexp.MustCompile(`^cert(ificate)?$`),
	regexp.MustCompile(`^cred(ential)?s?$`),
	regexp.MustCompile(`^database$`),
	regexp.MustCompile(`^database_url$`),
	regexp.MustCompile(`^db$`),
	regexp.MustCompile(`^db_url$`),
	regexp.MustCompile(`^token$`),
	regexp.MustCompile(`^username$`),
	regexp.MustCompile(`^last\s+name$`),
}

var urlPrefixes = []string{
	"http://",
	"https://",
	"ftp:",
	"file:",
	"mailto:",
	"postgres://",
	"mongodb://",
	"redis://",
}

// TODO: detect circular structures.
func sanitizeRecursive(name string, copy reflect.Value, original reflect.Value, customSanitizer SanitizeFunc) {
	if !copy.CanSet() || original.IsZero() {
		return
	}

	switch original.Kind() {
	// The first cases handle nested structures and translate them recursively

	// If it is a pointer we need to unwrap and call once again
	case reflect.Ptr:
		// To get the actual value of the original we have to call Elem()
		// At the same time this unwraps the pointer so we don't end up in
		// an infinite recursion
		originalValue := original.Elem()
		// Check if the pointer is nil
		if !originalValue.IsValid() {
			return
		}
		// Allocate a new object and set the pointer to it
		copy.Set(reflect.New(originalValue.Type()))

		copyElem := copy.Elem()

		// Unwrap the newly created pointer
		sanitizeRecursive(name, copyElem, originalValue, customSanitizer)

	// If it is an interface (which is very similar to a pointer), do basically the
	// same as for the pointer. Though a pointer is not the same as an interface so
	// note that we have to call Elem() after creating a new object because otherwise
	// we would end up with an actual pointer
	case reflect.Interface:
		// Get rid of the wrapping interface
		originalValue := original.Elem()
		// Create a new object. Now new gives us a pointer, but we want the value it
		// points to, so we have to call Elem() to unwrap it
		copyValue := reflect.New(originalValue.Type()).Elem()
		sanitizeRecursive(name, copyValue, originalValue, customSanitizer)
		copy.Set(copyValue)

	// If it is a struct we translate each field
	case reflect.Struct:
		for i := 0; i < original.NumField(); i += 1 {
			fieldName := copy.Type().Field(i).Name
			copyField := copy.Field(i)
			origField := original.Field(i)

			if !copyField.CanSet() {
				copyField = makeSettable(copyField)
				origField = makeSettable(origField)
			}
			sanitizeRecursive(fieldName, copyField, origField, customSanitizer)
		}

	// If it is a slice we create a new slice and translate each element
	case reflect.Slice:
		copy.Set(reflect.MakeSlice(original.Type(), original.Len(), original.Cap()))
		for i := 0; i < original.Len(); i += 1 {
			sanitizeRecursive(name, copy.Index(i), original.Index(i), customSanitizer)
		}

	// If it is a map we create a new map and translate each value
	case reflect.Map:
		copy.Set(reflect.MakeMap(original.Type()))
		for _, key := range original.MapKeys() {
			var itemName string
			if key.Type().Kind() == reflect.String {
				itemName = key.Interface().(string)
			} else {
				itemName = ""
			}

			originalValue := original.MapIndex(key)
			// New gives us a pointer, but again we want the value
			copyValue := reflect.New(originalValue.Type()).Elem()
			sanitizeRecursive(itemName, copyValue, originalValue, customSanitizer)
			copy.SetMapIndex(key, copyValue)
		}

	// Otherwise we cannot traverse anywhere so this finishes the the recursion
	case reflect.String:
		// If it is a string, sanitize it
		anon := sanitizeString(name, original.String(), customSanitizer)
		copy.SetString(anon)

	// And everything else will simply be taken from the original
	default:
		copy.Set(original)
	}
}

func makeSettable(v reflect.Value) reflect.Value {
	return reflect.NewAt(v.Type(), unsafe.Pointer(v.UnsafeAddr())).Elem()
}

func sanitizeString(name string, value string, customSanitizer SanitizeFunc) string {

	if customSanitizer != nil {
		handled, sanitizedValue := customSanitizer(name, value, customSanitizer)
		if handled {
			// with the custom sanitizer, we return immediately if handled.
			// this allows implementers to override default behavior if needed.
			return sanitizedValue
		}
	}

	if handled, sanitized := sanitizeUrl(name, value, customSanitizer); handled {
		value = sanitized
	}

	if handled, sanitized := sanitizeEmailAddress(name, value, customSanitizer); handled {
		value = sanitized
	}

	if handled, sanitized := sanitizeByNamePattern(name, value, customSanitizer); handled {
		value = sanitized
	}

	return value
}

// if this value can be parsed as a url, sanitize its parts, then reassemble them into a url string.
func sanitizeUrl(name string, value string, customSanitizer SanitizeFunc) (handled bool, sanitized string) {
	parsed, u := tryParseUrl(value)
	if !parsed {
		return false, ""
	}

	// sanitize user:password if present
	if u.User != nil {
		u.User = url.User(DefaultSanitizePlaceholder)
	}

	// sanitize query
	if len(u.RawQuery) > 0 {
		sanitizedQuery := Sanitize(u.Query(), customSanitizer).(url.Values)
		u.RawQuery = sanitizedQuery.Encode()
	}

	// sanitize other url components
	anonU := Sanitize(u, customSanitizer).(*url.URL)

	return true, anonU.String()
}

func tryParseUrl(value string) (parsed bool, u *url.URL) {
	if hasUrlPrefix(value) {
		u, err := url.Parse(value)
		return err == nil, u
	}

	return false, nil
}

func hasUrlPrefix(value string) bool {
	for _, pfx := range urlPrefixes {
		if strings.HasPrefix(value, pfx) {
			return true
		}
	}
	return false
}

// sanitize if we can successfully parse the value as a list of email addresses
func sanitizeEmailAddress(name string, value string, customSanitizer SanitizeFunc) (handled bool, sanitized string) {
	parsed, _ := tryParseEmailAddress(value)
	if !parsed {
		return false, ""
	}
	return true, DefaultSanitizePlaceholder
}

func tryParseEmailAddress(value string) (parsed bool, email string) {
	if strings.Index(value, "@") > 0 {
		addrs, err := mail.ParseAddressList(value)
		if err == nil && len(addrs) > 0 {
			return true, addrs[0].Address
		}
	}
	return false, ""
}

// sanitize if the name matches
func sanitizeByNamePattern(name string, value string, customSanitizer SanitizeFunc) (handled bool, sanitized string) {
	lcName := strings.ToLower(name)
	for _, re := range sanitizedNamePatterns {
		if re.Match([]byte(lcName)) {
			return true, DefaultSanitizePlaceholder
		}
	}
	return false, ""
}
