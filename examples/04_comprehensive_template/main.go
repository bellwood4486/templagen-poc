package main

import (
	"bytes"
	"fmt"
)

func main() {
	// Example: Using type-safe render functions with comprehensive template features
	fmt.Println("=== Comprehensive Template Example ===")
	fmt.Println()

	// 1. Render basic_fields template
	fmt.Println("--- Template 1: basic_fields ---")
	fmt.Println("Features: 1. Basic Field Reference & 2. Nested Field Reference")
	fmt.Println()
	var buf1 bytes.Buffer
	_ = RenderBasic_fields(&buf1, Basic_fields{
		Title: "Q4 2024 Report",
		Author: Basic_fieldsAuthor{
			Name:  "Alice Johnson",
			Email: "alice@example.com",
		},
	})
	fmt.Println(buf1.String())
	fmt.Println()

	// 2. Render control_flow template
	fmt.Println("--- Template 2: control_flow ---")
	fmt.Println("Features: 3. Conditional (if) & 4. With Statement and Else Clause")
	fmt.Println()
	var buf2 bytes.Buffer
	_ = RenderControl_flow(&buf2, Control_flow{
		Status: "Published",
		Summary: Control_flowSummary{
			Content:     "This report summarizes the key achievements and metrics for Q4 2024.",
			LastUpdated: "2024-12-31",
		},
		DefaultMessage: "No summary provided.",
	})
	fmt.Println(buf2.String())
	fmt.Println()

	// 3. Render collections template
	fmt.Println("--- Template 3: collections ---")
	fmt.Println("Features: 5. Range Over Slice & 6. Map Access with Index Function")
	fmt.Println()
	var buf3 bytes.Buffer
	_ = RenderCollections(&buf3, Collections{
		Items: []CollectionsItemsItem{
			{
				ID:          "1",
				Title:       "Revenue Growth",
				Description: "Revenue increased by 25% compared to Q3 2024.",
			},
			{
				ID:          "2",
				Title:       "User Engagement",
				Description: "Active users grew by 40% with improved retention rates.",
			},
			{
				ID:          "3",
				Title:       "Product Launch",
				Description: "Successfully launched three new features.",
			},
		},
		Meta: map[string]string{
			"env":     "production",
			"version": "2.4.1",
			"build":   "2024-12-30T10:00:00Z",
		},
	})
	fmt.Println(buf3.String())
	fmt.Println()

	// 4. Render advanced template
	fmt.Println("--- Template 4: advanced ---")
	fmt.Println("Features: 7. Nested Structures (With + Range) & 8. Deep Nested Path")
	fmt.Println()
	var buf4 bytes.Buffer
	_ = RenderAdvanced(&buf4, Advanced{
		Project: AdvancedProject{
			Name:        "Q4 Initiatives",
			Description: "Key projects and initiatives completed in Q4 2024",
			Tasks: []AdvancedTasksItem{
				{
					Title:  "Infrastructure Upgrade",
					Status: "Completed",
				},
				{
					Title:  "API Redesign",
					Status: "In Progress",
				},
				{
					Title:  "Security Audit",
					Status: "Completed",
				},
			},
		},
		Company: AdvancedCompany{
			Department: AdvancedDepartment{
				Team: AdvancedTeam{
					Manager: AdvancedManager{
						Name: "Bob Smith",
					},
				},
			},
		},
	})
	fmt.Println(buf4.String())
	fmt.Println()
}
