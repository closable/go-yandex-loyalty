package config

import (
	"fmt"
)

func ExampleFirstValue() {
	a, b := "ABC", ""

	c, d := "", "CDE"

	e, f := "ABC", "CDE"

	out1 := FirstValue(&a, &b)
	fmt.Println(out1)

	out2 := FirstValue(&c, &d)
	fmt.Println(out2)

	out3 := FirstValue(&e, &f)
	fmt.Println(out3)

	// Output:
	// ABC
	// CDE
	// ABC
}
