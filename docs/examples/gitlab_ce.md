# Gitlab Community Edition

This page shows how you can configure Gitlab Community Edition to authenticate users with Glim. [Gitlab documentation](https://docs.gitlab.com/ee/administration/auth/ldap/) provides full information about LDAP configuration. Here we offer a full example with our suggestions.

In our example, we'll use the following settings uncommenting and editing the /etc/gitlab/gitlab.rb file. Variable names are fully-explanatory. Once we finish editing the configuration file we restart Gitlab to use Glim as our LDAP authentication server.

```(bash)
...
gitlab_rails['ldap_enabled'] = true
# gitlab_rails['prevent_ldap_sign_in'] = false

###! **remember to close this block with 'EOS' below**
gitlab_rails['ldap_servers'] = YAML.load <<-'EOS'
   main: # 'main' is the GitLab 'provider ID' of this LDAP server
     label: 'LDAP'
     host: '192.168.1.136'
     port: 1636
     uid: 'uid'
     encryption: 'simple_tls' # "start_tls" or "simple_tls" or "plain"
     verify_certificates: false
     smartcard_auth: false
     active_directory: false
     allow_username_or_email_login: false
     lowercase_usernames: false
     block_auto_created_users: false
     base: 'ou=Users,dc=example,dc=org'
EOS
...
```

Sample log showing successful authentication, user information retrieval, getting groups...:

```(bash)
2022-07-05T20:03:19+02:00 [LDAP] ⇨ serving LDAPS connection from 172.22.0.2:38980
2022-07-05T20:03:19+02:00 [LDAP] ⇨ bind requested by client: 172.22.0.2:38980
2022-07-05T20:03:19+02:00 [LDAP] ⇨ bind protocol version: 3 client 172.22.0.2:38980
2022-07-05T20:03:19+02:00 [LDAP] ⇨ bind name: cn=search,dc=example,dc=org client 172.22.0.2:38980
2022-07-05T20:03:19+02:00 [LDAP] ⇨ bind password: ********** client 172.22.0.2:38980
2022-07-05T20:03:19+02:00 [LDAP] ⇨ success: valid credentials provided
2022-07-05T20:03:19+02:00 [LDAP] ⇨ search requested by client 172.22.0.2:38980
2022-07-05T20:03:19+02:00 [LDAP] ⇨ search base object: 
2022-07-05T20:03:19+02:00 [LDAP] ⇨ wrong domain
2022-07-05T20:03:19+02:00 [LDAP] ⇨ search requested by client 172.22.0.2:38980
2022-07-05T20:03:19+02:00 [LDAP] ⇨ search base object: ou=users,dc=example,dc=org
2022-07-05T20:03:19+02:00 [LDAP] ⇨ search scope: wholeSubtree
2022-07-05T20:03:19+02:00 [LDAP] ⇨ search maximum number of entries to be returned (0 - No limit restriction): 1
2022-07-05T20:03:19+02:00 [LDAP] ⇨ search maximum time limit (0 - No limit restriction): 0
2022-07-05T20:03:19+02:00 [LDAP] ⇨ search show types only: false
2022-07-05T20:03:19+02:00 [LDAP] ⇨ search filter: (uid=mcabrerizo)
2022-07-05T20:03:19+02:00 [LDAP] ⇨ search attributes: ALL
2022-07-05T20:03:20+02:00 [LDAP] ⇨ bind requested by client: 172.22.0.2:38980
2022-07-05T20:03:20+02:00 [LDAP] ⇨ bind protocol version: 3 client 172.22.0.2:38980
2022-07-05T20:03:20+02:00 [LDAP] ⇨ bind name: uid=mcabrerizo,ou=Users,dc=example,dc=org client 172.22.0.2:38980
2022-07-05T20:03:20+02:00 [LDAP] ⇨ bind password: ********** client 172.22.0.2:38980
2022-07-05T20:03:20+02:00 [LDAP] ⇨ success: valid credentials provided
2022-07-05T20:03:21+02:00 [LDAP] ⇨ connection closed by client 172.22.0.2:38980
2022-07-05T20:03:24+02:00 [LDAP] ⇨ serving LDAPS connection from 172.22.0.2:55158
2022-07-05T20:03:24+02:00 [LDAP] ⇨ bind requested by client: 172.22.0.2:55158
2022-07-05T20:03:24+02:00 [LDAP] ⇨ bind protocol version: 3 client 172.22.0.2:55158
2022-07-05T20:03:24+02:00 [LDAP] ⇨ bind name: cn=search,dc=example,dc=org client 172.22.0.2:55158
2022-07-05T20:03:24+02:00 [LDAP] ⇨ bind password: ********** client 172.22.0.2:55158
2022-07-05T20:03:24+02:00 [LDAP] ⇨ success: valid credentials provided
2022-07-05T20:03:24+02:00 [LDAP] ⇨ search requested by client 172.22.0.2:55158
2022-07-05T20:03:24+02:00 [LDAP] ⇨ search base object: 
2022-07-05T20:03:24+02:00 [LDAP] ⇨ wrong domain
2022-07-05T20:03:24+02:00 [LDAP] ⇨ search requested by client 172.22.0.2:55158
2022-07-05T20:03:24+02:00 [LDAP] ⇨ search base object: uid=mcabrerizo,ou=users,dc=example,dc=org
2022-07-05T20:03:24+02:00 [LDAP] ⇨ search scope: baseObject
2022-07-05T20:03:24+02:00 [LDAP] ⇨ search maximum number of entries to be returned (0 - No limit restriction): 0
2022-07-05T20:03:24+02:00 [LDAP] ⇨ search maximum time limit (0 - No limit restriction): 0
2022-07-05T20:03:24+02:00 [LDAP] ⇨ search show types only: false
2022-07-05T20:03:24+02:00 [LDAP] ⇨ search filter: (objectClass=*)
2022-07-05T20:03:24+02:00 [LDAP] ⇨ search attributes: dn uid cn mail email userPrincipalName sAMAccountName userid
2022-07-05T20:03:24+02:00 [LDAP] ⇨ connection closed by client 172.22.0.2:55158
```
