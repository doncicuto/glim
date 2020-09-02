package ldap

import (
	"fmt"
	"io"
	"math/rand"
	"net"
	"sync"
	"time"

	ber "github.com/go-asn1-ber/asn1-ber"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/sqlite" // Sqlite3 database
	"github.com/joho/godotenv"
	"github.com/labstack/gommon/log"
	"github.com/muultipla/glim/server/db"
)

func printLog(msg string) {
	log.SetHeader("${time_rfc3339} [LDAP] ⇨")
	log.Print(msg)
}

func handleConnection(c net.Conn, db *gorm.DB) {
	defer c.Close()

	remoteAddress := c.RemoteAddr().String()
	printLog(fmt.Sprintf("serving LDAP connection from %s", remoteAddress))
L:
	for {
		p, err := ber.ReadPacket(c)
		if err != nil {
			if err == io.EOF || err == io.ErrUnexpectedEOF {
				printLog(fmt.Sprintf("connection closed by client %s", remoteAddress))
				break
			}
			fmt.Println("Error", err)
			break
		}
		message, err := DecodeMessage(p)
		if err != nil {
			printLog(err.Error())
			break
		}

		switch message.Op {
		case BindRequest:
			printLog(fmt.Sprintf("bind requested by client %s", remoteAddress))
			p, err := HandleBind(message, db, remoteAddress)
			if err != nil {
				printLog(err.Error())
			}
			_, err = c.Write(p.Bytes())
			if err != nil {
				printLog(err.Error())
			}
		case ExtendedRequest:
			p, err := HandleExtRequest(message)
			if err != nil {
				printLog(err.Error())
			}
			_, err = c.Write(p.Bytes())
			if err != nil {
				printLog(err.Error())
			}
		case SearchRequest:
			printLog(fmt.Sprintf("search requested by client %s", remoteAddress))
			p, err := HandleSearchRequest(message)
			if err != nil {
				printLog(err.Error())
			}
			for i := 0; i < len(p); i++ {
				_, err = c.Write(p[i].Bytes())
				if err != nil {
					printLog(err.Error())
				}
			}
			break L
		case UnbindRequest:
			printLog(fmt.Sprintf("unbind requested by client %s", remoteAddress))
		default:
			printLog(fmt.Sprintf("operation %d not supported requested by client %s", message.Op, remoteAddress))
			for i := 0; i < len(message.Request); i++ {
				fmt.Println(message.Request[i])
			}
			p, err := HandleUnsupportedOperation(message)
			if err != nil {
				printLog(err.Error())
			}
			_, err = c.Write(p.Bytes())
			if err != nil {
				printLog(err.Error())
			}
			break L
		}
	}
}

// Server - TODO comment
func Server(wg *sync.WaitGroup) {
	defer wg.Done()

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

	log.SetHeader("${time_rfc3339} [Glim] ⇨")
	log.Print("starting LDAP server in port 1389...")

	for {
		c, err := l.Accept()
		if err != nil {
			fmt.Println(err)
			return
		}
		go handleConnection(c, database)
	}
}
