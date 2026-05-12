package store

import (
	"context"
	"fmt"
	"os"
	"strings"

	_ "modernc.org/sqlite"

	"go.mau.fi/whatsmeow/store/sqlstore"
	waLog "go.mau.fi/whatsmeow/util/log"
)

func OpenSQLite(ctx context.Context, path string, log waLog.Logger) (*sqlstore.Container, error) {
	if strings.ContainsAny(path, "?&#") {
		return nil, fmt.Errorf("WHATSAPP_DB_PATH must be a plain file path, not a URI (got %q)", path)
	}

	dsn := fmt.Sprintf("file:%s?_foreign_keys=on&_busy_timeout=5000", path)
	container, err := sqlstore.New(ctx, "sqlite", dsn, log)
	if err != nil {
		return nil, fmt.Errorf("open sqlite store: %w", err)
	}

	if err := os.Chmod(path, 0600); err != nil && !os.IsNotExist(err) {
		return nil, fmt.Errorf("restrict db permissions: %w", err)
	}

	return container, nil
}
