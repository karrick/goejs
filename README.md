# goejs

Miniature Go library for properly encoding strings to JSON and back again

## Description

Want to quickly serialize Go UTF-8 strings into JSON using all of the
relevant RFC documents? Then this library is for you.

While the [defacto JSON website](http://json.org) only mentions a few
rules for serializing strings into JSON, there have been many RFC
documents that seek to clarify some aspects.

## References

## Usage

When encoding to JSON this library appends to a pre-existing slice of
bytes, using the runtime to append to this slice, so it can minimize
allocations if the byte slice capacity is large enough to accomodate
the encoded form of the string.

```Go
func ExampleEncode() {
	encoded := goejs.EncodedJSONFromString([]byte("prefix:"), "⌘ a")
	fmt.Printf("%s", encoded)
	// Output: prefix:"\u0001\u2318 a"
}
```

When decoding from JSON this library consumes bytes from an existing
byte slice and not only returns the decoded string but also a byte
slice of the remaining bytes, pointed at the original byte slice's
backing array.

```Go
func ExampleDecode() {
	decoded, remainder, err := goejs.DecodedStringFromJSON([]byte("\"\\u0001\\u2318 a\" some extra bytes after final quote"))
	if err != nil {
		fmt.Println(err)
	}
	if actual, expected := string(remainder), " some extra bytes after final quote"; actual != expected {
		fmt.Printf("Remainder Actual: %#q; Expected: %#q\n", actual, expected)
	}
	fmt.Printf("%v", decoded)
	// Output: ⌘ a
}
```
