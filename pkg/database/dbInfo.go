package database

import (
	"time"
)

type ReportHistory struct {
	ID            uint      `json:"report_id" gorm:"column:id;primaryKey;autoIncrement"`
	CreateTime    time.Time `json:"created_at" gorm:"column:created_at;comment:创建时间" binding:"required"`
	Project       string    `json:"project" gorm:"column:project_name;comment:项目名称" binding:"required"`
	Datasource    string    `json:"datasource" gorm:"column:datasource;comment:数据源" binding:"required"`
	ReportUrl     string    `json:"report_url" gorm:"column:report_url;comment:报告地址"`
	MaxValue      float64   `json:"max_value" gorm:"column:max_value;comment:最大值"`
	MinValue      float64   `json:"min_value" gorm:"column:min_value;comment:最小值"`
	Average       float64   `json:"average" gorm:"column:average;comment:平均值"`
	AlertCount    int       `json:"alert_count" gorm:"column:alert_count;comment:告警数量"`         // 告警数量
	CriticalCount int       `json:"critical_count" gorm:"column:critical_count;comment:严重告警数量"` // 严重告警数量
	WarningCount  int       `json:"warning_count" gorm:"column:warning_count;comment:警告数量"`     // 警告数量
	TotalCount    int       `json:"total_count" gorm:"column:total_count;comment:总指标数"`         // 总指标数
	DeleteTime    time.Time `json:"delete_time" gorm:"column:delete_time;comment:删除时间"`
}
