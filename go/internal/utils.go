package internal

import (
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
)

func GetDepthSnapshot(snapshotUrl string) *DepthSnapshot {
	resp, err := http.Get(snapshotUrl)
	if err != nil {
		slog.Error("Error getting snapshot", slog.String("snapshotUrl", snapshotUrl))
		return nil
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		slog.Error("Error reading snapshot response", slog.String("snapshotUrl", snapshotUrl))
		return nil
	}
	snapshot := DepthSnapshot{}
	err = json.Unmarshal(body, &snapshot)
	if err != nil {
		slog.Error("Error unmarshalling depth snapshot", slog.String("snapshotUrl", snapshotUrl))
		return nil
	}
	return &snapshot
}
