package isbn

import (
	_ "embed"
	"encoding/xml"
	"fmt"
)

const (
	// rangeMessageURL is the canonocal URL for the ISBN range message XML file.
	rangeMessageURL = "https://www.isbn-international.org/export_rangemessage.xml"
)

// ISBNRangeMessage represents the structure of an ISBN range message XML file.
type ISBNRangeMessage struct {
	XMLName             xml.Name           `xml:"ISBNRangeMessage"`
	MessageSource       string             `xml:"MessageSource,omitempty"`
	MessageSerialNumber string             `xml:"MessageSerialNumber,omitempty"`
	MessageDate         string             `xml:"MessageDate"`
	EANUCCPrefixes      EANUCCPrefixes     `xml:"EAN.UCCPrefixes"`
	RegistrationGroups  RegistrationGroups `xml:"RegistrationGroups"`
}

// EANUCCPrefixes represents the EAN.UCC prefixes in an ISBN range message.
type EANUCCPrefixes struct {
	EANUCC []EANUCC `xml:"EAN.UCC"`
}

// EANUCC represents an EAN.UCC in an ISBN range message.
type EANUCC struct {
	Prefix string `xml:"Prefix"`
	Agency string `xml:"Agency"`
	Rules  Rules  `xml:"Rules"`
}

// RegistrationGroups represents the registration groups in an ISBN range message.
type RegistrationGroups struct {
	Group []Group `xml:"Group"`
}

// Group represents a group in an ISBN range message.
type Group struct {
	Prefix string `xml:"Prefix"`
	Agency string `xml:"Agency"`
	Rules  Rules  `xml:"Rules"`
}

// Rules represents the rules in an ISBN range message.
type Rules struct {
	Rule []Rule `xml:"Rule"`
}

// Rule represents a rule in an ISBN range message.
type Rule struct {
	Range  string `xml:"Range"`
	Length string `xml:"Length"`
}

// ParseRangeMessageXML parses the XML data from an ISBN range message document into the
// corresponding ISBNRangeMessage struct.
func ParseRangeMessageXML(data []byte) (ISBNRangeMessage, error) {
	var result ISBNRangeMessage
	if err := xml.Unmarshal(data, &result); err != nil {
		return ISBNRangeMessage{}, fmt.Errorf("parse range message: %w", err)
	}
	return result, nil
}

// ISBNRangesFromXML converts an ISBNRangeMessage to an ISBNRanges.
func ISBNRangesFromXML(msg ISBNRangeMessage) ISBNRanges {
	var out ISBNRanges
	out.ISBNRangeMessage.MessageDate = msg.MessageDate
	for _, g := range msg.RegistrationGroups.Group {
		gr := GroupRule{Prefix: g.Prefix}
		for _, r := range g.Rules.Rule {
			gr.Rules.Rule = append(gr.Rules.Rule, RangeRule{Range: r.Range, Length: r.Length})
		}
		out.ISBNRangeMessage.RegistrationGroups.Group = append(out.ISBNRangeMessage.RegistrationGroups.Group, gr)
	}
	return out
}
