package migration

import (
	"breast-implant-warranty-system/core/singleton"
	"breast-implant-warranty-system/pkg/dbutil"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/sirupsen/logrus"
)

type GMMedRouteConfigItf interface {
	singleton.WriteDBConfigItf
}

// Migrate 執行資料庫遷移
func Migrate(ctx context.Context, singletonGroup *singleton.Group, cfg GMMedRouteConfigItf) error {
	writeDB := singletonGroup.GetWriteDB(ctx, cfg)

	logrus.Info("Starting database migration...")

	// 建立遷移記錄表
	if err := createMigrationTable(ctx, writeDB); err != nil {
		return fmt.Errorf("failed to create migration table: %w", err)
	}

	// 取得遷移文件列表
	migrationFiles, err := getMigrationFiles()
	if err != nil {
		return fmt.Errorf("failed to get migration files: %w", err)
	}

	// 取得已執行的遷移
	executedMigrations, err := getExecutedMigrations(ctx, writeDB)
	if err != nil {
		return fmt.Errorf("failed to get executed migrations: %w", err)
	}

	// 執行未執行的遷移
	for _, file := range migrationFiles {
		if _, executed := executedMigrations[file]; !executed {
			if err := executeMigration(ctx, writeDB, file); err != nil {
				return fmt.Errorf("failed to execute migration %s: %w", file, err)
			}
			logrus.Infof("Migration %s executed successfully", file)
		} else {
			logrus.Debugf("Migration %s already executed, skipping", file)
		}
	}

	logrus.Info("Database migration completed successfully")
	return nil
}

// createMigrationTable 建立遷移記錄表
func createMigrationTable(ctx context.Context, db dbutil.PgxClientItf) error {
	query := `
		CREATE TABLE IF NOT EXISTS schema_migrations (
			id SERIAL PRIMARY KEY,
			filename VARCHAR(255) UNIQUE NOT NULL,
			executed_at TIMESTAMP DEFAULT NOW()
		)
	`
	_, err := db.Exec(ctx, query)
	return err
}

// getMigrationFiles 取得遷移文件列表
func getMigrationFiles() ([]string, error) {
	files, err := os.ReadDir("migrations")
	if err != nil {
		return nil, err
	}

	var migrationFiles []string
	for _, file := range files {
		if !file.IsDir() && strings.HasSuffix(file.Name(), ".sql") {
			migrationFiles = append(migrationFiles, file.Name())
		}
	}

	// 按文件名排序確保執行順序
	sort.Strings(migrationFiles)
	return migrationFiles, nil
}

// getExecutedMigrations 取得已執行的遷移
func getExecutedMigrations(ctx context.Context, db dbutil.PgxClientItf) (map[string]bool, error) {
	query := "SELECT filename FROM schema_migrations"
	rows, err := db.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	executed := make(map[string]bool)
	for rows.Next() {
		var filename string
		if err := rows.Scan(&filename); err != nil {
			return nil, err
		}
		executed[filename] = true
	}

	return executed, rows.Err()
}

// executeMigration 執行單個遷移文件
func executeMigration(ctx context.Context, db dbutil.PgxClientItf, filename string) error {
	// 讀取遷移文件
	content, err := os.ReadFile(filepath.Join("migrations", filename))
	if err != nil {
		return fmt.Errorf("failed to read migration file: %w", err)
	}

	// 開始事務
	tx, err := db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	// 執行遷移 SQL
	if _, err := tx.Exec(ctx, string(content)); err != nil {
		return fmt.Errorf("failed to execute migration SQL: %w", err)
	}

	// 記錄遷移執行
	if _, err := tx.Exec(ctx, "INSERT INTO schema_migrations (filename) VALUES ($1)", filename); err != nil {
		return fmt.Errorf("failed to record migration: %w", err)
	}

	// 提交事務
	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit migration transaction: %w", err)
	}

	return nil
}
