package main

import (
	"bytes"
	"fmt"
)

func main() {
	// Helper functions to create pointer values
	strPtr := func(s string) *string { return &s }
	intPtr := func(i int) *int { return &i }
	float64Ptr := func(f float64) *float64 { return &f }

	fmt.Println("=== All Param Types Example ===")
	fmt.Println()

	// 1. Render basic_types template
	fmt.Println("--- 1. Basic Types ---")
	var buf1 bytes.Buffer
	_ = RenderBasic_types(&buf1, Basic_types{
		Name:   "John Doe",
		Age:    30,
		Score:  98765,
		Price:  99.99,
		Active: true,
	})
	fmt.Print(buf1.String())
	fmt.Println()

	// 2. Render pointer_types template
	fmt.Println("--- 2. Pointer Types (Optional Fields) ---")
	var buf2 bytes.Buffer
	_ = RenderPointer_types(&buf2, Pointer_types{
		Email:       strPtr("john@example.com"),
		PhoneNumber: nil, // Not provided
		MiddleScore: intPtr(85),
		Discount:    float64Ptr(15.5),
	})
	fmt.Print(buf2.String())
	fmt.Println()

	// 3. Render slice_types template
	fmt.Println("--- 3. Slice Types ---")
	var buf3 bytes.Buffer
	_ = RenderSlice_types(&buf3, Slice_types{
		Tags:        []string{"golang", "template", "example"},
		CategoryIDs: []int{1, 2, 3, 5, 8},
		Ratings:     []float64{4.5, 3.8, 5.0, 4.2},
		Flags:       []bool{true, false, true, true},
	})
	fmt.Print(buf3.String())
	fmt.Println()

	// 4. Render map_types template
	fmt.Println("--- 4. Map Types ---")
	var buf4 bytes.Buffer
	_ = RenderMap_types(&buf4, Map_types{
		Metadata: map[string]string{
			"author":  "templagen",
			"version": "1.0",
			"license": "MIT",
		},
		Counters: map[string]int{
			"views":     1000,
			"downloads": 250,
			"stars":     42,
		},
		Prices: map[string]float64{
			"basic":      9.99,
			"premium":    29.99,
			"enterprise": 99.99,
		},
		Features: map[string]bool{
			"authentication": true,
			"logging":        true,
			"analytics":      false,
		},
	})
	fmt.Print(buf4.String())
	fmt.Println()

	// 5. Render struct_types template
	fmt.Println("--- 5. Struct Types (Nested Fields) ---")
	var buf5 bytes.Buffer
	_ = RenderStruct_types(&buf5, Struct_types{
		User: Struct_typesUser{
			ID:    12345,
			Name:  "Alice Smith",
			Email: "alice@example.com",
		},
		Product: Struct_typesProduct{
			SKU:     "PROD-001",
			Price:   149.99,
			InStock: true,
		},
	})
	fmt.Print(buf5.String())
	fmt.Println()

	// 6. Render complex_types template
	fmt.Println("--- 6. Complex/Nested Types ---")
	var buf6 bytes.Buffer
	_ = RenderComplex_types(&buf6, Complex_types{
		Items: []Complex_typesItemsItem{
			{
				ID:    1,
				Title: "Learning Go",
				Tags:  []string{"book", "programming", "go"},
				Price: 39.99,
			},
			{
				ID:    2,
				Title: "Template Patterns",
				Tags:  []string{"book", "design", "templates"},
				Price: 29.99,
			},
			{
				ID:    3,
				Title: "Advanced Testing",
				Tags:  []string{"testing", "quality", "go"},
				Price: 49.99,
			},
		},
		OptionalItems: &[]string{"item1", "item2", "item3"},
		Records: []Complex_typesRecordsItem{
			{
				Name:  "Record A",
				Age:   25,
				Score: intPtr(95),
			},
			{
				Name:  "Record B",
				Age:   30,
				Score: nil, // Not set
			},
			{
				Name:  "Record C",
				Age:   28,
				Score: intPtr(88),
			},
		},
	})
	fmt.Print(buf6.String())
	fmt.Println()
}
