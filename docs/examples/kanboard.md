# Kanboard - LDAP authentication using proxy

This page shows how you can configure Kanboard to authenticate users with Glim using proxy authentication. [Kanboard documentation](https://docs.kanboard.org/en/1.2.21/admin_guide/ldap_examples.html) has many examples for LDAP configurations. Here we offer a full example with our suggestions.

Kanboard has a config.php file in /var/www/app. We've added the following entries to this file assuming that we've already created a kanboard-admins and manager groups in Glim as these groups will be used to provide admin or project manajer roles to your users.

```(php)
<?php

defined('ENABLE_URL_REWRITE') or define('ENABLE_URL_REWRITE', true);
defined('LOG_DRIVER') or define('LOG_DRIVER', 'system');

define('LDAP_BIND_TYPE', 'proxy');
define('LDAP_USERNAME', 'cn=search,dc=example,dc=org');
define('LDAP_PASSWORD', 'test');
define('LDAP_AUTH', true);
define('LDAP_SERVER', 'ldaps://192.168.1.136:1636');
define('LDAP_SSL_VERIFY', false);
define('LDAP_START_TLS', false);
define('LDAP_USER_BASE_DN', 'ou=Users,dc=example,dc=org');
define('LDAP_USER_FILTER', 'uid=%s');
define('LDAP_GROUP_ADMIN_DN', 'cn=kanboard-admins,ou=Groups,dc=example,dc=org');
define('LDAP_GROUP_MANAGER_DN', 'cn=manager,ou=Groups,dc=example,dc=org');
define('LDAP_GROUP_PROVIDER', true);
define('LDAP_GROUP_BASE_DN', 'ou=Groups,dc=example,dc=org');
define('LDAP_GROUP_FILTER', '(&(objectClass=groupOfNames)(cn=%s*))');
define('LDAP_GROUP_ATTRIBUTE_NAME', 'cn');
?>
```

Sample log showing successful authentication:

```(bash)
2022-07-05T17:56:00+02:00 [LDAP] ⇨ bind requested by client: 172.20.0.2:42882
2022-07-05T17:56:00+02:00 [LDAP] ⇨ bind protocol version: 3 client 172.20.0.2:42882
2022-07-05T17:56:00+02:00 [LDAP] ⇨ bind name: cn=search,dc=example,dc=org client 172.20.0.2:42882
2022-07-05T17:56:00+02:00 [LDAP] ⇨ bind password: ********** client 172.20.0.2:42882
2022-07-05T17:56:00+02:00 [LDAP] ⇨ success: valid credentials provided
2022-07-05T17:56:00+02:00 [LDAP] ⇨ search requested by client 172.20.0.2:42882
2022-07-05T17:56:00+02:00 [LDAP] ⇨ search base object: ou=Users,dc=example,dc=org
2022-07-05T17:56:00+02:00 [LDAP] ⇨ search scope: wholeSubtree
2022-07-05T17:56:00+02:00 [LDAP] ⇨ search maximum number of entries to be returned (0 - No limit restriction): 0
2022-07-05T17:56:00+02:00 [LDAP] ⇨ search maximum time limit (0 - No limit restriction): 1
2022-07-05T17:56:00+02:00 [LDAP] ⇨ search show types only: false
2022-07-05T17:56:00+02:00 [LDAP] ⇨ search filter: (uid=mcabrerizo)
2022-07-05T17:56:00+02:00 [LDAP] ⇨ search attributes: uid cn mail memberof
2022-07-05T17:56:00+02:00 [LDAP] ⇨ bind requested by client: 172.20.0.2:42882
2022-07-05T17:56:00+02:00 [LDAP] ⇨ bind protocol version: 3 client 172.20.0.2:42882
2022-07-05T17:56:00+02:00 [LDAP] ⇨ bind name: uid=mcabrerizo,ou=Users,dc=example,dc=org client 172.20.0.2:42882
2022-07-05T17:56:00+02:00 [LDAP] ⇨ bind password: ********** client 172.20.0.2:42882
2022-07-05T17:56:00+02:00 [LDAP] ⇨ success: valid credentials provided
2022-07-05T17:56:00+02:00 [LDAP] ⇨ unbind requested by client: 172.20.0.2:42882
2022-07-05T17:56:00+02:00 [LDAP] ⇨ connection closed by client 172.20.0.2:42882
```
