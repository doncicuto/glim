package client

var url = "https://127.0.0.1:1323" // TODO - this should not be hardcoded

var (
	fullname, email, groups, gid                string
	groupName, groupDesc, groupMembers          string
	uid, username, password, cfgFile            string
	tlscert, tlskey                             string
	manager, readonly, plainuser, passwordStdin bool
	replaceMembers, replaceMembersOf            bool
	userID, groupID                             uint32
)
