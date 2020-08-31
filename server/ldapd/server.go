package ldapd

import (
	"fmt"
	"io"
	"log"
	"math/rand"
	"net"
	"time"

	ber "github.com/go-asn1-ber/asn1-ber"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/sqlite" // Sqlite3 database
	"github.com/joho/godotenv"
	"github.com/muultipla/glim/server/db"
	"github.com/muultipla/glim/server/ldapd/ldap"
)

func handleConnection(c net.Conn, db *gorm.DB) {
	defer c.Close()
	fmt.Printf("Serving %s\n", c.RemoteAddr().String())
L:
	for {
		p, err := ber.ReadPacket(c)
		if err != nil {
			if err == io.EOF || err == io.ErrUnexpectedEOF {
				log.Println("connection closed by client")
				break
			}
			fmt.Println("Error", err)
			break
		}
		message, err := ldap.DecodeMessage(p)
		if err != nil {
			log.Println(err)
			break
		}

		switch message.Op {
		case ldap.BindRequest:
			log.Println("bind requested by client")
			p, err := ldap.HandleBind(message, db)
			if err != nil {
				log.Println(err)
			}
			_, err = c.Write(p.Bytes())
			if err != nil {
				log.Println(err)
			}
		case ldap.ExtendedRequest:
			p, err := ldap.HandleExtRequest(message)
			if err != nil {
				log.Println(err)
			}
			_, err = c.Write(p.Bytes())
			if err != nil {
				log.Println(err)
			}
		case ldap.SearchRequest:
			log.Println("search requested by client")
			p, err := ldap.HandleSearchRequest(message)
			if err != nil {
				log.Println(err)
			}
			for i := 0; i < len(p); i++ {
				_, err = c.Write(p[i].Bytes())
				if err != nil {
					log.Println(err)
				}
			}
			break L
		case ldap.UnbindRequest:
			log.Println("unbind requested by client")
		default:
			log.Printf("Operation %d not supported\n", message.Op)
			for i := 0; i < len(message.Request); i++ {
				fmt.Println(message.Request[i])
			}
			p, err := ldap.HandleUnsupportedOperation(message)
			if err != nil {
				log.Println(err)
			}
			_, err = c.Write(p.Bytes())
			if err != nil {
				log.Println(err)
			}
			break L
		}
	}
}

// Server - TODO comment
func Server() {
	// Get environment variables
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Error getting env, not comming through %v", err)
	}

	PORT := ":1389"
	l, err := net.Listen("tcp4", PORT)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer l.Close()
	rand.Seed(time.Now().Unix())

	// Database
	database, err := db.Initialize()
	if err != nil {
		log.Fatal(err)
	}
	defer database.Close()

	log.Println("Starting ldap server in port 1389")

	for {
		c, err := l.Accept()
		if err != nil {
			fmt.Println(err)
			return
		}
		go handleConnection(c, database)
	}
}
