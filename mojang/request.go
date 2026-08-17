package mojang

import (
	"context"
	"time"

	"github.com/valyala/fasthttp"
)

const (
	jsonContentType = "application/json"
	timeout         = 10 * time.Second
)

// request is the part of an HTTP call that actually differs between the OAuth,
// Xbox, Yggdrasil and session endpoints; everything shared lives in send.
type request struct {
	client      *fasthttp.Client
	method      string
	url         string
	contentType string
	accept      string
	bearer      string
	body        []byte
	timeout     time.Duration
}

// send answers with the body copied out of the pooled response, so a caller may
// hold on to it after the response has been released
func send(ctx context.Context, r request) ([]byte, int, error) {
	outgoing := fasthttp.AcquireRequest()
	incoming := fasthttp.AcquireResponse()
	defer fasthttp.ReleaseRequest(outgoing)
	defer fasthttp.ReleaseResponse(incoming)

	outgoing.Header.SetMethod(r.method)
	outgoing.SetRequestURI(r.url)

	if r.contentType != "" {
		outgoing.Header.SetContentType(r.contentType)
	}
	if r.accept != "" {
		outgoing.Header.Set(fasthttp.HeaderAccept, r.accept)
	}
	if r.bearer != "" {
		outgoing.Header.Set(fasthttp.HeaderAuthorization, "Bearer "+r.bearer)
	}
	if r.body != nil {
		outgoing.SetBody(r.body)
	}

	if err := r.deliver(outgoing, incoming, until(ctx, r.timeout)); err != nil {
		return nil, 0, err
	}

	return append([]byte(nil), incoming.Body()...), incoming.StatusCode(), nil
}

// a caller that brought no client of its own falls back to the shared default,
// which spares the package from lazily assigning one and racing over it
func (r request) deliver(outgoing *fasthttp.Request, incoming *fasthttp.Response, deadline time.Time) error {
	if r.client == nil {
		return fasthttp.DoDeadline(outgoing, incoming, deadline)
	}

	return r.client.DoDeadline(outgoing, incoming, deadline)
}

// the per-call timeout is a ceiling, never an extension of the caller's deadline
func until(ctx context.Context, timeout time.Duration) time.Time {
	latest := time.Now().Add(timeout)

	bound, deadlined := ctx.Deadline()
	if deadlined && bound.Before(latest) {
		return bound
	}

	return latest
}
