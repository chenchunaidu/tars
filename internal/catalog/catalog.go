package catalog

import (
	"encoding/json"
	"os"
	"path/filepath"
	"time"

	"agent-tools/internal/formula"
	"agent-tools/internal/paths"
)

// ToolRecord is the merged, model-facing view for a single tool.
// Written to ~/.agent-tools/catalog/tools.json and per-tool model.json.
type ToolRecord struct {
	Name        string           `json:"name"`
	Version     string           `json:"version"`
	Tap         string           `json:"tap,omitempty"`
	InstallPath string           `json:"install_path"`
	ArtifactURL string           `json:"artifact_url"`
	SHA256      string           `json:"sha256_verified"`
	UpdatedAt   time.Time        `json:"updated_at"`
	Model       *formula.ModelMeta `json:"model,omitempty"`
}

// Index is the aggregate file any agent/model can read.
type Index struct {
	Schema  string       `json:"schema"`
	Version int          `json:"version"`
	Updated time.Time    `json:"updated_at"`
	Tools   []ToolRecord `json:"tools"`
}

const schemaURL = "https://agent-tools.dev/schema/catalog-v1.json"

// WriteTool writes per-tool model.json under install path and refreshes global index.
func WriteTool(rec ToolRecord) error {
	cdir, err := paths.CatalogDir()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(cdir, 0o755); err != nil {
		return err
	}
	if rec.InstallPath != "" {
		modelPath := filepath.Join(rec.InstallPath, "model.json")
		b, err := json.MarshalIndent(rec, "", "  ")
		if err != nil {
			return err
		}
		if err := os.WriteFile(modelPath, b, 0o644); err != nil {
			return err
		}
	}
	return mergeIndex(rec)
}

func mergeIndex(newRec ToolRecord) error {
	cdir, err := paths.CatalogDir()
	if err != nil {
		return err
	}
	idxPath := filepath.Join(cdir, "tools.json")

	var idx Index
	if b, err := os.ReadFile(idxPath); err == nil {
		_ = json.Unmarshal(b, &idx)
	}
	idx.Schema = schemaURL
	idx.Version = 1
	idx.Updated = time.Now().UTC()

	byName := map[string]int{}
	for i, t := range idx.Tools {
		byName[t.Name] = i
	}
	if i, ok := byName[newRec.Name]; ok {
		idx.Tools[i] = newRec
	} else {
		idx.Tools = append(idx.Tools, newRec)
	}

	out, err := json.MarshalIndent(idx, "", "  ")
	if err != nil {
		return err
	}
	tmp := idxPath + ".tmp"
	if err := os.WriteFile(tmp, out, 0o644); err != nil {
		return err
	}
	return os.Rename(tmp, idxPath)
}

// RemoveTool drops a tool from the aggregate index (per-install model.json may remain until dir removed).
func RemoveTool(name string) error {
	cdir, err := paths.CatalogDir()
	if err != nil {
		return err
	}
	idxPath := filepath.Join(cdir, "tools.json")
	b, err := os.ReadFile(idxPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	var idx Index
	if err := json.Unmarshal(b, &idx); err != nil {
		return err
	}
	filtered := idx.Tools[:0]
	for _, t := range idx.Tools {
		if t.Name != name {
			filtered = append(filtered, t)
		}
	}
	idx.Tools = filtered
	idx.Updated = time.Now().UTC()
	out, err := json.MarshalIndent(idx, "", "  ")
	if err != nil {
		return err
	}
	tmp := idxPath + ".tmp"
	if err := os.WriteFile(tmp, out, 0o644); err != nil {
		return err
	}
	return os.Rename(tmp, idxPath)
}
