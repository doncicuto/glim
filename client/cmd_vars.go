/*
Copyright © 2020 Miguel Ángel Álvarez Cabrerizo <mcabrerizo@arrakis.ovh>

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package client

var (
	fullname, email, groups, gid                      string
	groupName, groupDesc, groupMembers                string
	uid, username, password, cfgFile                  string
	tlscacert, tlscert, tlskey                        string
	serverAddress, ldapAddress, restAddress           string
	badgerKV                                          string
	manager, readonly, plainuser, passwordStdin       bool
	replaceMembers, replaceMembersOf, removeMembersOf bool
	userID, groupID                                   uint32
)
