package ldap

import (
	"crypto/tls"
	"fmt"
	"io"
	"net"
	"os"
	"sync"

	ber "github.com/go-asn1-ber/asn1-ber"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/sqlite" // Sqlite3 database
	"github.com/labstack/gommon/log"
)

//Settings - TODO comment
type Settings struct {
	DB      *gorm.DB
	TLSCert string
	TLSKey  string
}

const ldapAddr = ":1636"

func printLog(msg string) {
	log.SetHeader("${time_rfc3339} [LDAP] ⇨")
	log.Print(msg)
}

func handleConnection(c net.Conn, db *gorm.DB) {
	defer c.Close()

	remoteAddress := c.RemoteAddr().String()
	printLog(fmt.Sprintf("serving LDAPS connection from %s", remoteAddress))
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
			printLog(fmt.Sprintf("bind requested by client: %s", remoteAddress))
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
			printLog(fmt.Sprintf("unbind requested by client: %s", remoteAddress))
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
func Server(wg *sync.WaitGroup, shutdownChannel chan bool, settings Settings) {
	defer wg.Done()

	addr, ok := os.LookupEnv("LDAP_SERVER_ADDRESS")
	if !ok {
		addr = ldapAddr
	}

	// Load server certificate and private key
	cer, err := tls.LoadX509KeyPair(settings.TLSCert, settings.TLSKey)
	if err != nil {
		log.SetHeader("${time_rfc3339} [Glim] ⇨")
		log.Fatal("could not load server certificate and private key pair")
		return
	}

	// Start TLS listener
	config := &tls.Config{Certificates: []tls.Certificate{cer}}
	l, err := tls.Listen("tcp", addr, config)
	if err != nil {
		log.SetHeader("${time_rfc3339} [Glim] ⇨")
		log.Fatal("")
		return
	}
	defer l.Close()

	log.SetHeader("${time_rfc3339} [Glim] ⇨")
	log.Printf("starting LDAPS server in address %s...", addr)

	// Handle LDAP connections in a for loop
	for {
		// Wait for shutdown signals and gracefully shutdown TLS listener
		// Reference: https://gist.github.com/rcrowley/5474430
		// Reference: https://echo.labstack.com/cookbook/graceful-shutdown
		if <-shutdownChannel {
			log.SetHeader("${time_rfc3339} [Glim] ⇨")
			log.Printf("shutting down LDAPS server...")
			l.Close()
			break
		}

		c, err := l.Accept()
		if err != nil {
			log.SetHeader("${time_rfc3339} [Glim] ⇨")
			log.Printf("an error occurred accepting connections...")
			break
		}
		go handleConnection(c, settings.DB)
	}
}
