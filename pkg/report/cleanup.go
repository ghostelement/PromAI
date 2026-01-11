package report

import (
	"PromAI/pkg/database"
	"os"
	"time"
)

// CleanupReports 清理旧报告
func CleanupReports(dbClient *database.DBclient, maxAge int) error {
	reportsDir := "reports"
	now := time.Now()

	// 计算清理报告的时间线
	cleanTime := now.Add(-time.Duration(maxAge) * 24 * time.Hour)
	// 查询报告
	lists, err := dbClient.GetReportHistoryByTime(cleanTime)
	if err != nil {
		return err
	}
	for _, list := range lists {
		// 删除报告
		if err := dbClient.DeleteReportHistory(list.ID); err != nil {
			return err
		}
		os.Remove(reportsDir + "/" + list.ReportUrl)
	}
	return nil
}
