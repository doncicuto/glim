# Testing Glim with OpenLDAP

In order to test if Glim can speak some LDAP, I've used the following tests that may help you when evaluating Glim and checking if it is working fine.

> Note: You must specify the location for your CA certificate so OpenLDAP can verify Glim server's certificate and the TLS handshake can run smoothly. You can use the **LDAPTLS_CACERT** environment variable.

## Who am I?

```(bash)
$ LDAPTLS_CACERT=/tmp/ca.pem ldapwhoami -x -D "cn=admin,dc=example,dc=org" -W -H ldaps://127.0.0.1:1636
Enter LDAP Password: (type the admin password and press Enter)
dn:cn=admin,dc=example,dc=org (cool this is who I am)
```

## Get all users inside ou=Users

```(bash)
$ LDAPTLS_CACERT=/tmp/ca.pem ldapsearch -x -D "cn=admin,dc=example,dc=org" -W -b "ou=Users,dc=example,dc=org" -H ldaps://127.0.0.1:1636
Enter LDAP Password: (type the admin password and press Enter)

# extended LDIF
#
# LDAPv3
# base <ou=Users,dc=example,dc=org> with scope subtree
# filter: (objectclass=*)
# requesting: ALL
#

# Users, example.org
dn: ou=Users,dc=example,dc=org
objectClass: organizationalUnit
objectClass: top
ou: Users

# mcabrerizo, Users, example.org
dn: uid=mcabrerizo,ou=Users,dc=example,dc=org
objectClass: top
objectClass: person
objectClass: inetOrgPerson
objectClass: organizationalPerson
uid: mcabrerizo
cn: Miguel Cabrerizo
sn: Cabrerizo
givenName: Miguel
mail: mcabrerizo@xxxxxx
memberOf: cn=devel,ou=Groups,dc=example,dc=org

# search result
search: 2
result: 0 Success

# numResponses: 3
# numEntries: 2

```

## Get all groups inside ou=Groups

```(bash)
$ LDAPTLS_CACERT=/tmp/ca.pem ldapsearch -x -D "cn=admin,dc=example,dc=org" -W -b "ou=Groups,dc=example,dc=org" -H ldaps://127.0.0.1:1636 
Enter LDAP Password: 
# extended LDIF
#
# LDAPv3
# base <ou=Groups,dc=example,dc=org> with scope subtree
# filter: (objectclass=*)
# requesting: ALL
#

# Groups, example.org
dn: ou=Groups,dc=example,dc=org
ou: Groups
objectClass: organizationalUnit
objectClass: top

# devel, Groups, example.org
dn: cn=devel,ou=Groups,dc=example,dc=org
cn: devel
objectClass: groupOfNames
member: uid=mcabrerizo,ou=Users,dc=example,dc=org

# search result
search: 2
result: 0 Success

# numResponses: 3
# numEntries: 2
```
