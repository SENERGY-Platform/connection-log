package controller

import (
	"context"
	"fmt"
	"time"

	"github.com/SENERGY-Platform/connection-log/pkg/model"
)

// GetOfflineSince lists the ids whose last stored state is disconnected, with the time of that state,
// oldest first. Ids without any stored state are not listed.
func (this *Controller) GetOfflineSince(ctx context.Context, ids []string, kind string) ([]model.OfflineSinceResponse, error) {
	table, err := historyTable(kind)
	if err != nil {
		return nil, err
	}
	result := []model.OfflineSinceResponse{}
	if len(ids) == 0 {
		return result, nil
	}
	stmt, args := offlineSinceStatement(table, ids)
	rows, err := this.pg.Query(ctx, stmt, args...)
	if err != nil {
		return nil, fmt.Errorf("unable to query offline states: %w", err)
	}
	defer rows.Close()
	for rows.Next() {
		var id string
		var since time.Time
		if err := rows.Scan(&id, &since); err != nil {
			return nil, fmt.Errorf("unable to read offline states: %w", err)
		}
		result = append(result, model.OfflineSinceResponse{ID: id, OfflineSince: since})
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("unable to read offline states: %w", err)
	}
	return result, nil
}
