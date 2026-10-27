package service

import (
	"fmt"
	"time"
)

func addIssue(update *Update, code, message, batchID, productID string, now time.Time) {
	for _, issue := range update.Issues {
		if issue.Code == code && issue.Message == message && issue.BatchID == batchID && issue.ProductID == productID && !issue.Resolved {
			return
		}
	}
	id := fmt.Sprintf("issue-%d-%d", now.UnixNano(), len(update.Issues)+1)
	update.Issues = append(update.Issues, Issue{
		ID: id, Code: code, Message: message, BatchID: batchID, ProductID: productID, CreatedAt: now,
	})
}
