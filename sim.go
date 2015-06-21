package formation

import (
	"reflect"
	"strings"
)

var PseudoParams = map[string]string{
	"AWS::AccountId":        "123456789012",
	"AWS::NotificationARNs": "arn1, arn2", // []string{"arn1, arn2"},
	"AWS::NoValue":          "",
	"AWS::Region":           "us-west-2",
	"AWS::StackId":          "arn:aws:cloudformation:us-west-2:123456789012:stack/teststack/51af3dc0-da77-11e4-872e-1234567db123",
	"AWS::StackName":        "teststack",
}

func IsRef(obj reflect.Value) bool {
	if obj.Type() != reflect.TypeOf(make(map[string]interface{})) {
		return false
	}

	if len(obj.MapKeys()) != 1 {
		return false
	}

	return obj.MapKeys()[0].String() == "Ref"
}

func RefValue(obj reflect.Value) string {
	k := obj.MapKeys()[0]
	v := obj.MapIndex(k)

	return PseudoParams[v.Elem().String()]
}

func IsFnJoin(obj reflect.Value) bool {
	if obj.Type() != reflect.TypeOf(make(map[string]interface{})) {
		return false
	}

	if len(obj.MapKeys()) != 1 {
		return false
	}

	return obj.MapKeys()[0].String() == "Fn::Join"
}

func translate(obj interface{}) interface{} {
	// Wrap the original in a reflect.Value
	original := reflect.ValueOf(obj)

	// fmt.Printf("TRANSLATE %+v (%+v)\n", obj, original.Type())

	copy := reflect.New(original.Type()).Elem()
	translateRecursive(copy, original)

	// Remove the reflection wrapper
	return copy.Interface()
}

func translateRecursive(copy, original reflect.Value) {
	// fmt.Printf("%+v ; %+v\n", original, original.Type())

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
		// Unwrap the newly created pointer
		translateRecursive(copy.Elem(), originalValue)

	// If it is an interface (which is very similar to a pointer), do basically the
	// same as for the pointer. Though a pointer is not the same as an interface so
	// note that we have to call Elem() after creating a new object because otherwise
	// we would end up with an actual pointer
	case reflect.Interface:
		// Get rid of the wrapping interface
		originalValue := original.Elem()

		if IsRef(originalValue) {
			copyValue := reflect.New(reflect.TypeOf("")).Elem()
			copyValue.SetString(RefValue(originalValue))
			translateRecursive(copyValue, copyValue)
			copy.Set(copyValue)
		} else if IsFnJoin(originalValue) {
			k := originalValue.MapKeys()[0]
			v := originalValue.MapIndex(k)

			delim := v.Elem().Index(0).Elem().String()
			parts := v.Elem().Index(1).Elem()

			p := make([]string, parts.Len())
			for i := 0; i < parts.Len(); i++ {
				e := parts.Index(i).Elem()

				if IsRef(e) {
					p[i] = RefValue(e)
				} else {
					p[i] = parts.Index(i).Elem().String()
				}
			}

			copyValue := reflect.New(reflect.TypeOf("")).Elem()
			copyValue.SetString(strings.Join(p, delim))
			translateRecursive(copyValue, copyValue)
			copy.Set(copyValue)
		} else {
			// Create a new object. Now new gives us a pointer, but we want the value it
			// points to, so we have to call Elem() to unwrap it
			copyValue := reflect.New(originalValue.Type()).Elem()
			translateRecursive(copyValue, originalValue)
			copy.Set(copyValue)
		}

	// If it is a struct we translate each field
	case reflect.Struct:
		for i := 0; i < original.NumField(); i += 1 {
			translateRecursive(copy.Field(i), original.Field(i))
		}

	// If it is a slice we create a new slice and translate each element
	case reflect.Slice:
		copy.Set(reflect.MakeSlice(original.Type(), original.Len(), original.Cap()))
		for i := 0; i < original.Len(); i += 1 {
			translateRecursive(copy.Index(i), original.Index(i))
		}

	// If it is a map we create a new map and translate each value
	case reflect.Map:
		copy.Set(reflect.MakeMap(original.Type()))
		for _, key := range original.MapKeys() {
			originalValue := original.MapIndex(key)
			// New gives us a pointer, but again we want the value
			copyValue := reflect.New(originalValue.Type()).Elem()
			translateRecursive(copyValue, originalValue)
			copy.SetMapIndex(key, copyValue)
		}

	// Otherwise we cannot traverse anywhere so this finishes the the recursion

	// And everything else will simply be taken from the original
	default:
		copy.Set(original)
	}
}
