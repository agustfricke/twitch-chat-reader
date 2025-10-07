package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"strings"
	"time"
)

const (
	server = "irc.chat.twitch.tv:6667"
)

func main() {
	// For anonymous read-only access
	username := fmt.Sprintf("random_username%d", time.Now().Unix()%100000)
	channel := "#ohnepixel" // Channel to read - must include #

	conn, err := net.Dial("tcp", server)
	if err != nil {
		log.Fatal("Failed to connect:", err)
	}
	defer conn.Close()

	fmt.Println("Connected to Twitch IRC")

	// Send connection commands
	send(conn, "CAP REQ :twitch.tv/commands twitch.tv/tags")
	send(conn, "NICK "+username)
	send(conn, "JOIN "+channel)

	fmt.Printf("Joined %s, reading messages...\n\n", channel)

	scanner := bufio.NewScanner(conn)
	for scanner.Scan() {
		line := scanner.Text()
		
		// Debug: print raw line
		// fmt.Println("RAW:", line)

		// Respond to PING immediately
		if strings.HasPrefix(line, "PING") {
			pong := strings.Replace(line, "PING", "PONG", 1)
			send(conn, pong)
			fmt.Println("↔ Ping/Pong")
			continue
		}

		// Check for successful join
		if strings.Contains(line, "JOIN") && strings.Contains(line, username) {
			fmt.Println("✓ Successfully joined channel")
			continue
		}

		// Parse chat messages
		if strings.Contains(line, "PRIVMSG") {
			msg := parseMessage(line)
			if msg != "" {
				fmt.Println(msg)
			}
		}

		// Handle connection errors
		if strings.Contains(line, "NOTICE") && strings.Contains(line, "Login authentication failed") {
			log.Fatal("Authentication failed")
		}
	}

	if err := scanner.Err(); err != nil {
		log.Fatal("Connection error:", err)
	}
}

func send(conn net.Conn, message string) {
	fmt.Fprintf(conn, "%s\r\n", message)
}

func parseMessage(line string) string {
	// Format: @tags :user!user@user.tmi.twitch.tv PRIVMSG #channel :message
	
	// Remove tags if present
	if strings.HasPrefix(line, "@") {
		parts := strings.SplitN(line, " :", 2)
		if len(parts) > 1 {
			line = ":" + parts[1]
		}
	}

	// Split into parts
	parts := strings.Split(line, "PRIVMSG")
	if len(parts) < 2 {
		return ""
	}

	// Extract username
	userPart := strings.TrimPrefix(parts[0], ":")
	userPart = strings.TrimSpace(userPart)
	username := strings.Split(userPart, "!")[0]

	// Extract message
	msgParts := strings.SplitN(parts[1], ":", 2)
	if len(msgParts) < 2 {
		return ""
	}
	message := strings.TrimSpace(msgParts[1])

	return fmt.Sprintf("[%s]: %s", username, message)
}
