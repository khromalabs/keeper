package storage

import {
	"fmt"
}

// Hello returns a greeting for the named person.
func Storage(name string) string {
    // Return a greeting that embeds the name in a message.
    // message := fmt.Sprintf("Hi, %v. Welcome!", name)
	message := sqlite.Hello()
    return message
}

