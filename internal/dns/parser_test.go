package dns

import (
	"encoding/binary"
	"testing"
)

func TestParser_ParseQuery(t *testing.T) {
	parser := NewParser()

	// Create a simple DNS query for "example.com" A record
	queryData := createTestQuery()

	query, err := parser.ParseQuery(queryData)
	if err != nil {
		t.Fatalf("Failed to parse query: %v", err)
	}

	if query.Question.Name != "example.com" {
		t.Errorf("Expected domain 'example.com', got '%s'", query.Question.Name)
	}

	if query.Question.Type != TypeA {
		t.Errorf("Expected type A (%d), got %d", TypeA, query.Question.Type)
	}

	if query.Question.Class != ClassIN {
		t.Errorf("Expected class IN (%d), got %d", ClassIN, query.Question.Class)
	}
}

func TestParser_SerializeQuery(t *testing.T) {
	parser := NewParser()

	query := &DNSQuery{
		ID: 0x1234,
		Question: DNSQuestion{
			Name:  "example.com",
			Type:  TypeA,
			Class: ClassIN,
		},
	}

	data, err := parser.SerializeQuery(query)
	if err != nil {
		t.Fatalf("Failed to serialize query: %v", err)
	}

	// Parse it back
	parsedQuery, err := parser.ParseQuery(data)
	if err != nil {
		t.Fatalf("Failed to parse serialized query: %v", err)
	}

	if parsedQuery.ID != query.ID {
		t.Errorf("Expected ID %d, got %d", query.ID, parsedQuery.ID)
	}

	if parsedQuery.Question.Name != query.Question.Name {
		t.Errorf("Expected name '%s', got '%s'", query.Question.Name, parsedQuery.Question.Name)
	}

	if parsedQuery.Question.Type != query.Question.Type {
		t.Errorf("Expected type %d, got %d", query.Question.Type, parsedQuery.Question.Type)
	}
}

func TestParser_SerializeResponse(t *testing.T) {
	parser := NewParser()

	response := &DNSResponse{
		ID: 0x1234,
		Question: DNSQuestion{
			Name:  "example.com",
			Type:  TypeA,
			Class: ClassIN,
		},
		Answers: []DNSRecord{
			{
				Name:  "example.com",
				Type:  TypeA,
				Class: ClassIN,
				TTL:   300,
				Data:  []byte{192, 168, 1, 1}, // 192.168.1.1
			},
		},
		ResponseCode: RCodeNoError,
	}

	data, err := parser.SerializeResponse(response)
	if err != nil {
		t.Fatalf("Failed to serialize response: %v", err)
	}

	// Parse it back
	parsedResponse, err := parser.ParseResponse(data)
	if err != nil {
		t.Fatalf("Failed to parse serialized response: %v", err)
	}

	if parsedResponse.ID != response.ID {
		t.Errorf("Expected ID %d, got %d", response.ID, parsedResponse.ID)
	}

	if parsedResponse.Question.Name != response.Question.Name {
		t.Errorf("Expected name '%s', got '%s'", response.Question.Name, parsedResponse.Question.Name)
	}

	if len(parsedResponse.Answers) != len(response.Answers) {
		t.Errorf("Expected %d answers, got %d", len(response.Answers), len(parsedResponse.Answers))
	}

	if len(parsedResponse.Answers) > 0 {
		answer := parsedResponse.Answers[0]
		expectedAnswer := response.Answers[0]

		if answer.Name != expectedAnswer.Name {
			t.Errorf("Expected answer name '%s', got '%s'", expectedAnswer.Name, answer.Name)
		}

		if answer.Type != expectedAnswer.Type {
			t.Errorf("Expected answer type %d, got %d", expectedAnswer.Type, answer.Type)
		}

		if answer.TTL != expectedAnswer.TTL {
			t.Errorf("Expected answer TTL %d, got %d", expectedAnswer.TTL, answer.TTL)
		}

		if len(answer.Data) != len(expectedAnswer.Data) {
			t.Errorf("Expected answer data length %d, got %d", len(expectedAnswer.Data), len(answer.Data))
		}
	}
}

func TestParser_ParseEmptyQuery(t *testing.T) {
	parser := NewParser()

	// Test with empty data
	_, err := parser.ParseQuery([]byte{})
	if err == nil {
		t.Error("Expected error for empty query data")
	}

	// Test with short data
	_, err = parser.ParseQuery([]byte{1, 2, 3})
	if err == nil {
		t.Error("Expected error for short query data")
	}
}

func TestParser_ParseInvalidQuery(t *testing.T) {
	parser := NewParser()

	// Create invalid query (response flag set)
	invalidData := make([]byte, 12)
	binary.BigEndian.PutUint16(invalidData[0:2], 0x1234) // ID
	binary.BigEndian.PutUint16(invalidData[2:4], FlagQR) // QR flag set (response)
	binary.BigEndian.PutUint16(invalidData[4:6], 1)      // QDCOUNT

	_, err := parser.ParseQuery(invalidData)
	if err == nil {
		t.Error("Expected error for invalid query (response flag set)")
	}
}

func TestParser_NameEncoding(t *testing.T) {
	parser := NewParser()

	testCases := []struct {
		name     string
		expected string
	}{
		{"example.com", "example.com"},
		{"sub.example.com", "sub.example.com"},
		{"a.b.c.d.example.com", "a.b.c.d.example.com"},
		{"test", "test"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			query := &DNSQuery{
				ID: 0x1234,
				Question: DNSQuestion{
					Name:  tc.name,
					Type:  TypeA,
					Class: ClassIN,
				},
			}

			data, err := parser.SerializeQuery(query)
			if err != nil {
				t.Fatalf("Failed to serialize query for %s: %v", tc.name, err)
			}

			parsedQuery, err := parser.ParseQuery(data)
			if err != nil {
				t.Fatalf("Failed to parse query for %s: %v", tc.name, err)
			}

			if parsedQuery.Question.Name != tc.expected {
				t.Errorf("Expected name '%s', got '%s'", tc.expected, parsedQuery.Question.Name)
			}
		})
	}
}

// Helper function to create a test DNS query
func createTestQuery() []byte {
	// Create a DNS query for "example.com" A record
	data := make([]byte, 0, 512)

	// Header
	data = append(data, 0x12, 0x34) // ID
	data = append(data, 0x01, 0x00) // Flags (RD=1)
	data = append(data, 0x00, 0x01) // QDCOUNT
	data = append(data, 0x00, 0x00) // ANCOUNT
	data = append(data, 0x00, 0x00) // NSCOUNT
	data = append(data, 0x00, 0x00) // ARCOUNT

	// Question: example.com
	data = append(data, 7) // Length of "example"
	data = append(data, []byte("example")...)
	data = append(data, 3) // Length of "com"
	data = append(data, []byte("com")...)
	data = append(data, 0) // End of name

	// QTYPE and QCLASS
	data = append(data, 0x00, 0x01) // A record
	data = append(data, 0x00, 0x01) // IN class

	return data
}
