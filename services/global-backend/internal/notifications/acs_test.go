package notifications

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"
)

type acsTransport func(*http.Request) (*http.Response, error)

func (transport acsTransport) RoundTrip(request *http.Request) (*http.Response, error) {
	return transport(request)
}

func acsResponse(status int, body string) (*http.Response, error) {
	return &http.Response{StatusCode: status, Header: make(http.Header), Body: io.NopCloser(strings.NewReader(body))}, nil
}

func testACSSender(t *testing.T, transport acsTransport) *ACSEmailSender {
	t.Helper()
	sender, err := NewACSEmailSender("https://unit.communication.azure.com", base64.StdEncoding.EncodeToString(make([]byte, 32)), "sender@example.invalid")
	if err != nil {
		t.Fatal(err)
	}
	sender.client.Transport = transport
	return sender
}

func testACSMessage() EmailMessage {
	return EmailMessage{
		OperationID: "f5b85301-3fde-44d9-a2b6-359bab1cab21", FirstSentAt: time.Now().UTC().Truncate(time.Second), FirstAttempt: true,
		Recipient: "recipient@example.invalid", Subject: "Registration confirmed", BodyText: "Confirmed", BodyHTML: "<p>Confirmed</p>",
	}
}

func checkACSPost(t *testing.T, request *http.Request, message EmailMessage) {
	t.Helper()
	if request.Method != http.MethodPost || request.URL.RequestURI() != "/emails:send?api-version=2023-03-31" {
		t.Fatalf("expected first request to submit the email, got %s %s", request.Method, request.URL.RequestURI())
	}
	if request.Header.Get("Operation-Id") != message.OperationID || request.Header.Get("Repeatability-Request-Id") != message.OperationID || request.Header.Get("Repeatability-First-Sent") != message.FirstSentAt.Format(http.TimeFormat) {
		t.Fatal("persisted operation identity or repeatability headers changed")
	}
	if !strings.HasPrefix(request.Header.Get("Authorization"), "HMAC-SHA256 ") || request.Header.Get("x-ms-client-request-id") != message.OperationID {
		t.Fatal("authenticated request identity is missing")
	}
	var payload struct {
		Sender     string `json:"senderAddress"`
		Recipients struct {
			To []struct {
				Address string `json:"address"`
			} `json:"to"`
		} `json:"recipients"`
	}
	if err := json.NewDecoder(request.Body).Decode(&payload); err != nil || payload.Sender != "sender@example.invalid" || len(payload.Recipients.To) != 1 || payload.Recipients.To[0].Address != message.Recipient {
		t.Fatal("send payload lost its sender or recipient")
	}
}

func checkACSGet(t *testing.T, request *http.Request, message EmailMessage) {
	t.Helper()
	if request.Method != http.MethodGet || request.URL.RequestURI() != "/emails/operations/"+message.OperationID+"?api-version=2023-03-31" {
		t.Fatalf("retry must reconcile the same operation: %s %s", request.Method, request.URL.RequestURI())
	}
}

func TestACSFirstSendDoesNotLookupUnknownOperation(t *testing.T) {
	message := testACSMessage()
	calls := 0
	sender := testACSSender(t, func(request *http.Request) (*http.Response, error) {
		calls++
		// A preflight GET is the observed production timeout regression.
		if request.Method == http.MethodGet {
			return nil, context.DeadlineExceeded
		}
		checkACSPost(t, request, message)
		return acsResponse(http.StatusAccepted, `{"status":"Succeeded"}`)
	})
	if err := sender.Send(context.Background(), message); err != nil || calls != 1 {
		t.Fatalf("first send failed or performed an unnecessary lookup: calls=%d err=%v", calls, err)
	}
}

func TestACSAcceptanceIsPolledWithThePersistedIdentity(t *testing.T) {
	message := testACSMessage()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	calls := 0
	sender := testACSSender(t, func(request *http.Request) (*http.Response, error) {
		calls++
		if calls == 1 {
			checkACSPost(t, request, message)
			response, _ := acsResponse(http.StatusAccepted, `{"status":"Running"}`)
			response.Header.Set("Retry-After", "1")
			return response, nil
		}
		checkACSGet(t, request, message)
		return acsResponse(http.StatusOK, `{"status":"Succeeded"}`)
	})
	if err := sender.Send(ctx, message); err != nil || calls != 2 {
		t.Fatalf("provider acceptance was not reconciled: calls=%d err=%v", calls, err)
	}
}

func TestACSStatusTimeoutNeverReposts(t *testing.T) {
	message := testACSMessage()
	message.FirstAttempt = false
	calls := 0
	sender := testACSSender(t, func(request *http.Request) (*http.Response, error) {
		calls++
		checkACSGet(t, request, message)
		<-request.Context().Done()
		return nil, request.Context().Err()
	})
	sender.client.Timeout = 10 * time.Millisecond
	if err := sender.Send(context.Background(), message); err == nil || calls != 1 {
		t.Fatalf("ambiguous status timeout must fail without another POST: calls=%d err=%v", calls, err)
	}
}

func TestACSPostTimeoutReconcilesWithoutDuplicateSend(t *testing.T) {
	message := testACSMessage()
	calls, posts := 0, 0
	sender := testACSSender(t, func(request *http.Request) (*http.Response, error) {
		calls++
		if request.Method == http.MethodPost {
			posts++
			checkACSPost(t, request, message)
			return nil, context.DeadlineExceeded
		}
		checkACSGet(t, request, message)
		return acsResponse(http.StatusOK, `{"status":"Succeeded"}`)
	})
	if err := sender.Send(context.Background(), message); err == nil {
		t.Fatal("ambiguous POST was reported as success")
	}
	message.FirstAttempt = false
	if err := sender.Send(context.Background(), message); err != nil || calls != 2 || posts != 1 {
		t.Fatalf("retry did not reconcile the original POST: calls=%d posts=%d err=%v", calls, posts, err)
	}
}

func TestACSMissingOperationRetryPreservesRepeatability(t *testing.T) {
	message := testACSMessage()
	message.FirstAttempt = false
	calls := 0
	sender := testACSSender(t, func(request *http.Request) (*http.Response, error) {
		calls++
		if calls == 1 {
			checkACSGet(t, request, message)
			return acsResponse(http.StatusNotFound, "")
		}
		checkACSPost(t, request, message)
		return acsResponse(http.StatusAccepted, `{"status":"Succeeded"}`)
	})
	if err := sender.Send(context.Background(), message); err != nil || calls != 2 {
		t.Fatalf("safe retry did not preserve the provider identity: calls=%d err=%v", calls, err)
	}
}

func TestACSAcceptedSendStatusTimeoutReconcilesOnRetry(t *testing.T) {
	message := testACSMessage()
	calls, posts := 0, 0
	sender := testACSSender(t, func(request *http.Request) (*http.Response, error) {
		calls++
		if calls == 1 {
			posts++
			checkACSPost(t, request, message)
			response, _ := acsResponse(http.StatusAccepted, `{"status":"Running"}`)
			response.Header.Set("Retry-After", "1")
			return response, nil
		}
		checkACSGet(t, request, message)
		if calls == 2 {
			return nil, context.DeadlineExceeded
		}
		return acsResponse(http.StatusOK, `{"status":"Succeeded"}`)
	})
	if err := sender.Send(context.Background(), message); err == nil {
		t.Fatal("accepted send with an unknown final outcome was marked successful")
	}
	message.FirstAttempt = false
	if err := sender.Send(context.Background(), message); err != nil || calls != 3 || posts != 1 {
		t.Fatalf("accepted send was not reconciled without another POST: calls=%d posts=%d err=%v", calls, posts, err)
	}
}

func TestACSExpiredOperationCanStillReconcileSuccess(t *testing.T) {
	message := testACSMessage()
	message.FirstAttempt = false
	message.FirstSentAt = time.Now().Add(-6 * time.Minute)
	calls := 0
	sender := testACSSender(t, func(request *http.Request) (*http.Response, error) {
		calls++
		checkACSGet(t, request, message)
		return acsResponse(http.StatusOK, `{"status":"Succeeded"}`)
	})
	if err := sender.Send(context.Background(), message); err != nil || calls != 1 {
		t.Fatalf("expired replay window blocked confirmed acceptance: calls=%d err=%v", calls, err)
	}
}

func TestACSExpiredWindowNeverCreatesAnotherSend(t *testing.T) {
	for _, first := range []bool{true, false} {
		name := "retry"
		if first {
			name = "first-attempt"
		}
		t.Run(name, func(t *testing.T) {
			message := testACSMessage()
			message.FirstSentAt = time.Now().Add(-5 * time.Minute)
			message.FirstAttempt = first
			calls := 0
			sender := testACSSender(t, func(request *http.Request) (*http.Response, error) {
				calls++
				checkACSGet(t, request, message)
				return acsResponse(http.StatusNotFound, "")
			})
			err := sender.Send(context.Background(), message)
			if err == nil || !strings.Contains(err.Error(), "replay window elapsed") || (first && calls != 0) || (!first && calls != 1) {
				t.Fatalf("expired operation was resubmitted: first=%t calls=%d err=%v", first, calls, err)
			}
		})
	}
}

func TestACSUnconfirmedRetryStatesNeverPost(t *testing.T) {
	for _, test := range []struct {
		name   string
		status int
		body   string
	}{
		{"gateway-timeout", http.StatusGatewayTimeout, ""},
		{"unauthorized", http.StatusUnauthorized, ""},
		{"rate-limit", http.StatusTooManyRequests, ""},
		{"invalid-response", http.StatusOK, "{"},
		{"failed", http.StatusOK, `{"status":"Failed"}`},
		{"unknown", http.StatusOK, `{"status":"Unexpected"}`},
	} {
		t.Run(test.name, func(t *testing.T) {
			message := testACSMessage()
			message.FirstAttempt = false
			calls := 0
			sender := testACSSender(t, func(request *http.Request) (*http.Response, error) {
				calls++
				checkACSGet(t, request, message)
				return acsResponse(test.status, test.body)
			})
			if err := sender.Send(context.Background(), message); err == nil || calls != 1 {
				t.Fatalf("unconfirmed retry caused success or resubmission: calls=%d err=%v", calls, err)
			}
		})
	}
}

func TestACSMissingIdentityDoesNotContactProvider(t *testing.T) {
	message := testACSMessage()
	sender := testACSSender(t, func(request *http.Request) (*http.Response, error) {
		t.Fatal("invalid message contacted provider")
		return nil, errors.New("unexpected request")
	})
	message.OperationID = ""
	if err := sender.Send(context.Background(), message); err == nil {
		t.Fatal("missing operation identity was accepted")
	}
	message = testACSMessage()
	message.FirstSentAt = time.Time{}
	if err := sender.Send(context.Background(), message); err == nil {
		t.Fatal("missing persisted first attempt was accepted")
	}
}

func TestACSSenderCategories(t *testing.T) {
	t.Run("account default does not inherit event sender", func(t *testing.T) {
		sender, err := NewACSEmailSender("https://unit.communication.azure.com", base64.StdEncoding.EncodeToString(make([]byte, 32)), "notify@mail.hackways.in")
		if err != nil {
			t.Fatal(err)
		}
		calls := 0
		sender.client.Transport = acsTransport(func(request *http.Request) (*http.Response, error) {
			calls++
			var payload struct {
				Sender string `json:"senderAddress"`
			}
			if err := json.NewDecoder(request.Body).Decode(&payload); err != nil || payload.Sender != "noreply@mail.hackways.in" {
				t.Fatal("account mail inherited the event sender")
			}
			return acsResponse(202, `{"status":"Succeeded"}`)
		})
		message := testACSMessage()
		message.SenderClass = "account"
		if err = sender.Send(context.Background(), message); err != nil || calls != 1 {
			t.Fatal("default account send did not succeed")
		}
	})
	for _, test := range []struct{ class, address string }{{"account", "noreply@mail.hackways.in"}, {"event", "notify@mail.hackways.in"}, {"promotion", "updates@mail.hackways.in"}} {
		t.Run(test.class, func(t *testing.T) {
			sender, err := NewACSEmailSender("https://unit.communication.azure.com", base64.StdEncoding.EncodeToString(make([]byte, 32)), "notify@mail.hackways.in", "noreply@mail.hackways.in", "updates@mail.hackways.in")
			if err != nil {
				t.Fatal(err)
			}
			calls := 0
			sender.client.Transport = acsTransport(func(request *http.Request) (*http.Response, error) {
				calls++
				var payload struct {
					Sender string `json:"senderAddress"`
				}
				if err := json.NewDecoder(request.Body).Decode(&payload); err != nil || payload.Sender != test.address {
					t.Fatal("wrong category sender")
				}
				return acsResponse(202, `{"status":"Succeeded"}`)
			})
			message := testACSMessage()
			message.SenderClass = test.class
			if err = sender.Send(context.Background(), message); err != nil || calls != 1 {
				t.Fatal("category send did not succeed")
			}
		})
	}
	sender := testACSSender(t, func(*http.Request) (*http.Response, error) {
		t.Fatal("unconfigured promotions reached provider")
		return nil, nil
	})
	message := testACSMessage()
	message.SenderClass = "promotion"
	if err := sender.Send(context.Background(), message); err == nil {
		t.Fatal("missing promotional sender silently fell back")
	}
}
