// Copyright 2019-2020 go-tcap authors. All rights reserved.
// Use of this source code is governed by a MIT-style license that can be
// found in the LICENSE file.

/*
Package tcap provides simple and painless handling of TCAP(Transaction Capabilities Application Part) in SS7/SIGTRAN protocol stack.

Though TCAP is ASN.1-based protocol, this implementation does not use any ASN.1 parser.
That makes this implementation flexible enough to create arbitrary payload with any combinations, which is useful for testing.
*/
package tcap

import (
	"encoding/binary"
	"fmt"
)

// TCAP represents a General Structure of TCAP Information Elements.
type TCAP struct {
	Transaction *Transaction
	Dialogue    *Dialogue
	Components  *Components
}

// NewBeginInvoke creates a new TCAP of type Transaction=Begin, Component=Invoke.
func NewBeginInvoke(otid uint32, invID, opCode int, payload []byte) *TCAP {
	t := &TCAP{
		Transaction: NewBegin(otid, []byte{}),
		Components:  NewComponents(NewInvoke(invID, -1, opCode, true, payload)),
	}
	t.SetLength()

	fmt.Println("NewBeginInvoke:len", t.MarshalLen())
	return t
}

// NewBeginInvokeWithDialogue creates a new TCAP of type Transaction=Begin, Component=Invoke with Dialogue Portion.
func NewBeginInvokeWithDialogue(otid uint32, dlgType, ctx, ctxver uint8, invID, opCode int, payload []byte) *TCAP {
	t := NewBeginInvoke(otid, invID, opCode, payload)
	t.Dialogue = NewDialogue(dlgType, 1, NewAARQ(1, ctx, ctxver), []byte{})
	t.SetLength()

	fmt.Println("NewBeginInvokeWithDialogue:len", t.MarshalLen())
	return t
}

// NewContinueInvoke creates a new TCAP of type Transaction=Continue, Component=Invoke.
func NewContinueInvoke(otid, dtid uint32, invID, opCode int, payload []byte) *TCAP {
	t := &TCAP{
		Transaction: NewContinue(otid, dtid, []byte{}),
		Components:  NewComponents(NewInvoke(invID, -1, opCode, true, payload)),
	}
	t.SetLength()

	return t
}

func NewContinueInvokeWithDialogue(otid, dtid uint32, invID, opCode int, dlgType, ctx, ctxver uint8, payload []byte) *TCAP {
	t := &TCAP{
		Transaction: NewContinue(otid, dtid, []byte{}),
		Components:  NewComponents(NewInvoke(invID, -1, opCode, true, payload)),
	}
	t.Dialogue = NewDialogue(dlgType, 1, NewAARE(1, ctx, ctxver, 0, 1, 0), []byte{})

	t.SetLength()

	return t
}

// NewEndInvokeWithDialogue create a new TCAP of type Transaction=End, Component=Invoke
func NewEndInvokeWithDialogue(dtid uint32, invID, opCode int, dlgType, ctx, ctxver uint8, payload []byte) *TCAP {
	t := &TCAP{
		Transaction: NewEnd(dtid, []byte{}),
		Components:  NewComponents(NewInvoke(invID, -1, opCode, true, payload)),
	}
	t.Dialogue = NewDialogue(dlgType, 1, NewAARE(1, ctx, ctxver, 0, 1, 0), []byte{})

	t.SetLength()

	return t
}

// NewEndReturnResult creates a new TCAP of type Transaction=End, Component=ReturnResult.
func NewEndReturnResult(dtid uint32, invID, opCode int, isLast bool, payload []byte) *TCAP {
	t := &TCAP{
		Transaction: NewEnd(dtid, []byte{}),
		Components:  NewComponents(NewReturnResult(invID, opCode, true, isLast, payload)),
	}
	t.SetLength()

	return t
}

func NewEndReturnError(dtid uint32, invId, errCode int, isLocal bool, param []byte) *TCAP {
	t := &TCAP{
		Transaction: NewEnd(dtid, []byte{}),
		Components:  NewComponents(NewReturnError(invId, errCode, isLocal, param)),
	}
	t.SetLength()

	return t
}

func NewEndReturnErrorWithDialogue(dtid uint32, dlgType, ctx, ctxver uint8, invId, errCode int, isLocal bool, param []byte) *TCAP {
	t := NewEndReturnError(dtid, invId, errCode, isLocal, param)
	t.Dialogue = NewDialogue(dlgType, 1, NewAARE(1, ctx, ctxver, Accepted, DialogueServiceUser, Null), []byte{})
	t.SetLength()
	return t
}

// NewEndReturnResultWithDialogue creates a new TCAP of type Transaction=End, Component=ReturnResult with Dialogue Portion.
func NewEndReturnResultWithDialogue(dtid uint32, dlgType, ctx, ctxver uint8, invID, opCode int, isLast bool, payload []byte) *TCAP {
	t := NewEndReturnResult(dtid, invID, opCode, isLast, payload)
	t.Dialogue = NewDialogue(dlgType, 1, NewAARE(1, ctx, ctxver, Accepted, DialogueServiceUser, Null), []byte{})
	t.SetLength()

	return t
}

// MarshalBinary returns the byte sequence generated from a TCAP instance.
func (t *TCAP) MarshalBinary() ([]byte, error) {
	fmt.Println("tcap:marshalbinary:len", t.MarshalLen())
	b := make([]byte, t.MarshalLen())
	if err := t.MarshalTo(b); err != nil {
		return nil, err
	}
	return b, nil
}

// MarshalTo puts the byte sequence in the byte array given as b.
func (t *TCAP) MarshalTo(b []byte) error {
	offset := 0
	fmt.Println("offset", offset)
	if portion := t.Transaction; portion != nil {
		fmt.Println("tcap:marshalto:Transaction:len", portion.MarshalLen())
		if err := portion.MarshalTo(b[offset : offset+portion.MarshalLen()]); err != nil {
			return err
		}
		offset += portion.MarshalLen()
	}

	fmt.Printf("bytes after Transaction:\n%x\n", b)
	fmt.Println("offset", offset)
	if portion := t.Dialogue; portion != nil {
		fmt.Println("tcap:marshalto:Dialogue:len", portion.MarshalLen())
		if err := portion.MarshalTo(b[offset : offset+portion.MarshalLen()]); err != nil {
			return err
		}
		offset += portion.MarshalLen()
	}
	fmt.Printf("bytes after Dialogue:\n%x\n", b)
	fmt.Println("offset", offset)

	if portion := t.Components; portion != nil {
		fmt.Println("tcap:marshalto:Components:len", portion.MarshalLen())
		if err := portion.MarshalTo(b[offset : offset+portion.MarshalLen()]); err != nil {
			return err
		}
	}
	fmt.Printf("bytes after Components:\n%x\n", b)
	fmt.Println("offset", offset)

	fmt.Println("tcap:marshalto:total length: ", len(b))
	return nil
}

// Parse parses given byte sequence as a TCAP.
func Parse(b []byte) (*TCAP, error) {
	t := &TCAP{}
	if err := t.UnmarshalBinary(b); err != nil {
		return nil, err
	}
	return t, nil
}

// UnmarshalBinary sets the values retrieved from byte sequence in a TCAP.
func (t *TCAP) UnmarshalBinary(b []byte) error {
	var err error
	offset := 0

	t.Transaction, err = ParseTransaction(b[offset:])
	if err != nil {
		return err
	}
	if len(t.Transaction.Payload) == 0 {
		return nil
	}

	switch t.Transaction.Payload[0] {
	case 0x6b:
		t.Dialogue, err = ParseDialogue(t.Transaction.Payload)
		if err != nil {
			return err
		}
		if len(t.Dialogue.Payload) == 0 {
			return nil
		}

		t.Components, err = ParseComponents(t.Dialogue.Payload)
		if err != nil {
			return err
		}
	case 0x6c:
		t.Components, err = ParseComponents(t.Transaction.Payload)
		if err != nil {
			return err
		}
	}

	return nil
}

// ParseBer parses given byte sequence as a TCAP.
//
// Deprecated: use ParseBER instead.
func ParseBer(b []byte) ([]*TCAP, error) {
	return ParseBER(b)
}

// ParseBER parses given byte sequence as a TCAP.
func ParseBER(b []byte) ([]*TCAP, error) {
	parsed, err := ParseAsBER(b)
	if err != nil {
		return nil, err
	}

	tcaps := make([]*TCAP, len(parsed))
	for i, tx := range parsed {
		t := &TCAP{
			Transaction: &Transaction{},
		}

		if err := t.Transaction.SetValsFrom(tx); err != nil {
			return nil, err
		}

		for _, dx := range tx.IE {
			switch dx.Tag {
			case 0x6b:
				t.Dialogue = &Dialogue{}
				if err := t.Dialogue.SetValsFrom(dx); err != nil {
					return nil, err
				}
			case 0x6c:
				t.Components = &Components{}
				if err := t.Components.SetValsFrom(dx); err != nil {
					return nil, err
				}
			}
		}

		tcaps[i] = t
	}

	return tcaps, nil
}

// MarshalLen returns the serial length of TCAP.
func (t *TCAP) MarshalLen() int {
	l := 0
	if portion := t.Components; portion != nil {
		l += portion.MarshalLen()
	}
	if portion := t.Dialogue; portion != nil {
		l += portion.MarshalLen()
	}
	if portion := t.Transaction; portion != nil {
		l += portion.MarshalLen()
	}
	return l
}

// SetLength sets the length in Length field.
func (t *TCAP) SetLength() {
	if portion := t.Components; portion != nil {
		portion.SetLength()
	}
	if portion := t.Dialogue; portion != nil {
		portion.SetLength()
	}
	if portion := t.Transaction; portion != nil {
		portion.SetLength()
		if c := t.Components; c != nil {
			portion.Length += uint8(c.MarshalLen())
		}
		if d := t.Dialogue; d != nil {
			portion.Length += uint8(d.MarshalLen())
		}
	}
}

// OTID returns the TCAP Originating Transaction ID in Transaction Portion in uint32.
func (t *TCAP) OTID() uint32 {
	if ts := t.Transaction; ts != nil {
		if otid := ts.OrigTransactionID; otid != nil {
			return binary.BigEndian.Uint32(otid.Value)
		}
	}

	return 0
}

// DTID returns the TCAP Originating Transaction ID in Transaction Portion in uint32.
func (t *TCAP) DTID() uint32 {
	if ts := t.Transaction; ts != nil {
		if dtid := ts.DestTransactionID; dtid != nil {
			return binary.BigEndian.Uint32(dtid.Value)
		}
	}

	return 0
}

// AppContextName returns the ACN in string.
func (t *TCAP) AppContextName() string {
	if d := t.Dialogue; d != nil {
		return d.Context()
	}

	return ""
}

// AppContextNameWithVersion returns the ACN with ACN Version in string.
//
// TODO: Looking for a better way to return the value in the same format...
func (t *TCAP) AppContextNameWithVersion() string {
	if d := t.Dialogue; d != nil {
		return d.Context() + "-v" + d.ContextVersion()
	}

	return ""
}

// AppContextNameOid returns the ACN with ACN Version in OID formatted string.
//
// TODO: Looking for a better way to return the value in the same format...
func (t *TCAP) AppContextNameOid() string {
	if r := t.Dialogue; r != nil {
		if rp := r.DialoguePDU; rp != nil {
			oid := "0."
			for i, x := range rp.ApplicationContextName.Value[2:] {
				oid += fmt.Sprint(x)
				if i <= 6 {
					break
				}
				oid += "."
			}
			return oid
		}
	}

	return ""
}

// ComponentType returns the ComponentType in Component Portion in the list of string.
//
// The returned value is of type []string, as it may have multiple Components.
func (t *TCAP) ComponentType() []string {
	if c := t.Components; c != nil {
		var iids []string
		for _, cm := range c.Component {
			iids = append(iids, cm.ComponentTypeString())
		}
		return iids
	}

	return nil
}

// InvokeID returns the InvokeID in Component Portion in the list of string.
//
// The returned value is of type []string, as it may have multiple Components.
func (t *TCAP) InvokeID() []uint8 {
	if c := t.Components; c != nil {
		var iids []uint8
		for _, cm := range c.Component {
			iids = append(iids, cm.InvID())
		}

		return iids
	}

	return nil
}

// OpCode returns the OpCode in Component Portion in the list of string.
//
// The returned value is of type []string, as it may have multiple Components.
func (t *TCAP) OpCode() []uint8 {
	if c := t.Components; c != nil {
		var ops []uint8
		for _, cm := range c.Component {
			ops = append(ops, cm.OpCode())
		}

		return ops
	}

	return nil
}

// LayerPayload returns the upper layer as byte slice.
//
// The returned value is of type [][]byte, as it may have multiple Components.
func (t *TCAP) LayerPayload() [][]byte {
	if c := t.Components; c != nil {
		var ret [][]byte
		for _, cm := range c.Component {
			ret = append(ret, cm.Parameter.Value)
		}

		return ret
	}

	return nil
}

// String returns TCAP in human readable string.
func (t *TCAP) String() string {
	return fmt.Sprintf("{Transaction: %v, Dialogue: %v, Components: %v}",
		t.Transaction,
		t.Dialogue,
		t.Components,
	)
}
