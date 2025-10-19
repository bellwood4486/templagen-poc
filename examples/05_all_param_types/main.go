package main

import (
	"bytes"
	"fmt"
	"os"
	"time"
)

func main() {
	// Helper function to create pointer values
	strPtr := func(s string) *string { return &s }
	intPtr := func(i int) *int { return &i }
	float64Ptr := func(f float64) *float64 { return &f }

	// Create sample data with all supported param types
	data := All_types{
		// 1. Basic Types
		Name:      "John Doe",
		Age:       30,
		Score:     98765,
		Price:     99.99,
		Active:    true,
		CreatedAt: time.Date(2024, 10, 19, 15, 30, 0, 0, time.UTC),

		// 2. Pointer Types (Optional values)
		Email:       strPtr("john@example.com"),
		PhoneNumber: nil, // Not provided
		MiddleScore: intPtr(85),
		Discount:    float64Ptr(15.5),

		// 3. Slice Types
		Tags:        []string{"golang", "template", "example"},
		CategoryIDs: []int{1, 2, 3, 5, 8},
		Ratings:     []float64{4.5, 3.8, 5.0, 4.2},
		Flags:       []bool{true, false, true, true},

		// 4. Map Types
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

		// 5. Struct Types
		User: All_typesUser{
			ID:    12345,
			Name:  "Alice Smith",
			Email: "alice@example.com",
		},
		Product: All_typesProduct{
			SKU:     "PROD-001",
			Price:   149.99,
			InStock: true,
		},

		// 6. Complex/Nested Types

		// Slice of structs with nested slice
		Items: []All_typesItemsItem{
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

		// Optional slice
		OptionalItems: &[]string{"item1", "item2", "item3"},

		// Slice of structs with optional value
		Records: []All_typesRecordsItem{
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
	}

	// Render the template
	var buf bytes.Buffer
	if err := RenderAll_types(&buf, data); err != nil {
		fmt.Fprintf(os.Stderr, "Error rendering template: %v\n", err)
		os.Exit(1)
	}

	fmt.Print(buf.String())
}
