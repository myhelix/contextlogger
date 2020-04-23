package sanitizer

import (
	"fmt"
	"testing"

	"github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestSuiteSanitizer(t *testing.T) {
	RegisterFailHandler(ginkgo.Fail)
	RunSpecs(t, "Sanitizer Tests")
}

var _ = Describe("Sanitize Helpers", func() {

	Context("Object traversal", func() {
		type I interface{}

		type StringAlias string

		type A struct {
			Greeting string
			Message  string
			Pi       float64
			Password string
			Token    StringAlias
			Url      StringAlias
		}

		type B struct {
			Struct      A
			Ptr         *A
			Answer      int
			IntMap      map[int]string
			Map         map[string]string
			IfaceMap    map[string]interface{}
			StructMap   map[string]A
			StringSlice []string
			StructSlice []A
		}

		type Empty struct{}

		type testCase struct {
			value           interface{}
			expectedValue   interface{}
			customSanitizer SanitizeFunc
			description     string
		}

		testCases := []testCase{
			{
				description: "nil input should return nil",
				value:       nil,
			},
			{
				description: "nil pointer to struct should return nil",
				value:       func() *B { return nil }(),
			},
			{
				description: "nil pointer to interface should return nil",
				value:       func() *I { return nil }(),
			},
			{
				description:   "empty struct should return a copy of the empty struct",
				value:         Empty{},
				expectedValue: Empty{},
			},
			{
				description:   "struct with just zero-value fields should return an empty copy",
				value:         B{},
				expectedValue: B{},
			},
			{
				description:   "traverses and anonymize structs",
				value:         B{Struct: A{Greeting: "hello", Password: "secret", Pi: 3.14159}},
				expectedValue: B{Struct: A{Greeting: "hello", Password: DefaultSanitizePlaceholder, Pi: 3.14159}},
			},
			{
				description:   "traverses and anonymize structs wrapped in interface",
				value:         func() I { return &B{Struct: A{Greeting: "hello", Password: "secret", Pi: 3.14159}} }(),
				expectedValue: func() I { return &B{Struct: A{Greeting: "hello", Password: DefaultSanitizePlaceholder, Pi: 3.14159}} }(),
			},
			{
				description:   "traverses and anonymize struct pointers",
				value:         B{Ptr: &A{Greeting: "hello", Password: "secret", Pi: 3.14159}},
				expectedValue: B{Ptr: &A{Greeting: "hello", Password: DefaultSanitizePlaceholder, Pi: 3.14159}},
			},
			{
				description:   "traverses and anonymizes string maps",
				value:         B{Map: map[string]string{"hello": "world", "password": "secret"}},
				expectedValue: B{Map: map[string]string{"hello": "world", "password": DefaultSanitizePlaceholder}},
			},
			{
				description:   "leaves int maps unmodified",
				value:         B{IntMap: map[int]string{1: "hello", 2: "world"}},
				expectedValue: B{IntMap: map[int]string{1: "hello", 2: "world"}},
			},
			{
				description:   "traverses and anonymizes maps of structs",
				value:         B{IfaceMap: map[string]interface{}{"hello": "world", "password": "secret", "a": A{Password: "secret"}}},
				expectedValue: B{IfaceMap: map[string]interface{}{"hello": "world", "password": DefaultSanitizePlaceholder, "a": A{Password: DefaultSanitizePlaceholder}}},
			},
			{
				description:   "traverses and anonymizes maps of interfaces",
				value:         B{StructMap: map[string]A{"a1": {Password: "secret"}, "a2": {Greeting: "hello"}}},
				expectedValue: B{StructMap: map[string]A{"a1": {Password: DefaultSanitizePlaceholder}, "a2": {Greeting: "hello"}}},
			},
			{
				description:   "traverses and anonymizes string slices",
				value:         B{StringSlice: []string{"foo", "foo@bar.com", "bar"}},
				expectedValue: B{StringSlice: []string{"foo", DefaultSanitizePlaceholder, "bar"}},
			},
			{
				description:   "traverses and anonymizes struct slices",
				value:         B{StructSlice: []A{{Greeting: "hello", Password: "secret"}, {Greeting: "World", Password: "moo"}}},
				expectedValue: B{StructSlice: []A{{Greeting: "hello", Password: DefaultSanitizePlaceholder}, {Greeting: "World", Password: DefaultSanitizePlaceholder}}},
			},
			{
				description:   "anonymizes string type aliases",
				value:         A{Token: "sensitive", Url: "http://foo.bar?password=secret"},
				expectedValue: A{Token: DefaultSanitizePlaceholder, Url: "http://foo.bar?password=----"},
			},
		}

		for _, t := range testCases {
			func(t testCase) {
				It(t.description, func() {
					actualValue := Sanitize(t.value, t.customSanitizer)
					if t.expectedValue == nil {
						Expect(actualValue).To(BeNil())
					} else {
						Expect(actualValue).To(Equal(t.expectedValue))
					}
					// assert not the same instance
					Expect(&actualValue).ToNot(BeIdenticalTo(&t.value))
				})
			}(t)
		}
	})

	Context("Helpers", func() {

		Context("sanitizeEmailAddress()", func() {
			type testCase struct {
				value           string
				expectedValue   string
				expectedHandled bool
				description     string
			}

			testCases := []testCase{
				{description: "leave unhandled when not a well-formed email address", value: "foo"},
				{description: "leave unhandled when not a well-formed email address", value: "foo@"},
				{description: "leave unhandled when not a well-formed email address", value: "@foo"},
				{description: "leave unhandled when not a well-formed email address", value: "@foo"},
				{
					description:     "sanitize a single well-formed email address",
					value:           "foo@bar.com",
					expectedHandled: true,
					expectedValue:   DefaultSanitizePlaceholder,
				},
				{
					description:     "sanitize a well-formed name + email address",
					value:           "Jane Foo <foo@bar.com>",
					expectedHandled: true,
					expectedValue:   DefaultSanitizePlaceholder,
				},
				{
					description:     "sanitize a list of well-formed email addresses",
					value:           "foo@bar.com, bar@foo.com",
					expectedHandled: true,
					expectedValue:   DefaultSanitizePlaceholder,
				},
				{
					description:     "sanitize a list of well-formed name + email addresses",
					value:           "Jane Foo <foo@bar.com>, John Bar <bar@foo.com>",
					expectedHandled: true,
					expectedValue:   DefaultSanitizePlaceholder,
				},
			}

			for _, t := range testCases {
				func(t testCase) {
					It(fmt.Sprintf(`with value "%s", should %s`, t.value, t.description), func() {
						actualHandled, actualValue := sanitizeEmailAddress("", t.value, nil)
						Expect(actualHandled).To(Equal(t.expectedHandled))
						Expect(actualValue).To(Equal(t.expectedValue))
					})
				}(t)
			}
		})

		Context("sanitizeUrl()", func() {
			type testCase struct {
				value           string
				expectedValue   string
				expectedHandled bool
				description     string
			}

			testCases := []testCase{
				{description: "leave unhandled when not a well-formed url", value: "foo"},
				{
					description:     "leaves untouched if the url has no sensitive component",
					value:           "https://bar.com",
					expectedHandled: true,
					expectedValue:   "https://bar.com",
				},
				{
					description:     "strips user:password if present",
					value:           "https://user:pass@bar.com",
					expectedHandled: true,
					expectedValue:   "https://----@bar.com",
				},
				{
					description:     "sanitizes query string if present",
					value:           "https://bar.com?foo=bar&password=abcdef",
					expectedHandled: true,
					expectedValue:   "https://bar.com?foo=bar&password=----",
				},
				{
					description:     "sanitizes http://",
					value:           "https://bar.com?foo=bar&password=abcdef",
					expectedHandled: true,
					expectedValue:   "https://bar.com?foo=bar&password=----",
				},
				{
					description:     "sanitizes ftp:",
					value:           "ftp://user:pass@bar.com",
					expectedHandled: true,
					expectedValue:   "ftp://----@bar.com",
				},
				{
					description:     "sanitizes file:",
					value:           "file://user:pass@bar.com",
					expectedHandled: true,
					expectedValue:   "file://----@bar.com",
				},
				{
					description:     "sanitizes mailto:",
					value:           "mailto:fooo@bar.com",
					expectedHandled: true,
					expectedValue:   "mailto:----",
				},
				{
					description:     "sanitizes postgres:",
					value:           "postgres://user:pass@bar.com/db",
					expectedHandled: true,
					expectedValue:   "postgres://----@bar.com/db",
				},
			}

			for _, t := range testCases {
				func(t testCase) {
					It(fmt.Sprintf(`with value "%s", should %s`, t.value, t.description), func() {
						actualHandled, actualValue := sanitizeUrl("", t.value, nil)
						Expect(actualHandled).To(Equal(t.expectedHandled))
						Expect(actualValue).To(Equal(t.expectedValue))
					})
				}(t)
			}
		})

		Context("sanitizeByNamePattern()", func() {
			shouldMatchNames := []string{
				"password",
				"Password",
				"PaSsWoRd",
				"DBPassword",
			}

			shouldNotMatchNames := []string{
				"foo",
				"bar",
				"what up doc",
			}

			for _, shouldNotMatchName := range shouldNotMatchNames {
				func(shouldNotMatchName string) {
					It(fmt.Sprintf(`name "%s" should not be sanitized`, shouldNotMatchName), func() {
						actualHandled, actualValue := sanitizeByNamePattern(shouldNotMatchName, "foo", nil)
						Expect(actualHandled).To(Equal(false))
						Expect(actualValue).To(Equal(""))
					})
				}(shouldNotMatchName)
			}

			for _, shouldMatchName := range shouldMatchNames {
				func(shouldMatchName string) {
					It(fmt.Sprintf(`name "%s" should be sanitized`, shouldMatchName), func() {
						actualHandled, actualValue := sanitizeByNamePattern(shouldMatchName, "foo", nil)
						Expect(actualHandled).To(Equal(true))
						Expect(actualValue).To(Equal(DefaultSanitizePlaceholder))
					})
				}(shouldMatchName)
			}
		})

		Context("sanitizeString()", func() {

			Context("customSanitizer", func() {

				It("should be ignored if nil", func() {
					actualValue := sanitizeString("foo", "bar", nil)
					Expect(actualValue).To(Equal("bar"))

					actualValue = sanitizeString("password", "bar", nil)
					Expect(actualValue).To(Equal(DefaultSanitizePlaceholder))
				})

				It("should short-circuit default sanitizers if returns true", func() {
					customSanitizer := func(name string, value string, customSanitizer SanitizeFunc) (bool, string) {
						if name == "special" {
							return true, "oof"
						}
						return false, ""
					}

					actualValue := sanitizeString("special", "bar", customSanitizer)
					Expect(actualValue).To(Equal("oof"))
				})

				It("should fallback to default sanitizers if returns false", func() {
					customSanitizer := func(name string, value string, customSanitizer SanitizeFunc) (bool, string) {
						if name == "secial" {
							return true, "oof"
						}
						return false, ""
					}

					actualValue := sanitizeString("password", "bar", customSanitizer)
					Expect(actualValue).To(Equal(DefaultSanitizePlaceholder))
				})

				It("should be passed down the call chain and used recursively", func() {
					customSanitizer := func(name string, value string, customSanitizer SanitizeFunc) (bool, string) {
						if name == "special" {
							return true, "oof"
						}
						return false, ""
					}

					actualValue := sanitizeString("url", "https://foo.com/bar?special=especial", customSanitizer)
					Expect(actualValue).To(Equal("https://foo.com/bar?special=oof"))
				})
			})

			Context("default sanitizer cascading", func() {

				type testCase struct {
					name          string
					value         string
					expectedValue string
				}

				customSanitizer := func(name string, value string, customSanitizer SanitizeFunc) (bool, string) {
					if name == "secial" {
						return true, "oof"
					}
					return false, ""
				}

				testCases := []testCase{
					{name: "foo", value: "bar", expectedValue: "bar"},
					{name: "foo", value: "foo@bar.com", expectedValue: DefaultSanitizePlaceholder},
					{name: "password", value: "very secret", expectedValue: DefaultSanitizePlaceholder},
					{name: "my secret place", value: "https://foo.com/bar?password=bof", expectedValue: "https://foo.com/bar?password=----"},
				}

				for _, t := range testCases {
					func(t testCase) {
						It(fmt.Sprintf("should handle (%s, %s)", t.name, t.value), func() {
							actualValue := sanitizeString(t.name, t.value, customSanitizer)
							Expect(actualValue).To(Equal(t.expectedValue))
						})
					}(t)
				}

			})

			shouldMatchNames := []string{
				"password",
				"Password",
				"PaSsWoRd",
				"DBPassword",
			}

			shouldNotMatchNames := []string{
				"foo",
				"bar",
				"what up doc",
			}

			for _, shouldNotMatchName := range shouldNotMatchNames {
				func(shouldNotMatchName string) {
					It(fmt.Sprintf(`name "%s" should not be sanitized`, shouldNotMatchName), func() {
						actualHandled, actualValue := sanitizeByNamePattern(shouldNotMatchName, "foo", nil)
						Expect(actualHandled).To(Equal(false))
						Expect(actualValue).To(Equal(""))
					})
				}(shouldNotMatchName)
			}

			for _, shouldMatchName := range shouldMatchNames {
				func(shouldMatchName string) {
					It(fmt.Sprintf(`name "%s" should be sanitized`, shouldMatchName), func() {
						actualHandled, actualValue := sanitizeByNamePattern(shouldMatchName, "foo", nil)
						Expect(actualHandled).To(Equal(true))
						Expect(actualValue).To(Equal(DefaultSanitizePlaceholder))
					})
				}(shouldMatchName)
			}
		})
	})

})

// type I interface{}

// type StringAlias string

// type A struct {
// 	Greeting string
// 	Message  string
// 	Pi       float64
// 	Password string
// 	Sneaky   string
// 	Email    StringAlias
// 	Url      StringAlias
// }

// type B struct {
// 	Struct    A
// 	Ptr       *A
// 	Answer    int
// 	Map       map[string]string
// 	StructMap map[string]interface{}
// 	Slice     []string
// }

// func create() I {
// 	// The type C is actually hidden, but reflection allows us to look inside it
// 	type C struct {
// 		String string
// 	}

// 	return B{
// 		Struct: A{
// 			Greeting: "Hello!",
// 			Message:  "translate this",
// 			Pi:       3.14,
// 			Password: "secret pass",
// 			Email:    "foo@bar.com",
// 			Sneaky:   "bars@foo.com",
// 			// Url:      "http://foo.com?x=1&password=abcdef",
// 		},
// 		Ptr: &A{
// 			Greeting: "What's up?",
// 			Message:  "point here",
// 			Password: "so secret",
// 			Pi:       3.14,
// 			Url:      "http://user:foo@foo.com",
// 		},
// 		Map: map[string]string{
// 			"Test": "translate this as well",
// 		},
// 		StructMap: map[string]interface{}{
// 			"C": C{
// 				String: "deep",
// 			},
// 		},
// 		Slice: []string{
// 			"and one more",
// 		},
// 		Answer: 42,
// 	}
// }

// func TestSanitizer() {
// 	// Some example test cases so you can mess around and see if it's working
// 	// To check if it's correct look at the output, no automated checking here

// 	var sanitizer Sanitizer = nil

// Test the simple cases
// {
// 	fmt.Println("Test with nil pointer to struct:")
// 	var original *B
// 	translated := Sanitize(original, sanitizer)
// 	fmt.Println("original:  ", original)
// 	fmt.Println("translated:", translated)
// 	fmt.Println()
// }
// {
// 	fmt.Println("Test with nil pointer to interface:")
// 	var original *I
// 	translated := Sanitize(original, sanitizer)
// 	fmt.Println("original:  ", original)
// 	fmt.Println("translated:", translated)
// 	fmt.Println()
// }
// {
// 	fmt.Println("Test with struct that has no elements:")
// 	type E struct {
// 	}
// 	var original E
// 	translated := Sanitize(original, sanitizer)
// 	fmt.Println("original:  ", original)
// 	fmt.Println("translated:", translated)
// 	fmt.Println()
// }
// {
// 	fmt.Println("Test with empty struct:")
// 	var original B
// 	translated := Sanitize(original, sanitizer)
// 	fmt.Println("original:  ", original, "->", original.Ptr)
// 	fmt.Println("translated:", translated, "->", translated.(B).Ptr)
// 	fmt.Println()
// }

// 	// Imagine we have no influence on the value returned by create()
// 	created := create()
// 	{
// 		// Assume we know that `created` is of type B
// 		fmt.Println("Translating a struct:")
// 		original := created.(B)
// 		translated := Sanitize(original, sanitizer)
// 		fmt.Println("original:  ", original, "->", original.Ptr)
// 		fmt.Println("translated:", translated, "->", translated.(B).Ptr)
// 		fmt.Println()
// 	}
// 	{
// 		// Assume we don't know created's type
// 		fmt.Println("Translating a struct wrapped in an interface:")
// 		original := created
// 		translated := Sanitize(original, sanitizer)
// 		fmt.Println("original:  ", original, "->", original.(B).Ptr)
// 		fmt.Println("translated:", translated, "->", translated.(B).Ptr)
// 		fmt.Println()
// 	}
// 	{
// 		// Assume we don't know B's type and want to pass a pointer
// 		fmt.Println("Translating a pointer to a struct wrapped in an interface:")
// 		original := &created
// 		translated := Sanitize(original, sanitizer)
// 		fmt.Println("original:  ", (*original), "->", (*original).(B).Ptr)
// 		fmt.Println("translated:", (*translated.(*I)), "->", (*translated.(*I)).(B).Ptr)
// 		fmt.Println()
// 	}
// 	{
// 		// Assume we have a struct that contains an interface of an unknown type
// 		fmt.Println("Translating a struct containing a pointer to a struct wrapped in an interface:")
// 		type D struct {
// 			Payload *I
// 		}
// 		original := D{
// 			Payload: &created,
// 		}
// 		translated := Sanitize(original, sanitizer)
// 		fmt.Println("original:  ", original, "->", (*original.Payload), "->", (*original.Payload).(B).Ptr)
// 		fmt.Println("translated:", translated, "->", (*translated.(D).Payload), "->", (*(translated.(D).Payload)).(B).Ptr)
// 		fmt.Println()
// 	}
// }
