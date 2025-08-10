package dns

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"strings"
)

// Parser implements the DNSParser interface
type Parser struct{}

// NewParser creates a new DNS parser
func NewParser() DNSParser {
	return &Parser{}
}

// ParseQuery parses a DNS query from raw bytes
func (p *Parser) ParseQuery(data []byte) (*DNSQuery, error) {
	if len(data) < 12 {
		return nil, ErrShortMessage
	}
	
	// Parse header
	id := binary.BigEndian.Uint16(data[0:2])
	flags := binary.BigEndian.Uint16(data[2:4])
	qdcount := binary.BigEndian.Uint16(data[4:6])
	
	// Verify this is a query
	if flags&FlagQR != 0 {
		return nil, ErrInvalidQuery
	}
	
	// We only support single questions for now
	if qdcount != 1 {
		return nil, ErrInvalidQuery
	}
	
	// Parse question section
	question, err := p.parseQuestion(data[12:])
	if err != nil {
		return nil, fmt.Errorf("failed to parse question: %w", err)
	}
	
	return &DNSQuery{
		ID:       id,
		Question: *question,
	}, nil
}

// SerializeResponse serializes a DNS response to raw bytes
func (p *Parser) SerializeResponse(response *DNSResponse) ([]byte, error) {
	var buf bytes.Buffer
	
	// Write header
	binary.Write(&buf, binary.BigEndian, response.ID)
	
	// Set flags (QR=1 for response, RA=1 for recursion available)
	flags := uint16(FlagQR | FlagRA)
	if response.ResponseCode == RCodeNoError && len(response.Answers) > 0 {
		flags |= FlagAA // Set authoritative answer for successful responses
	}
	binary.Write(&buf, binary.BigEndian, flags)
	
	// Write counts
	binary.Write(&buf, binary.BigEndian, uint16(1)) // QDCOUNT
	binary.Write(&buf, binary.BigEndian, uint16(len(response.Answers))) // ANCOUNT
	binary.Write(&buf, binary.BigEndian, uint16(len(response.Authorities))) // NSCOUNT
	binary.Write(&buf, binary.BigEndian, uint16(len(response.Additional))) // ARCOUNT
	
	// Write question section
	if err := p.writeQuestion(&buf, response.Question); err != nil {
		return nil, fmt.Errorf("failed to write question: %w", err)
	}
	
	// Write answer sections
	for _, record := range response.Answers {
		if err := p.writeRecord(&buf, record); err != nil {
			return nil, fmt.Errorf("failed to write answer record: %w", err)
		}
	}
	
	for _, record := range response.Authorities {
		if err := p.writeRecord(&buf, record); err != nil {
			return nil, fmt.Errorf("failed to write authority record: %w", err)
		}
	}
	
	for _, record := range response.Additional {
		if err := p.writeRecord(&buf, record); err != nil {
			return nil, fmt.Errorf("failed to write additional record: %w", err)
		}
	}
	
	return buf.Bytes(), nil
}

// ParseResponse parses a DNS response from raw bytes
func (p *Parser) ParseResponse(data []byte) (*DNSResponse, error) {
	if len(data) < 12 {
		return nil, ErrShortMessage
	}
	
	// Parse header
	id := binary.BigEndian.Uint16(data[0:2])
	flags := binary.BigEndian.Uint16(data[2:4])
	qdcount := binary.BigEndian.Uint16(data[4:6])
	ancount := binary.BigEndian.Uint16(data[6:8])
	nscount := binary.BigEndian.Uint16(data[8:10])
	arcount := binary.BigEndian.Uint16(data[10:12])
	
	// Verify this is a response
	if flags&FlagQR == 0 {
		return nil, ErrInvalidQuery
	}
	
	response := &DNSResponse{
		ID:           id,
		ResponseCode: uint8(flags & 0x0F),
	}
	
	offset := 12
	
	// Parse question section
	if qdcount > 0 {
		question, newOffset, err := p.parseQuestionAt(data, offset)
		if err != nil {
			return nil, fmt.Errorf("failed to parse question: %w", err)
		}
		response.Question = *question
		offset = newOffset
	}
	
	// Parse answer records
	for i := 0; i < int(ancount); i++ {
		record, newOffset, err := p.parseRecordAt(data, offset)
		if err != nil {
			return nil, fmt.Errorf("failed to parse answer record: %w", err)
		}
		response.Answers = append(response.Answers, *record)
		offset = newOffset
	}
	
	// Parse authority records
	for i := 0; i < int(nscount); i++ {
		record, newOffset, err := p.parseRecordAt(data, offset)
		if err != nil {
			return nil, fmt.Errorf("failed to parse authority record: %w", err)
		}
		response.Authorities = append(response.Authorities, *record)
		offset = newOffset
	}
	
	// Parse additional records
	for i := 0; i < int(arcount); i++ {
		record, newOffset, err := p.parseRecordAt(data, offset)
		if err != nil {
			return nil, fmt.Errorf("failed to parse additional record: %w", err)
		}
		response.Additional = append(response.Additional, *record)
		offset = newOffset
	}
	
	return response, nil
}

// SerializeQuery serializes a DNS query to raw bytes
func (p *Parser) SerializeQuery(query *DNSQuery) ([]byte, error) {
	var buf bytes.Buffer
	
	// Write header
	binary.Write(&buf, binary.BigEndian, query.ID)
	binary.Write(&buf, binary.BigEndian, uint16(FlagRD)) // Recursion desired
	binary.Write(&buf, binary.BigEndian, uint16(1))      // QDCOUNT
	binary.Write(&buf, binary.BigEndian, uint16(0))      // ANCOUNT
	binary.Write(&buf, binary.BigEndian, uint16(0))      // NSCOUNT
	binary.Write(&buf, binary.BigEndian, uint16(0))      // ARCOUNT
	
	// Write question
	if err := p.writeQuestion(&buf, query.Question); err != nil {
		return nil, fmt.Errorf("failed to write question: %w", err)
	}
	
	return buf.Bytes(), nil
}

// parseQuestion parses a DNS question from data starting at offset 0
func (p *Parser) parseQuestion(data []byte) (*DNSQuestion, error) {
	question, _, err := p.parseQuestionAt(data, 0)
	return question, err
}

// parseQuestionAt parses a DNS question from data starting at given offset
func (p *Parser) parseQuestionAt(data []byte, offset int) (*DNSQuestion, int, error) {
	name, newOffset, err := p.parseName(data, offset)
	if err != nil {
		return nil, 0, err
	}
	
	if newOffset+4 > len(data) {
		return nil, 0, ErrShortMessage
	}
	
	qtype := binary.BigEndian.Uint16(data[newOffset : newOffset+2])
	qclass := binary.BigEndian.Uint16(data[newOffset+2 : newOffset+4])
	
	return &DNSQuestion{
		Name:  name,
		Type:  qtype,
		Class: qclass,
	}, newOffset + 4, nil
}

// parseRecordAt parses a DNS record from data starting at given offset
func (p *Parser) parseRecordAt(data []byte, offset int) (*DNSRecord, int, error) {
	name, newOffset, err := p.parseName(data, offset)
	if err != nil {
		return nil, 0, err
	}
	
	if newOffset+10 > len(data) {
		return nil, 0, ErrShortMessage
	}
	
	rtype := binary.BigEndian.Uint16(data[newOffset : newOffset+2])
	rclass := binary.BigEndian.Uint16(data[newOffset+2 : newOffset+4])
	ttl := binary.BigEndian.Uint32(data[newOffset+4 : newOffset+8])
	rdlength := binary.BigEndian.Uint16(data[newOffset+8 : newOffset+10])
	
	newOffset += 10
	
	if newOffset+int(rdlength) > len(data) {
		return nil, 0, ErrShortMessage
	}
	
	rdata := make([]byte, rdlength)
	copy(rdata, data[newOffset:newOffset+int(rdlength)])
	
	return &DNSRecord{
		Name:  name,
		Type:  rtype,
		Class: rclass,
		TTL:   ttl,
		Data:  rdata,
	}, newOffset + int(rdlength), nil
}

// parseName parses a DNS name with compression support
func (p *Parser) parseName(data []byte, offset int) (string, int, error) {
	var labels []string
	originalOffset := offset
	jumped := false
	jumps := 0
	
	for {
		if offset >= len(data) {
			return "", 0, ErrShortMessage
		}
		
		length := int(data[offset])
		
		// Check for compression pointer
		if length&0xC0 == 0xC0 {
			if offset+1 >= len(data) {
				return "", 0, ErrShortMessage
			}
			
			// Prevent infinite loops
			jumps++
			if jumps > 10 {
				return "", 0, ErrCompressionLoop
			}
			
			if !jumped {
				originalOffset = offset + 2
				jumped = true
			}
			
			// Extract pointer
			pointer := int(binary.BigEndian.Uint16(data[offset:offset+2]) & 0x3FFF)
			if pointer >= len(data) {
				return "", 0, ErrInvalidName
			}
			offset = pointer
			continue
		}
		
		// End of name
		if length == 0 {
			offset++
			break
		}
		
		// Regular label
		if length > 63 {
			return "", 0, ErrInvalidLabel
		}
		
		if offset+1+length > len(data) {
			return "", 0, ErrShortMessage
		}
		
		label := string(data[offset+1 : offset+1+length])
		labels = append(labels, label)
		offset += 1 + length
	}
	
	name := strings.Join(labels, ".")
	if len(name) > 255 {
		return "", 0, ErrNameTooLong
	}
	
	if jumped {
		return name, originalOffset, nil
	}
	return name, offset, nil
}

// writeQuestion writes a DNS question to the buffer
func (p *Parser) writeQuestion(buf *bytes.Buffer, question DNSQuestion) error {
	if err := p.writeName(buf, question.Name); err != nil {
		return err
	}
	binary.Write(buf, binary.BigEndian, question.Type)
	binary.Write(buf, binary.BigEndian, question.Class)
	return nil
}

// writeRecord writes a DNS record to the buffer
func (p *Parser) writeRecord(buf *bytes.Buffer, record DNSRecord) error {
	if err := p.writeName(buf, record.Name); err != nil {
		return err
	}
	binary.Write(buf, binary.BigEndian, record.Type)
	binary.Write(buf, binary.BigEndian, record.Class)
	binary.Write(buf, binary.BigEndian, record.TTL)
	binary.Write(buf, binary.BigEndian, uint16(len(record.Data)))
	buf.Write(record.Data)
	return nil
}

// writeName writes a DNS name to the buffer
func (p *Parser) writeName(buf *bytes.Buffer, name string) error {
	if name == "" || name == "." {
		buf.WriteByte(0)
		return nil
	}
	
	labels := strings.Split(name, ".")
	if labels[len(labels)-1] == "" {
		labels = labels[:len(labels)-1]
	}
	
	for _, label := range labels {
		if len(label) > 63 {
			return ErrInvalidLabel
		}
		buf.WriteByte(byte(len(label)))
		buf.WriteString(label)
	}
	buf.WriteByte(0)
	return nil
}