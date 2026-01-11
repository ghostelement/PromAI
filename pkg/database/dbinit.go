package database

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	_ "modernc.org/sqlite"
)

type DBclient struct {
	*sql.DB
}

// 创建sqlite数据库连接
func NewDBClient() (*DBclient, error) {
	db, err := sql.Open("sqlite", "data/sqlite.db")
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// 设置连接池参数
	db.SetMaxOpenConns(5)
	db.SetMaxIdleConns(2)
	db.SetConnMaxLifetime(5 * time.Minute)
	db.SetConnMaxIdleTime(10 * time.Second)
	// 配置 SQLite 特定选项
	_, err = db.Exec("PRAGMA journal_mode=WAL;")
	if err != nil {
		return nil, fmt.Errorf("failed to set WAL mode: %w", err)
	}

	_, err = db.Exec("PRAGMA synchronous=NORMAL;")
	if err != nil {
		return nil, fmt.Errorf("failed to set synchronous mode: %w", err)
	}

	_, err = db.Exec("PRAGMA cache_size=1000;")
	if err != nil {
		return nil, fmt.Errorf("failed to set cache size: %w", err)
	}

	_, err = db.Exec("PRAGMA temp_store=memory;")
	if err != nil {
		return nil, fmt.Errorf("failed to set temp store: %w", err)
	}

	// 测试连接
	if err = db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	// 检查表是否存在
	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name='report_history'").Scan(&count)
	if err != nil {
		return nil, fmt.Errorf("failed to check table existence: %w", err)
	}

	if count == 0 {
		// 创建表
		createTableSQL := `
            CREATE TABLE report_history (
                id INTEGER PRIMARY KEY AUTOINCREMENT,
                project_name TEXT NOT NULL,
                datasource TEXT NOT NULL,
				report_url  TEXT NOT NULL,
				task_id TEXT,
				task_time INTEGER,
				file_size INTEGER,
                created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
                max_value REAL,
                min_value REAL,
                average REAL,
                total_count INTEGER NOT NULL,
                alert_count INTEGER NOT NULL,
                critical_count INTEGER NOT NULL,
                warning_count INTEGER NOT NULL,
                delete_time TIMESTAMP
            );
            CREATE INDEX idx_report_history_id ON report_history(id);
            CREATE INDEX idx_report_history_project_name ON report_history(project_name);
            CREATE INDEX idx_report_history_datasource ON report_history(datasource);
            CREATE INDEX idx_report_history_created_at ON report_history(created_at);
        `

		if _, err = db.Exec(createTableSQL); err != nil {
			return nil, fmt.Errorf("failed to create table: %w", err)
		}

		log.Println("Report history table created successfully")
	}

	log.Println("Database initialized successfully")
	return &DBclient{db}, nil
}

// 插入report巡检报告
func (c *DBclient) SaveReportHistory(history *ReportHistory) error {
	query := `
        INSERT INTO report_history 
        (project_name, datasource, report_url, task_id, task_time, file_size, created_at, max_value, min_value, average, total_count, alert_count, critical_count, warning_count) 
        VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
    `
	createAtStr := history.CreateTime.Format("2006-01-02 15:04:05")
	_, err := c.Exec(query,
		history.Project,
		history.Datasource,
		history.ReportUrl,
		history.TaskId,
		history.TaskTime,
		history.FileSize,
		createAtStr,
		history.MaxValue,
		history.MinValue,
		history.Average,
		history.TotalCount,
		history.AlertCount,
		history.CriticalCount,
		history.WarningCount,
	)

	if err != nil {
		return fmt.Errorf("failed to save report history: %w", err)
	}

	return nil
}

// 查询报告
func (c *DBclient) GetReportHistory(limit, offset int) ([]ReportHistory, error) {
	query := `
        SELECT id, project_name, datasource, report_url, task_id, task_time, file_size, created_at, max_value, min_value, average, total_count, alert_count, critical_count, warning_count
        FROM report_history WHERE delete_time IS NULL
        ORDER BY created_at DESC
        LIMIT ? OFFSET ?
    `

	rows, err := c.Query(query, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to query report history: %w", err)
	}
	defer rows.Close()

	var histories []ReportHistory
	for rows.Next() {
		var history ReportHistory
		err := rows.Scan(
			&history.ID,
			&history.Project,
			&history.Datasource,
			&history.ReportUrl,
			&history.TaskId,
			&history.TaskTime,
			&history.FileSize,
			&history.CreateTime,
			&history.MaxValue,
			&history.MinValue,
			&history.Average,
			&history.TotalCount,
			&history.AlertCount,
			&history.CriticalCount,
			&history.WarningCount,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan report history: %w", err)
		}
		histories = append(histories, history)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating over rows: %w", err)
	}

	return histories, nil
}

// 根据id查询报告
func (c *DBclient) GetReportHistoryById(id uint) (*ReportHistory, error) {
	query := `
        SELECT id, project_name, datasource, report_url, task_id, task_time, file_size, created_at, max_value, min_value, average, total_count, alert_count, critical_count, warning_count
        FROM report_history WHERE delete_time IS NULL AND id = ?
    `
	var history ReportHistory
	err := c.QueryRow(query, id).Scan(
		&history.ID,
		&history.Project,
		&history.Datasource,
		&history.ReportUrl,
		&history.TaskId,
		&history.TaskTime,
		&history.FileSize,
		&history.CreateTime,
		&history.MaxValue,
		&history.MinValue,
		&history.Average,
		&history.TotalCount,
		&history.AlertCount,
		&history.CriticalCount,
		&history.WarningCount,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to scan report history: %w", err)
	}
	return &history, nil
}

// 根据创建时间查询所有报告id
func (c *DBclient) GetReportHistoryByTime(createTime time.Time) ([]ReportHistory, error) {
	query := `
        SELECT id, report_url , created_at
        FROM report_history WHERE delete_time IS NULL AND created_at <= ?
        ORDER BY created_at DESC
    `

	rows, err := c.Query(query, createTime)
	if err != nil {
		return nil, fmt.Errorf("failed to query report history: %w", err)
	}
	defer rows.Close()

	var histories []ReportHistory
	for rows.Next() {
		var history ReportHistory
		err := rows.Scan(
			&history.ID,
			&history.ReportUrl,
			&history.CreateTime,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan report history: %w", err)
		}
		histories = append(histories, history)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating over rows: %w", err)
	}

	return histories, nil
}

// 删除报告
func (c *DBclient) DeleteReportHistory(id uint) error {
	query := `
        UPDATE report_history SET delete_time = CURRENT_TIMESTAMP WHERE id = ?
    `

	_, err := c.Exec(query, id)
	if err != nil {
		return fmt.Errorf("failed to delete report history: %w", err)
	}

	return nil
}

// CloseDB 关闭数据库连接
func (c *DBclient) CloseDB() {
	c.DB.Close()
}
