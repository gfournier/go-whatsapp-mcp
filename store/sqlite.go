package store

import (
	"context"
	"fmt"

	_ "modernc.org/sqlite"

	"go.mau.fi/whatsmeow/store/sqlstore"
	waLog "go.mau.fi/whatsmeow/util/log"
)

func OpenSQLite(ctx context.Context, path string, log waLog.Logger) (*sqlstore.Container, error) {
	dsn := fmt.Sprintf("file:%s?_foreign_keys=on&_busy_timeout=5000", path)
	container, err := sqlstore.New(ctx, "sqlite", dsn, log)
	if err != nil {
		return nil, fmt.Errorf("open sqlite store: %w", err)
	}
	return container, nil
}
