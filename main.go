package main

import (
	"fmt"
	"log"
	"os"

	"github.com/emersion/go-imap/client"
)

func getMessageCount(user string, password string, mailbox string) uint32 {
	log.Println("Connecting...")

	c, err := client.DialTLS("imap.migadu.com:993", nil)
	if err != nil {
		panic(err)
	}

	defer c.Logout()

	if err := c.Login(user, password); err != nil {
		panic(err)
	}
	inbox, err := c.Select(mailbox, true)
	if err != nil {
		panic(err)
	}

	return inbox.Messages
}

func main() {
	messageCount := getMessageCount(os.Getenv("IMAP_USER"), os.Getenv("IMAP_PASS"), "INBOX")
	fmt.Println(messageCount)
}
