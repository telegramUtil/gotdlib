package client

import (
	"fmt"
	"math/rand"
	"strings"
)

type ExtraGenerator func() string

func UuidV4Generator() ExtraGenerator {
	return func() string {
		var uuid [16]byte
		rand.Read(uuid[:])

		uuid[6] = (uuid[6] & 0x0f) | 0x40
		uuid[8] = (uuid[8] & 0x3f) | 0x80

		return fmt.Sprintf("%08x-%04x-%04x-%04x-%012x", uuid[:4], uuid[4:6], uuid[6:8], uuid[8:10], uuid[10:])
	}
}

func IsCommmand(text string) bool {
	if text != "" {
		if text[0] == '/' {
			return true
		}
	}
	return false
}

func CheckCommand(text string, entities []*TextEntity) string {
	if IsCommmand(text) {
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
	if IsCommmand(text) {
		// e.g. ["/hello 123", "/hell o 123"]
		// Result: "123", "o 123"
		if i := strings.Index(text, " "); i != -1 {
			return text[i+1:]
		}
	}
	return ""
}
