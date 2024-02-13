// Copyright 2019-2020 go-tcap authors. All rights reserved.
// Use of this source code is governed by a MIT-style license that can be
// found in the LICENSE file.

package tcap

import (
	"encoding/binary"
	"fmt"
)

// Message Type definitions.
const (
	Unidirectional int = iota + 1
	Begin
	_
	End
	Continue
	_
	Abort
)

// Abort Cause definitions.
const (
	UnrecognizedMessageType uint8 = iota
	UnrecognizedTransactionID
	BadlyFormattedTransactionPortion
	IncorrectTransactionPortion
	ResourceLimitation
)

// Transaction represents a Transaction Portion of TCAP.
type Transaction struct {
	Type              Tag
	Length            uint8
	OrigTransactionID *IE
	DestTransactionID *IE
	PAbortCause       *IE
	Payload           []byte
}

// NewTransaction returns a new Transaction Portion.
func NewTransaction(mtype int, otid, dtid uint32, cause uint8, payload []byte) *Transaction {
	t := &Transaction{
		Type: NewApplicationWideConstructorTag(mtype),
		OrigTransactionID: &IE{
			Tag:   NewApplicationWidePrimitiveTag(8),
			Value: make([]byte, 4),
		},
		DestTransactionID: &IE{
			Tag:   NewApplicationWidePrimitiveTag(9),
			Value: make([]byte, 4),
		},
		PAbortCause: &IE{
			Tag:   NewApplicationWidePrimitiveTag(10),
			Value: []byte{cause},
		},
		Payload: payload,
	}
	binary.BigEndian.PutUint32(t.OrigTransactionID.Value, otid)
	binary.BigEndian.PutUint32(t.DestTransactionID.Value, dtid)
	t.SetLength()

	return t
}

// NewUnidirectional returns Unidirectional type of Transacion Portion.
func NewUnidirectional(payload []byte) *Transaction {
	t := NewTransaction(
		Unidirectional, // Type: Unidirectional
		0,              // otid
		0,              // dtid
		0,              // cause
		payload,        // payload
	)
	t.OrigTransactionID = nil
	t.DestTransactionID = nil
	t.PAbortCause = nil
	return t
}

// NewBegin returns Begin type of Transacion Portion.
func NewBegin(otid uint32, payload []byte) *Transaction {
	t := &Transaction{
		Type: NewApplicationWideConstructorTag(Begin),
		OrigTransactionID: &IE{
			Tag:   NewApplicationWidePrimitiveTag(8),
			Value: make([]byte, 4),
		},
		Payload: payload,
	}
	binary.BigEndian.PutUint32(t.OrigTransactionID.Value, otid)
	t.SetLength()

	return t
}

// NewEnd returns End type of Transacion Portion.
func NewEnd(otid uint32, payload []byte) *Transaction {
	t := &Transaction{
		Type: NewApplicationWideConstructorTag(End),
		DestTransactionID: &IE{
			Tag:   NewApplicationWidePrimitiveTag(9),
			Value: make([]byte, 4),
		},
		Payload: payload,
	}
	binary.BigEndian.PutUint32(t.DestTransactionID.Value, otid)
	t.SetLength()

	return t
}

// NewContinue returns Continue type of Transacion Portion.
func NewContinue(otid, dtid uint32, payload []byte) *Transaction {
	t := NewTransaction(
		Continue, // Type: Continue
		otid,     // otid
		dtid,     // dtid
		0,        // cause
		payload,  // payload
	)
	t.PAbortCause = nil
	return t
}

// NewAbort returns Abort type of Transacion Portion.
func NewAbort(dtid uint32, cause uint8, payload []byte) *Transaction {
	t := NewTransaction(
		Abort,   // Type: Abort
		0,       // otid
		dtid,    // dtid
		cause,   // cause
		payload, // payload
	)
	t.OrigTransactionID = nil
	return t
}

// NewContinueReturnResult creates a new TCAP of type Transaction=Continue, Component=ReturnResult.
func NewContinueReturnResult(otid, dtid uint32, invID, opCode int, payload []byte) *TCAP {
	t := &TCAP{
		Transaction: NewContinue(otid, dtid, []byte{}),
		Components:  NewComponents(NewReturnResult(invID, opCode, true, true, payload)),
	}
	t.SetLength()

	return t
}

// MarshalBinary returns the byte sequence generated from a Transaction instance.
func (t *Transaction) MarshalBinary() ([]byte, error) {
	b := make([]byte, t.MarshalLen())
	if err := t.MarshalTo(b); err != nil {
		return nil, err
	}
	return b, nil
}

// MarshalTo puts the byte sequence in the byte array given as b.
func (t *Transaction) MarshalTo(b []byte) error {
	var offset int = 2
	b[0] = uint8(t.Type)
	// if t.Length > 127 {
	// 	buf := make([]byte, 4)
	// 	t.Length = t.Length - 1
	// 	var count int
	// 	if (int64(t.Length) & int64(-16777216)) > 0 {
	// 		buf[0] = byte(t.Length >> 24 & 255)
	// 		buf[1] = byte(t.Length >> 16 & 255)
	// 		buf[2] = byte(t.Length >> 8 & 255)
	// 		buf[3] = byte(t.Length & 255)
	// 		count = 4
	// 	} else if (int64(t.Length) & 16711680) > 0 {
	// 		buf[0] = byte(t.Length >> 16 & 255)
	// 		buf[1] = byte(t.Length >> 8 & 255)
	// 		buf[2] = byte(t.Length & 255)
	// 		count = 3
	//
	// 	} else if (int64(t.Length) & 65280) > 0 {
	// 		buf[0] = byte(t.Length >> 8 & 255)
	// 		buf[1] = byte(t.Length & 255)
	// 		count = 2
	// 	} else {
	// 		buf[0] = byte(t.Length & 255)
	// 		count = 1
	// 	}
	//
	// 	b[offset-1] = byte(128 | count)
	// 	for i := 0; i < count; i++ {
	// 		b[offset+i] = buf[i]
	// 	}
	// 	offset = offset + count
	//
	// } else {
	// 	b[1] = t.Length
	// 	offset = 2
	// }

	// b[1] = t.Length
	// offset = 2
	//
	//
	//
	//
	offset = writeLength(b, t.Length)

	switch t.Type.Code() {
	case Unidirectional:
		break
	case Begin:
		if field := t.OrigTransactionID; field != nil {
			if err := field.MarshalTo(b[offset : offset+field.MarshalLen()]); err != nil {
				return err
			}
			offset += field.MarshalLen()
		}
	case End:
		if field := t.DestTransactionID; field != nil {
			if err := field.MarshalTo(b[offset : offset+field.MarshalLen()]); err != nil {
				return err
			}
			offset += field.MarshalLen()
		}
	case Continue:
		if field := t.OrigTransactionID; field != nil {
			if err := field.MarshalTo(b[offset : offset+field.MarshalLen()]); err != nil {
				return err
			}
			offset += field.MarshalLen()
		}

		if field := t.DestTransactionID; field != nil {
			if err := field.MarshalTo(b[offset : offset+field.MarshalLen()]); err != nil {
				return err
			}
			offset += field.MarshalLen()
		}
	case Abort:
		if field := t.DestTransactionID; field != nil {
			if err := field.MarshalTo(b[offset : offset+field.MarshalLen()]); err != nil {
				return err
			}
			offset += field.MarshalLen()
		}

		if field := t.PAbortCause; field != nil {
			if err := field.MarshalTo(b[offset : offset+field.MarshalLen()]); err != nil {
				return err
			}
			offset += field.MarshalLen()
		}
	}
	copy(b[offset:t.MarshalLen()], t.Payload)
	return nil
}

// ParseTransaction parses given byte sequence as an Transaction.
func ParseTransaction(b []byte) (*Transaction, error) {
	t := &Transaction{}
	if err := t.UnmarshalBinary(b); err != nil {
		return nil, err
	}
	return t, nil
}

// func readLength(b []byte) int{
// 	var length int
// 	r := bytes.NewReader(b[1:])
// 	lengthByte, _ := r.ReadByte()
// 	if((lengthByte & 128) == 0){
// 		return int(lengthByte)
// 	} else {
// 		lengthByte = (lengthByte & 127)
// 		if(lengthByte == 0){
// 			return -1
// 		} else {
// 			for i := 0; i < int(lengthByte); i++ {
// 				tmp, _ := r.ReadByte()
// 				length = int(byte(length) << 8 | 255 & tmp)
// 			}
// 			return length
// 		}
// 	}
// }

// UnmarshalBinary sets the values retrieved from byte sequence in an Transaction.
func (t *Transaction) UnmarshalBinary(b []byte) error {
	t.Type = Tag(b[0])

	u, _ := readLength(b)
	t.Length = u

	var err error
	offset := 2
	if t.Length > 127 {
		offset = 3
	}

	switch t.Type.Code() {
	case Unidirectional:
		break
	case Begin:
		t.OrigTransactionID, err = ParseIE(b[offset : offset+6])
		if err != nil {
			return err
		}
		offset += t.OrigTransactionID.MarshalLen()
	case End:
		t.DestTransactionID, err = ParseIE(b[offset : offset+6])
		if err != nil {
			return err
		}
		offset += t.DestTransactionID.MarshalLen()
	case Continue:
		t.OrigTransactionID, err = ParseIE(b[offset : offset+6])
		if err != nil {
			return err
		}
		offset += t.OrigTransactionID.MarshalLen()
		t.DestTransactionID, err = ParseIE(b[offset : offset+6])
		if err != nil {
			return err
		}
		offset += t.DestTransactionID.MarshalLen()
	case Abort:
		t.DestTransactionID, err = ParseIE(b[offset : offset+6])
		if err != nil {
			return err
		}
		offset += t.DestTransactionID.MarshalLen()

		//t.PAbortCause, err = ParseIE(b[offset : ])
		//if err != nil {
		//	return err
		//}
		//t.PAbortCause.IE, _ = ParseAsBER(t.PAbortCause.Value)
		//offset += t.PAbortCause.MarshalLen()
	}
	t.Payload = b[offset:]
	return nil
}

// SetValsFrom sets the values from IE parsed by ParseBER.
func (t *Transaction) SetValsFrom(berParsed *IE) error {
	t.Type = berParsed.Tag
	t.Length = berParsed.Length
	for _, ie := range berParsed.IE {
		switch ie.Tag {
		case 0x48:
			t.OrigTransactionID = ie
		case 0x49:
			t.DestTransactionID = ie
		case 0x4a:
			t.PAbortCause = ie
		}
	}
	return nil
}

// MarshalLen returns the serial length of Transaction.
func (t *Transaction) MarshalLen() int {
	l := 0
	switch t.Type.Code() {
	case Unidirectional:
		break
	case Begin:
		if field := t.OrigTransactionID; field != nil {
			l += field.MarshalLen()
		}
	case End:
		if field := t.DestTransactionID; field != nil {
			l += field.MarshalLen()
		}
	case Continue:
		if field := t.OrigTransactionID; field != nil {
			l += field.MarshalLen()
		}
		if field := t.DestTransactionID; field != nil {
			l += field.MarshalLen()
		}
	case Abort:
		if field := t.DestTransactionID; field != nil {
			l += field.MarshalLen()
		}
		if field := t.PAbortCause; field != nil {
			l += field.MarshalLen()
		}
	}
	l += len(t.Payload)
	if t.Length > 127 {
		return l + 3
	} else {
		return l + 2
	}
}

// SetLength sets the length in Length field.
func (t *Transaction) SetLength() {
	if field := t.OrigTransactionID; field != nil {
		field.SetLength()
	}
	if field := t.DestTransactionID; field != nil {
		field.SetLength()
	}
	if field := t.PAbortCause; field != nil {
		field.SetLength()
	}
	t.Length = uint8(t.MarshalLen() - 2)
}

// MessageTypeString returns the name of Message Type in string.
func (t *Transaction) MessageTypeString() string {
	switch t.Type.Code() {
	case Unidirectional:
		return "Unidirectional"
	case Begin:
		return "Begin"
	case End:
		return "End"
	case Continue:
		return "Continue"
	case Abort:
		return "Abort"
	}
	return ""
}

// OTID returns the OrigTransactionID in string.
func (t *Transaction) OTID() string {
	switch t.Type.Code() {
	case Begin, Continue:
		if field := t.OrigTransactionID; field != nil {
			return fmt.Sprintf("%04x", field.Value)
		}
	}
	return ""
}

// DTID returns the DestTransactionID in string.
func (t *Transaction) DTID() string {
	switch t.Type.Code() {
	case End, Continue, Abort:
		if field := t.DestTransactionID; field != nil {
			return fmt.Sprintf("%04x", field.Value)
		}
	}
	return ""
}

// AbortCause returns the P-Abort Cause in string.
func (t *Transaction) AbortCause() string {
	cause := t.PAbortCause
	if cause == nil {
		return ""
	}

	if t.Type.Code() == Abort {
		switch t.PAbortCause.Value[0] {
		case UnrecognizedMessageType:
			return "UnrecognizedMessageType"
		case UnrecognizedTransactionID:
			return "UnrecognizedTransactionID"
		case BadlyFormattedTransactionPortion:
			return "BadlyFormattedTransactionPortion"
		case IncorrectTransactionPortion:
			return "IncorrectTransactionPortion"
		case ResourceLimitation:
			return "ResourceLimitation"
		}
	}
	return ""
}

// String returns Transaction in human readable string.
func (t *Transaction) String() string {
	return fmt.Sprintf("{Type: %#x, Length: %d, OrigTransactionID: %v, DestTransactionID: %v, PAbortCause: %v, Payload: %x}",
		t.Type,
		t.Length,
		t.OrigTransactionID,
		t.DestTransactionID,
		t.PAbortCause,
		t.Payload,
	)
}
