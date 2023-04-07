package storage

// import "fmt"
import {
	"github.com/khromalabs/keeper/storage/sqlite"
}

// Hello returns a greeting for the named person.
func Storage(name string) string {
    // Return a greeting that embeds the name in a message.
    // message := fmt.Sprintf("Hi, %v. Welcome!", name)
	message := SqliteHello()
    return message
}

