package log

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"reflect"

	"github.com/getsentry/sentry-go"
)

var sentryDsn string
var sentryEnv string

func init() {
	sentryDsn = os.Getenv("LOG_SENTRY_DSN")
	sentryEnv = os.Getenv("LOG_SENTRY_ENV")
	if sentryDsn != "" {
		if sentryEnv == "" {
			sentryEnv = "production"
		}
		err := sentry.Init(sentry.ClientOptions{
			Dsn: sentryDsn,
			// SampleRate: 0.4,
			// Enable printing of SDK debug messages.
			// Useful when getting started or trying to figure something out.
			Debug:       true,
			Environment: sentryEnv,
		})
		if err != nil {
			fmt.Println("SENTRY ERR", err)
			sentryDsn = ""
		}
	}
}

func NewSentryErr(ctx context.Context, err error, code E, internal_message string, field M) error {
	hub := sentry.CurrentHub()
	client, scope := hub.Client(), hub.Scope()
	if client == nil || scope == nil {
		return nil
	}
	event := sentry.NewEvent()
	event.Level = sentry.LevelError

	const maxErrorDepth = 10
	for i := 0; i < maxErrorDepth && err != nil; i++ {
		event.Exception = append(event.Exception, sentry.Exception{
			Value:      err.Error(),
			Type:       reflect.TypeOf(err).String(),
			Stacktrace: sentry.ExtractStacktrace(err),
		})
		switch previous := err.(type) {
		case interface{ Unwrap() error }:
			err = previous.Unwrap()
		case interface{ Cause() error }:
			err = previous.Cause()
		default:
			err = nil
		}
	}

	// Add a trace of the current stack to the most recent error in a chain if
	// it doesn't have a stack trace yet.
	// We only add to the most recent error to avoid duplication and because the
	// current stack is most likely unrelated to errors deeper in the chain.
	if event.Exception[0].Stacktrace == nil {
		event.Exception[0].Stacktrace = sentry.NewStacktrace()
	}

	// event.Exception should be sorted such that the most recent error is last.
	reverseSentry(event.Exception)
	event.Environment = sentryEnv
	event.Fingerprint = []string{internal_message}
	for key, value := range field {
		valueb, _ := json.Marshal(value)
		event.Tags[key] = string(valueb)
	}

	hub.Client().CaptureEvent(event, &sentry.EventHint{OriginalException: err}, scope)
	return err
}

// reverse reverses the slice a in place.
func reverseSentry(a []sentry.Exception) {
	for i := len(a)/2 - 1; i >= 0; i-- {
		opp := len(a) - 1 - i
		a[i], a[opp] = a[opp], a[i]
	}
}
