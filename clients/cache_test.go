package clients

import (
	"testing"

	"github.com/google/go-cmp/cmp"
)

type NestedStruct struct {
	TestString string
}

type TestStruct struct {
	TestSlice []int
	Nested    NestedStruct
	TestMap   map[string]NestedStruct
}

func TestByteConversions(t *testing.T) {
	expectedStruct := TestStruct{
		TestSlice: []int{2, 3, 5, 7, 11, 13},
		Nested:    NestedStruct{"nestedTestValue"},
		TestMap: map[string]NestedStruct{
			"testKey1": NestedStruct{"testValue1"},
			"testKey2": NestedStruct{"testValue2"},
		},
	}

	for i := 0; i < 100; i++ {
		testBytes, _ := ToBytes(expectedStruct)

		var actualStruct TestStruct
		FromBytes(testBytes, &actualStruct)

		if !cmp.Equal(expectedStruct, actualStruct) {
			t.Errorf("byte/struct conversion failed, got: %v, want: %v.", actualStruct, expectedStruct)
		}
	}
}
