package main

import (
	"bytes"
	"fmt"
)

func main() {
	// Example: Using type-safe RenderReport with comprehensive template features
	fmt.Println("=== Comprehensive Template Example ===")

	var buf bytes.Buffer
	_ = RenderReport(&buf, Report{
		// 1. Basic field reference
		Title: "Q4 2024 Report",

		// 2. Nested field reference
		Author: ReportAuthor{
			Name:  "Alice Johnson",
			Email: "alice@example.com",
		},

		// 3. Conditional rendering
		Status: "Published",

		// 4. With statement and else clause
		Summary: ReportSummary{
			Content:     "This report summarizes the key achievements and metrics for Q4 2024.",
			LastUpdated: "2024-12-31",
		},
		DefaultMessage: "No summary provided.",

		// 5. Range over slice
		Items: []ReportItemsItem{
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

		// 6. Index function for map access
		Meta: map[string]string{
			"env":     "production",
			"version": "2.4.1",
			"build":   "2024-12-30T10:00:00Z",
		},

		// 7. Nested structures: with + range
		Project: ReportProject{
			Name:        "Q4 Initiatives",
			Description: "Key projects and initiatives completed in Q4 2024",
			Tasks: []ReportTasksItem{
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

		// 8. Deep nested path
		Company: ReportCompany{
			Department: ReportDepartment{
				Team: ReportTeam{
					Manager: ReportManager{
						Name: "Bob Smith",
					},
				},
			},
		},
	})

	fmt.Println(buf.String())
}
