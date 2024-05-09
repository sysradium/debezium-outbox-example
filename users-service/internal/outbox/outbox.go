package outbox

import (
	"context"

	"google.golang.org/protobuf/proto"
)

type Storer interface {
	Store(context.Context, string, proto.Message) error
}
