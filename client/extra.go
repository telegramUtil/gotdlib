package client

import (
	"strings"

	"github.com/google/uuid"
)

type ExtraGenerator func() string

func UuidV4Generator() ExtraGenerator {
	return func() string {
		return uuid.NewString()
	}
}

func IsCommand(text string) bool {
	if text != "" {
		if text[0] == '/' {
			return true
		}
	}
	return false
}

func CheckCommand(text string, entities []*TextEntity) string {
	if IsCommand(text) {
		// Check text entities and make bot happy!
		if len(entities) >= 1 {
			// Get first command
			if entities[0].Type.TextEntityTypeType() == "textEntityTypeBotCommand" {
				// e.g.: { "text": "/hello@world_bot", "textEntity": { offset: 0, length: 16 } }
				// Result: "/hello"
				if i := strings.Index(text[:entities[0].Length], "@"); i != -1 {
					return text[:i]
				}
				return text[:entities[0].Length]
			}
		} else {
			// Since userbot does not have bot command entities in Private Chat, so make userbot happy too!
			// e.g.: ["/hello@world_bot", "/hello@", "/hello@123"]
			// Result: "/hello"
			if i := strings.Index(text, "@"); i != -1 {
				return text[:i]
			}
			// e.g. ["/hello 123", "/hell o 123"]
			// Result: "/hello", "/hell"
			if i := strings.Index(text, " "); i != -1 {
				return text[:i]
			}
			return text
		}
	}
	return ""
}

func CommandArgument(text string) string {
	if IsCommand(text) {
		// e.g. ["/hello 123", "/hell o 123"]
		// Result: "123", "o 123"
		if i := strings.Index(text, " "); i != -1 {
			return text[i+1:]
		}
	}
	return ""
}
