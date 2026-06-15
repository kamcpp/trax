package trax

import (
	"encoding/json"
	"net/http/httptest"
	"net/url"
	"strconv"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
)

func TestSessionParamsBuilder(t *testing.T) {
	t.Run("Basic session params building", func(t *testing.T) {
		session := NewSessionParamsBuilder().
			SessionId("session123").
			AuthProvider("oauth").
			TokenType("bearer").
			Token("token456").
			Identity("user789").
			Build()

		if session.SessionId != "session123" {
			t.Errorf("Expected SessionId 'session123', got '%s'", session.SessionId)
		}
		if session.AuthProvider != "oauth" {
			t.Errorf("Expected AuthProvider 'oauth', got '%s'", session.AuthProvider)
		}
		if session.TokenType != "bearer" {
			t.Errorf("Expected TokenType 'bearer', got '%s'", session.TokenType)
		}
		if session.Token != "token456" {
			t.Errorf("Expected Token 'token456', got '%s'", session.Token)
		}
		if session.Identity != "user789" {
			t.Errorf("Expected Identity 'user789', got '%s'", session.Identity)
		}
	})

	t.Run("Anonymous session", func(t *testing.T) {
		session := NewSessionParamsBuilder().Anonymous().Build()

		expectedFields := []string{session.SessionId, session.AuthProvider, session.TokenType, session.Token, session.Identity}
		for i, field := range expectedFields {
			if field != "none" {
				t.Errorf("Expected anonymous field %d to be 'none', got '%s'", i, field)
			}
		}
	})
}

func TestPayloadBuilder(t *testing.T) {
	t.Run("Basic payload building", func(t *testing.T) {
		payload := NewPayloadBuilder().
			Metadata("test metadata").
			Type("custom/type").
			ContentType("application/json").
			Encoding("utf-16").
			Data("test data").
			Build()

		if payload.Metadata != "test metadata" {
			t.Errorf("Expected Metadata 'test metadata', got '%s'", payload.Metadata)
		}
		if payload.Type != "custom/type" {
			t.Errorf("Expected Type 'custom/type', got '%s'", payload.Type)
		}
		if payload.ContentType != "application/json" {
			t.Errorf("Expected ContentType 'application/json', got '%s'", payload.ContentType)
		}
		if payload.Encoding != "utf-16" {
			t.Errorf("Expected Encoding 'utf-16', got '%s'", payload.Encoding)
		}
		if payload.Data != "test data" {
			t.Errorf("Expected Data 'test data', got '%s'", payload.Data)
		}
	})

	t.Run("JSON payload", func(t *testing.T) {
		testData := `{"key": "value"}`
		payload := NewPayloadBuilder().Json(testData).Build()

		if payload.ContentType != "application/json" {
			t.Errorf("Expected ContentType 'application/json', got '%s'", payload.ContentType)
		}
		if payload.Encoding != "utf-8" {
			t.Errorf("Expected Encoding 'utf-8', got '%s'", payload.Encoding)
		}
		if payload.Data != testData {
			t.Errorf("Expected Data '%s', got '%s'", testData, payload.Data)
		}
	})

	t.Run("XML payload", func(t *testing.T) {
		testData := "<root><item>value</item></root>"
		payload := NewPayloadBuilder().Xml(testData).Build()

		if payload.ContentType != "application/xml" {
			t.Errorf("Expected ContentType 'application/xml', got '%s'", payload.ContentType)
		}
		if payload.Encoding != "utf-8" {
			t.Errorf("Expected Encoding 'utf-8', got '%s'", payload.Encoding)
		}
		if payload.Data != testData {
			t.Errorf("Expected Data '%s', got '%s'", testData, payload.Data)
		}
	})

	t.Run("Plain text payload", func(t *testing.T) {
		testData := "Hello, World!"
		payload := NewPayloadBuilder().PlainText(testData).Build()

		if payload.ContentType != "text/plain" {
			t.Errorf("Expected ContentType 'text/plain', got '%s'", payload.ContentType)
		}
		if payload.Encoding != "utf-8" {
			t.Errorf("Expected Encoding 'utf-8', got '%s'", payload.Encoding)
		}
		if payload.Data != testData {
			t.Errorf("Expected Data '%s', got '%s'", testData, payload.Data)
		}
	})
}

func TestPayloadFactoryMethods(t *testing.T) {
	t.Run("NewJsonPayload without metadata", func(t *testing.T) {
		testData := `{"test": true}`
		payload := NewJsonPayload(testData)

		if payload.ContentType != "application/json" {
			t.Errorf("Expected ContentType 'application/json', got '%s'", payload.ContentType)
		}
		if payload.Data != testData {
			t.Errorf("Expected Data '%s', got '%s'", testData, payload.Data)
		}
		if payload.Metadata != "" {
			t.Errorf("Expected empty Metadata, got '%s'", payload.Metadata)
		}
	})

	t.Run("NewJsonPayload with metadata", func(t *testing.T) {
		testData := `{"test": true}`
		metadata := "json metadata"
		payload := NewJsonPayload(testData, metadata)

		if payload.Metadata != metadata {
			t.Errorf("Expected Metadata '%s', got '%s'", metadata, payload.Metadata)
		}
	})

	t.Run("NewXmlPayload with metadata", func(t *testing.T) {
		testData := "<test>true</test>"
		metadata := "xml metadata"
		payload := NewXmlPayload(testData, metadata)

		if payload.ContentType != "application/xml" {
			t.Errorf("Expected ContentType 'application/xml', got '%s'", payload.ContentType)
		}
		if payload.Metadata != metadata {
			t.Errorf("Expected Metadata '%s', got '%s'", metadata, payload.Metadata)
		}
	})

	t.Run("NewPlainTextPayload with metadata", func(t *testing.T) {
		testData := "plain text"
		metadata := "text metadata"
		payload := NewPlainTextPayload(testData, metadata)

		if payload.ContentType != "text/plain" {
			t.Errorf("Expected ContentType 'text/plain', got '%s'", payload.ContentType)
		}
		if payload.Metadata != metadata {
			t.Errorf("Expected Metadata '%s', got '%s'", metadata, payload.Metadata)
		}
	})
}

func TestTraxMessageBuilder(t *testing.T) {
	t.Run("Auto-generated fields", func(t *testing.T) {
		msg := NewTraxMessageBuilder().Build()

		if len(msg.MessageId) == 0 {
			t.Error("Expected auto-generated MessageId")
		}
		if len(msg.TraceId) == 0 {
			t.Error("Expected auto-generated TraceId")
		}
		if len(msg.Timestamp) == 0 {
			t.Error("Expected auto-generated Timestamp")
		}
		if msg.Metadata == nil {
			t.Error("Expected initialized Metadata map")
		}
		if msg.Tags == nil {
			t.Error("Expected initialized Tags slice")
		}
		if msg.Payloads == nil {
			t.Error("Expected initialized Payloads slice")
		}
	})

	t.Run("Manual field setting", func(t *testing.T) {
		session := NewSessionParamsBuilder().SessionId("test").Build()
		msg := NewTraxMessageBuilder().
			MessageId("msg123").
			RefMessageId("ref456").
			ExecutionId("exec789").
			RefExecutionId("refexec000").
			TraceId("trace111").
			ClusterId("cluster456").
			Timestamp("1234567890").
			Origin("test-origin").
			Issuer("test-issuer").
			Referrer("test-referrer").
			Session(session).
			Build()

		if msg.MessageId != "msg123" {
			t.Errorf("Expected MessageId 'msg123', got '%s'", msg.MessageId)
		}
		if msg.RefMessageId != "ref456" {
			t.Errorf("Expected RefMessageId 'ref456', got '%s'", msg.RefMessageId)
		}
		if msg.ExecutionId != "exec789" {
			t.Errorf("Expected ExecutionId 'exec789', got '%s'", msg.ExecutionId)
		}
		if msg.RefExecutionId != "refexec000" {
			t.Errorf("Expected RefExecutionId 'refexec000', got '%s'", msg.RefExecutionId)
		}
		if msg.TraceId != "trace111" {
			t.Errorf("Expected TraceId 'trace111', got '%s'", msg.TraceId)
		}
		if msg.ClusterId != "cluster456" {
			t.Errorf("Expected ClusterId 'cluster456', got '%s'", msg.ClusterId)
		}
		if msg.Timestamp != "1234567890" {
			t.Errorf("Expected Timestamp '1234567890', got '%s'", msg.Timestamp)
		}
		if msg.Origin != "test-origin" {
			t.Errorf("Expected Origin 'test-origin', got '%s'", msg.Origin)
		}
		if msg.Issuer != "test-issuer" {
			t.Errorf("Expected Issuer 'test-issuer', got '%s'", msg.Issuer)
		}
		if msg.Referrer != "test-referrer" {
			t.Errorf("Expected Referrer 'test-referrer', got '%s'", msg.Referrer)
		}
		if msg.Session.SessionId != "test" {
			t.Errorf("Expected Session.SessionId 'test', got '%s'", msg.Session.SessionId)
		}
	})

	t.Run("Metadata handling", func(t *testing.T) {
		metadata := map[string]string{"key1": "value1", "key2": "value2"}
		msg := NewTraxMessageBuilder().
			Metadata(metadata).
			AddMetadata("key3", "value3").
			Build()

		if len(msg.Metadata) != 3 {
			t.Errorf("Expected 3 metadata entries, got %d", len(msg.Metadata))
		}
		if msg.Metadata["key1"] != "value1" {
			t.Errorf("Expected key1='value1', got '%s'", msg.Metadata["key1"])
		}
		if msg.Metadata["key3"] != "value3" {
			t.Errorf("Expected key3='value3', got '%s'", msg.Metadata["key3"])
		}
	})

	t.Run("Tags handling", func(t *testing.T) {
		tags := []string{"tag1", "tag2"}
		msg := NewTraxMessageBuilder().
			Tags(tags).
			AddTag("tag3").
			Build()

		if len(msg.Tags) != 3 {
			t.Errorf("Expected 3 tags, got %d", len(msg.Tags))
		}
		expectedTags := []string{"tag1", "tag2", "tag3"}
		for i, tag := range expectedTags {
			if msg.Tags[i] != tag {
				t.Errorf("Expected tag[%d]='%s', got '%s'", i, tag, msg.Tags[i])
			}
		}
	})
}

func TestTraxMessageBuilderPayloads(t *testing.T) {
	t.Run("Single payload methods", func(t *testing.T) {
		payload := NewJsonPayload(`{"test": true}`)
		msg := NewTraxMessageBuilder().Payload(payload).Build()

		if len(msg.Payloads) != 1 {
			t.Errorf("Expected 1 payload, got %d", len(msg.Payloads))
		}
		if msg.Payloads[0].ContentType != "application/json" {
			t.Errorf("Expected payload ContentType 'application/json', got '%s'", msg.Payloads[0].ContentType)
		}
	})

	t.Run("Multiple payload methods", func(t *testing.T) {
		payload1 := NewJsonPayload(`{"test": 1}`)
		payload2 := NewXmlPayload("<test>2</test>")

		msg := NewTraxMessageBuilder().
			AddPayload(payload1).
			AddPayload(payload2).
			Build()

		if len(msg.Payloads) != 2 {
			t.Errorf("Expected 2 payloads, got %d", len(msg.Payloads))
		}
		if msg.Payloads[0].ContentType != "application/json" {
			t.Errorf("Expected first payload ContentType 'application/json', got '%s'", msg.Payloads[0].ContentType)
		}
		if msg.Payloads[1].ContentType != "application/xml" {
			t.Errorf("Expected second payload ContentType 'application/xml', got '%s'", msg.Payloads[1].ContentType)
		}
	})

	t.Run("Convenience payload methods", func(t *testing.T) {
		msg := NewTraxMessageBuilder().
			AddJsonPayload(`{"step": 1}`).
			AddXmlPayload("<step>2</step>").
			AddPlainTextPayload("Step 3").
			Build()

		if len(msg.Payloads) != 3 {
			t.Errorf("Expected 3 payloads, got %d", len(msg.Payloads))
		}

		expectedTypes := []string{"application/json", "application/xml", "text/plain"}
		for i, expectedType := range expectedTypes {
			if msg.Payloads[i].ContentType != expectedType {
				t.Errorf("Expected payload[%d] ContentType '%s', got '%s'", i, expectedType, msg.Payloads[i].ContentType)
			}
		}
	})

	t.Run("Payload replacement", func(t *testing.T) {
		builder := NewTraxMessageBuilder()
		builder.AddJsonPayload(`{"initial": true}`)
		builder.AddXmlPayload("<initial>true</initial>")

		// This should replace all existing payloads
		builder.JsonPayload(`{"replaced": true}`)
		msg := builder.Build()

		if len(msg.Payloads) != 1 {
			t.Errorf("Expected 1 payload after replacement, got %d", len(msg.Payloads))
		}
		if msg.Payloads[0].Data != `{"replaced": true}` {
			t.Errorf("Expected replaced payload data, got '%s'", msg.Payloads[0].Data)
		}
	})

	t.Run("Payload management methods", func(t *testing.T) {
		builder := NewTraxMessageBuilder()

		if builder.PayloadCount() != 0 {
			t.Errorf("Expected 0 payloads initially, got %d", builder.PayloadCount())
		}
		if builder.HasPayloads() {
			t.Error("Expected no payloads initially")
		}

		builder.AddJsonPayload(`{"test": true}`)

		if builder.PayloadCount() != 1 {
			t.Errorf("Expected 1 payload after adding, got %d", builder.PayloadCount())
		}
		if !builder.HasPayloads() {
			t.Error("Expected to have payloads after adding")
		}

		builder.ClearPayloads()

		if builder.PayloadCount() != 0 {
			t.Errorf("Expected 0 payloads after clearing, got %d", builder.PayloadCount())
		}
		if builder.HasPayloads() {
			t.Error("Expected no payloads after clearing")
		}
	})

	t.Run("PayloadWithMetadata method", func(t *testing.T) {
		msg := NewTraxMessageBuilder().
			PayloadWithMetadata("custom/type", "utf-16", "test data", "test metadata").
			Build()

		if len(msg.Payloads) != 1 {
			t.Errorf("Expected 1 payload, got %d", len(msg.Payloads))
		}

		payload := msg.Payloads[0]
		if payload.Type != "custom/type" {
			t.Errorf("Expected type 'custom/type', got '%s'", payload.Type)
		}
		if payload.Encoding != "utf-16" {
			t.Errorf("Expected encoding 'utf-16', got '%s'", payload.Encoding)
		}
		if payload.Data != "test data" {
			t.Errorf("Expected data 'test data', got '%s'", payload.Data)
		}
		if payload.Metadata != "test metadata" {
			t.Errorf("Expected metadata 'test metadata', got '%s'", payload.Metadata)
		}
	})
}

func TestMessageFactoryMethods(t *testing.T) {
	t.Run("NewAnonymousMessage", func(t *testing.T) {
		msg := NewAnonymousMessage().Build()

		if msg.Session.SessionId != "none" {
			t.Errorf("Expected anonymous session, got SessionId '%s'", msg.Session.SessionId)
		}
	})

	t.Run("NewAuthenticatedMessage", func(t *testing.T) {
		session := NewSessionParamsBuilder().SessionId("auth123").Build()
		msg := NewAuthenticatedMessage(session).Build()

		if msg.Session.SessionId != "auth123" {
			t.Errorf("Expected SessionId 'auth123', got '%s'", msg.Session.SessionId)
		}
	})

	t.Run("NewJsonMessage", func(t *testing.T) {
		testData := `{"factory": true}`
		msg := NewJsonMessage(testData).Build()

		if len(msg.Payloads) != 1 {
			t.Errorf("Expected 1 payload, got %d", len(msg.Payloads))
		}
		if msg.Payloads[0].ContentType != "application/json" {
			t.Errorf("Expected JSON payload ContentType, got '%s'", msg.Payloads[0].ContentType)
		}
		if msg.Payloads[0].Data != testData {
			t.Errorf("Expected data '%s', got '%s'", testData, msg.Payloads[0].Data)
		}
	})

	t.Run("NewAnonymousJsonMessage", func(t *testing.T) {
		testData := `{"anonymous": true}`
		msg := NewAnonymousJsonMessage(testData).Build()

		if msg.Session.SessionId != "none" {
			t.Errorf("Expected anonymous session, got SessionId '%s'", msg.Session.SessionId)
		}
		if len(msg.Payloads) != 1 {
			t.Errorf("Expected 1 payload, got %d", len(msg.Payloads))
		}
		if msg.Payloads[0].Data != testData {
			t.Errorf("Expected data '%s', got '%s'", testData, msg.Payloads[0].Data)
		}
	})

	t.Run("NewReplyMessage", func(t *testing.T) {
		original := NewTraxMessageBuilder().
			MessageId("original123").
			TraceId("trace456").
			ExecutionId("exec789").
			Build()

		reply := NewReplyMessage(original).Build()

		if reply.RefMessageId != "original123" {
			t.Errorf("Expected RefMessageId 'original123', got '%s'", reply.RefMessageId)
		}
		if reply.TraceId != "trace456" {
			t.Errorf("Expected TraceId 'trace456', got '%s'", reply.TraceId)
		}
		if reply.RefExecutionId != "exec789" {
			t.Errorf("Expected RefExecutionId 'exec789', got '%s'", reply.RefExecutionId)
		}
	})

	t.Run("NewMultiPayloadMessage", func(t *testing.T) {
		payload1 := NewJsonPayload(`{"multi": 1}`)
		payload2 := NewXmlPayload("<multi>2</multi>")
		payload3 := NewPlainTextPayload("multi 3")

		msg := NewMultiPayloadMessage(payload1, payload2, payload3).Build()

		if len(msg.Payloads) != 3 {
			t.Errorf("Expected 3 payloads, got %d", len(msg.Payloads))
		}

		expectedTypes := []string{"application/json", "application/xml", "text/plain"}
		for i, expectedType := range expectedTypes {
			if msg.Payloads[i].ContentType != expectedType {
				t.Errorf("Expected payload[%d] ContentType '%s', got '%s'", i, expectedType, msg.Payloads[i].ContentType)
			}
		}
	})

	t.Run("NewAnonymousMultiPayloadMessage", func(t *testing.T) {
		payload1 := NewJsonPayload(`{"anon": true}`)
		payload2 := NewPlainTextPayload("anonymous multi")

		msg := NewAnonymousMultiPayloadMessage(payload1, payload2).Build()

		if msg.Session.SessionId != "none" {
			t.Errorf("Expected anonymous session, got SessionId '%s'", msg.Session.SessionId)
		}
		if len(msg.Payloads) != 2 {
			t.Errorf("Expected 2 payloads, got %d", len(msg.Payloads))
		}
	})
}

func TestMessageSerialization(t *testing.T) {
	t.Run("JSON serialization and deserialization", func(t *testing.T) {
		original := NewTraxMessageBuilder().
			MessageId("test123").
			AnonymousSession().
			AddJsonPayload(`{"serialization": "test"}`).
			AddMetadata("test", "value").
			AddTag("serialization").
			Build()

		// Serialize to JSON
		jsonBytes, err := json.Marshal(original)
		if err != nil {
			t.Fatalf("Failed to marshal message: %v", err)
		}

		// Deserialize from JSON
		var deserialized TraxMessage
		err = json.Unmarshal(jsonBytes, &deserialized)
		if err != nil {
			t.Fatalf("Failed to unmarshal message: %v", err)
		}

		// Verify key fields
		if deserialized.MessageId != original.MessageId {
			t.Errorf("MessageId mismatch after serialization: expected '%s', got '%s'", original.MessageId, deserialized.MessageId)
		}
		if len(deserialized.Payloads) != len(original.Payloads) {
			t.Errorf("Payloads length mismatch: expected %d, got %d", len(original.Payloads), len(deserialized.Payloads))
		}
		if len(deserialized.Metadata) != len(original.Metadata) {
			t.Errorf("Metadata length mismatch: expected %d, got %d", len(original.Metadata), len(deserialized.Metadata))
		}
		if len(deserialized.Tags) != len(original.Tags) {
			t.Errorf("Tags length mismatch: expected %d, got %d", len(original.Tags), len(deserialized.Tags))
		}
	})
}

func TestTimestampHandling(t *testing.T) {
	t.Run("TimestampNow sets current time", func(t *testing.T) {
		beforeTime := time.Now().UnixMilli()
		msg := NewTraxMessageBuilder().TimestampNow().Build()
		afterTime := time.Now().UnixMilli()

		// Parse timestamp from message (should be unix milliseconds)
		msgTime, err := strconv.ParseInt(msg.Timestamp, 10, 64)
		if err != nil {
			t.Fatalf("Could not parse timestamp '%s' as unix millis: %v", msg.Timestamp, err)
		}

		// Verify timestamp is within reasonable bounds (allowing for test execution time)
		if msgTime < beforeTime-1000 || msgTime > afterTime+1000 {
			t.Errorf("Timestamp %d not within expected range [%d, %d]", msgTime, beforeTime, afterTime)
		}
	})
}

func TestTraceIdFromGinContext(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("TraceId from header", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		req := httptest.NewRequest("GET", "/test", nil)
		req.Header.Set("x-trace-id", "trace123")
		c.Request = req

		msg := NewTraxMessageBuilder().
			TraceIdFromGinContext(c).
			Build()

		if msg.TraceId != "trace123" {
			t.Errorf("Expected TraceId 'trace123', got '%s'", msg.TraceId)
		}
	})

	t.Run("TraceId from query param", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		req := httptest.NewRequest("GET", "/test?trace-id=trace456", nil)
		c.Request = req

		// Parse query params
		c.Request.URL.RawQuery = "trace-id=trace456"
		values, _ := url.ParseQuery(c.Request.URL.RawQuery)
		c.Request.Form = values

		msg := NewTraxMessageBuilder().
			TraceIdFromGinContext(c).
			Build()

		if msg.TraceId != "trace456" {
			t.Errorf("Expected TraceId 'trace456', got '%s'", msg.TraceId)
		}
	})

	t.Run("Header takes precedence over query param", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		req := httptest.NewRequest("GET", "/test?trace-id=trace-query", nil)
		req.Header.Set("x-trace-id", "trace-header")
		c.Request = req

		// Parse query params
		c.Request.URL.RawQuery = "trace-id=trace-query"
		values, _ := url.ParseQuery(c.Request.URL.RawQuery)
		c.Request.Form = values

		msg := NewTraxMessageBuilder().
			TraceIdFromGinContext(c).
			Build()

		if msg.TraceId != "trace-header" {
			t.Errorf("Expected header TraceId 'trace-header', got '%s'", msg.TraceId)
		}
	})

	t.Run("No trace ID found keeps auto-generated", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		req := httptest.NewRequest("GET", "/test", nil)
		c.Request = req

		builder := NewTraxMessageBuilder()
		originalTraceId := builder.message.TraceId
		msg := builder.TraceIdFromGinContext(c).Build()

		if msg.TraceId != originalTraceId {
			t.Errorf("Expected original auto-generated TraceId to be preserved")
		}
		if len(msg.TraceId) == 0 {
			t.Error("Expected auto-generated TraceId to be present")
		}
	})
}

func TestJsonMethods(t *testing.T) {
	t.Run("SessionParams Json method", func(t *testing.T) {
		session := NewSessionParamsBuilder().
			SessionId("test123").
			AuthProvider("oauth").
			TokenType("bearer").
			Token("token456").
			Identity("user789").
			Build()

		jsonStr := session.Json()

		// Parse back to verify it's valid JSON
		var parsed SessionParams
		err := json.Unmarshal([]byte(jsonStr), &parsed)
		if err != nil {
			t.Errorf("Failed to parse JSON back to SessionParams: %v", err)
		}

		if parsed.SessionId != "test123" {
			t.Errorf("Expected SessionId 'test123', got '%s'", parsed.SessionId)
		}
		if parsed.AuthProvider != "oauth" {
			t.Errorf("Expected AuthProvider 'oauth', got '%s'", parsed.AuthProvider)
		}
	})

	t.Run("Payload Json method", func(t *testing.T) {
		payload := NewPayloadBuilder().
			Metadata("test metadata").
			Type("application/json").
			ContentType("application/json").
			Encoding("utf-8").
			Data(`{"test": "data"}`).
			Build()

		jsonStr := payload.Json()

		// Parse back to verify it's valid JSON
		var parsed Payload
		err := json.Unmarshal([]byte(jsonStr), &parsed)
		if err != nil {
			t.Errorf("Failed to parse JSON back to Payload: %v", err)
		}

		if parsed.Metadata != "test metadata" {
			t.Errorf("Expected Metadata 'test metadata', got '%s'", parsed.Metadata)
		}
		if parsed.Type != "application/json" {
			t.Errorf("Expected Type 'application/json', got '%s'", parsed.Type)
		}
		if parsed.ContentType != "application/json" {
			t.Errorf("Expected ContentType 'application/json', got '%s'", parsed.ContentType)
		}
	})

	t.Run("TraxMessage Json method", func(t *testing.T) {
		msg := NewTraxMessageBuilder().
			MessageId("msg123").
			Origin("test-origin").
			Issuer("test-issuer").
			AnonymousSession().
			AddJsonPayload(`{"test": "data"}`).
			Build()

		jsonStr := msg.Json()

		// Parse back to verify it's valid JSON
		var parsed TraxMessage
		err := json.Unmarshal([]byte(jsonStr), &parsed)
		if err != nil {
			t.Errorf("Failed to parse JSON back to TraxMessage: %v", err)
		}

		if parsed.MessageId != "msg123" {
			t.Errorf("Expected MessageId 'msg123', got '%s'", parsed.MessageId)
		}
		if parsed.Origin != "test-origin" {
			t.Errorf("Expected Origin 'test-origin', got '%s'", parsed.Origin)
		}
		if parsed.Issuer != "test-issuer" {
			t.Errorf("Expected Issuer 'test-issuer', got '%s'", parsed.Issuer)
		}
		if len(parsed.Payloads) != 1 {
			t.Errorf("Expected 1 payload, got %d", len(parsed.Payloads))
		}
	})
}

func TestFollowSagaSubmitterPayloadBuilder(t *testing.T) {
	t.Run("Basic FollowSagaSubmitterPayload building", func(t *testing.T) {
		payload := NewFollowSagaSubmitterPayloadBuilder().
			SagaSubmitterId("submitter123").
			Build()

		if payload.SagaSubmitterId != "submitter123" {
			t.Errorf("Expected SagaSubmitterId 'submitter123', got '%s'", payload.SagaSubmitterId)
		}
	})

	t.Run("FollowSagaSubmitterPayload Json method", func(t *testing.T) {
		payload := NewFollowSagaSubmitterPayloadBuilder().
			SagaSubmitterId("submitter456").
			Build()

		jsonStr := payload.Json()

		// Parse back to verify it's valid JSON
		var parsed FollowSagaSubmitterPayload
		err := json.Unmarshal([]byte(jsonStr), &parsed)
		if err != nil {
			t.Errorf("Failed to parse JSON back to FollowSagaSubmitterPayload: %v", err)
		}

		if parsed.SagaSubmitterId != "submitter456" {
			t.Errorf("Expected SagaSubmitterId 'submitter456', got '%s'", parsed.SagaSubmitterId)
		}
	})

	t.Run("Builder and Build flow", func(t *testing.T) {
		builder := NewFollowSagaSubmitterPayloadBuilder().
			SagaSubmitterId("submitter789")

		payload := builder.Build()
		jsonStr := payload.Json()

		// Parse back to verify it's valid JSON
		var parsed FollowSagaSubmitterPayload
		err := json.Unmarshal([]byte(jsonStr), &parsed)
		if err != nil {
			t.Errorf("Failed to parse JSON back to FollowSagaSubmitterPayload: %v", err)
		}

		if parsed.SagaSubmitterId != "submitter789" {
			t.Errorf("Expected SagaSubmitterId 'submitter789', got '%s'", parsed.SagaSubmitterId)
		}
	})

	t.Run("Empty FollowSagaSubmitterPayload", func(t *testing.T) {
		payload := NewFollowSagaSubmitterPayloadBuilder().Build()

		if payload.SagaSubmitterId != "" {
			t.Errorf("Expected empty SagaSubmitterId, got '%s'", payload.SagaSubmitterId)
		}

		// Should still serialize to valid JSON
		jsonStr := payload.Json()
		var parsed FollowSagaSubmitterPayload
		err := json.Unmarshal([]byte(jsonStr), &parsed)
		if err != nil {
			t.Errorf("Failed to parse JSON for empty payload: %v", err)
		}
	})
}
