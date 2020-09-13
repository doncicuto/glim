package ldap

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/asaskevich/govalidator"
)

// Domain - TODO comment
func Domain() string {
	const defaultDomain string = "dc=example,dc=org"
	domain := os.Getenv("LDAP_DOMAIN")
	if domain == "" {
		return defaultDomain
	}

	if !govalidator.IsDNSName(domain) {
		fmt.Printf("%s [Glim] ⇨ LDAP_DOMAIN env does not contain a valid domain, using example.org...\n", time.Now().Format(time.RFC3339))
		return defaultDomain
	}

	ldapDomain := ""
	domainParts := strings.Split(domain, ".")
	for i, part := range domainParts {
		ldapDomain += fmt.Sprintf("dc=%s", part)
		if len(domainParts) != i+1 {
			ldapDomain += ","
		}
	}
	return ldapDomain
}
