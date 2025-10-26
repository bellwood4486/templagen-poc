package main

import (
	"bytes"
	"fmt"
)

func main() {
	fmt.Println("=== Example: Template Grouping ===")
	fmt.Println()

	// Use grouped templates - MailInvite
	fmt.Println("--- Mail Invite ---")
	var inviteTitleBuf, inviteContentBuf bytes.Buffer

	_ = RenderMailInviteTitle(&inviteTitleBuf, MailInviteTitle{
		SiteName:    "MyApp",
		InviterName: "Alice",
	})
	fmt.Println("Title:", inviteTitleBuf.String())

	_ = RenderMailInviteContent(&inviteContentBuf, MailInviteContent{
		RecipientName: "Bob",
		InviterName:   "Alice",
		SiteName:      "MyApp",
		InviteURL:     "https://myapp.com/invite/abc123",
	})
	fmt.Println("Content:")
	fmt.Println(inviteContentBuf.String())
	fmt.Println()

	// Use grouped templates - MailAccountCreated
	fmt.Println("--- Mail Account Created ---")
	var accountTitleBuf, accountContentBuf bytes.Buffer

	_ = RenderMailAccountCreatedTitle(&accountTitleBuf, MailAccountCreatedTitle{
		SiteName: "MyApp",
	})
	fmt.Println("Title:", accountTitleBuf.String())

	_ = RenderMailAccountCreatedContent(&accountContentBuf, MailAccountCreatedContent{
		Username: "bob123",
		Email:    "bob@example.com",
		SiteName: "MyApp",
	})
	fmt.Println("Content:")
	fmt.Println(accountContentBuf.String())
	fmt.Println()

	// Use generic Render with grouped template
	fmt.Println("--- Mail Article Created (via generic Render) ---")
	var articleTitleBuf bytes.Buffer
	_ = Render(&articleTitleBuf, Template.MailArticleCreated.Title, MailArticleCreatedTitle{
		ArticleTitle: "10 Tips for Better Go Code",
	})
	fmt.Println("Title:", articleTitleBuf.String())
	fmt.Println()

	// Show all available templates
	fmt.Println("--- Available Templates ---")
	templates := Templates()
	fmt.Printf("Total templates: %d\n", len(templates))
	for name := range templates {
		fmt.Printf("  - %s\n", name)
	}
}
