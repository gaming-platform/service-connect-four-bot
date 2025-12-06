package bot

import "context"

type Bot interface {
	Play(ctx context.Context) error
}
