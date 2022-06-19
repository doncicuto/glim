/*
Copyright © 2022 Miguel Ángel Álvarez Cabrerizo <mcabrerizo@arrakis.ovh>

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

package ldap

import (
	"errors"

	ber "github.com/go-asn1-ber/asn1-ber"
)

// Message - TODO comment
type Message struct {
	ID      int64
	Op      int64
	Request []*ber.Packet
}

func messageID(p *ber.Packet) (int64, error) {
	if p.ClassType != ber.ClassUniversal ||
		p.TagType != ber.TypePrimitive ||
		p.Tag != ber.TagInteger ||
		len(p.Children) != 0 {
		return 0, errors.New("wrong message id definition")
	}

	id, err := ber.ParseInt64(p.ByteValue)
	if err != nil {
		return 0, errors.New("could not parse message id")
	}

	return id, nil
}

func protocolOp(p *ber.Packet) (int64, error) {
	if p.ClassType != ber.ClassApplication {
		return 0, errors.New("wrong protocol operation definition")
	}

	return int64(p.Tag), nil
}

func protocolVersion(p *ber.Packet) (int64, *ServerError) {
	if p.ClassType != ber.ClassUniversal ||
		p.TagType != ber.TypePrimitive ||
		p.Tag != ber.TagInteger ||
		len(p.Children) != 0 {
		return 0, &ServerError{
			Msg:  "wrong protocol version definition",
			Code: ProtocolError,
		}
	}

	v, err := ber.ParseInt64(p.ByteValue)
	if err != nil {
		return 0, &ServerError{
			Msg:  "could not parse protocol version",
			Code: Other,
		}
	}

	if v != Version3 {
		return 0, &ServerError{
			Msg:  "historical protocol version requested, use LDAPv3 instead",
			Code: ProtocolError,
		}
	}
	return v, nil
}

func requestName(p *ber.Packet) (string, *ServerError) {
	if p.ClassType != ber.ClassContext ||
		p.TagType != ber.TypePrimitive ||
		p.Tag != ber.TagEOC {
		return "", &ServerError{
			Msg:  "wrong extended request name definition",
			Code: ProtocolError,
		}
	}
	return p.Data.String(), nil
}

//DecodeMessage - TODO comment
func DecodeMessage(p *ber.Packet) (*Message, error) {
	message := new(Message)

	if p.ClassType != ber.ClassUniversal ||
		p.TagType != ber.TypeConstructed ||
		p.Tag != ber.TagSequence ||
		len(p.Children) < 2 {
		return nil, errors.New("wrong ASN.1 Envelope")
	}

	id, err := messageID(p.Children[0])
	if err != nil {
		return nil, err
	}
	message.ID = id

	op, err := protocolOp(p.Children[1])
	if err != nil {
		return nil, err
	}
	message.Op = op

	message.Request = p.Children[1].Children

	return message, nil
}
