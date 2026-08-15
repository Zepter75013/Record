package report

import (
	"encoding/json"
	"errors"
	"io"
	"os"
	"path/filepath"
	"time"
)

// Service stocke exactement un seul rapport généré (le plus récent) sur
// disque : le fichier lui-même plus un petit fichier JSON de métadonnées. Pas
// besoin de table en base — même approche légère que les sauvegardes BDD.
//
// Le rapport peut être généré dans plusieurs formats (PDF/XLSX/DOCX/CSV) —
// le fichier est donc nommé latest.<ext> et l'extension/type MIME courants
// sont conservés dans les métadonnées pour pouvoir le re-servir correctement.
type Service struct {
	dir string
}

type Metadata struct {
	GeneratedAt time.Time `json:"generated_at"`
	Criteria    string    `json:"criteria"`
	SizeBytes   int64     `json:"size_bytes"`
	Format      string    `json:"format"`
	ContentType string    `json:"content_type"`
}

var formatInfo = map[string]struct {
	Extension   string
	ContentType string
}{
	"pdf":  {"pdf", "application/pdf"},
	"xlsx": {"xlsx", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet"},
	"docx": {"docx", "application/vnd.openxmlformats-officedocument.wordprocessingml.document"},
	"csv":  {"csv", "text/csv"},
}

const metadataFilename = "latest.json"

var ErrNoReport = errors.New("aucun rapport généré pour l'instant")
var ErrUnsupportedFormat = errors.New("format de rapport non supporté")

func NewService(dir string) (*Service, error) {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, err
	}

	return &Service{dir: dir}, nil
}

func (s *Service) filePath(extension string) string {
	return filepath.Join(s.dir, "latest."+extension)
}

func (s *Service) metadataPath() string {
	return filepath.Join(s.dir, metadataFilename)
}

// SaveLatest écrase le rapport précédemment généré (s'il existe, y compris
// dans un autre format) par un nouveau — seul le rapport le plus récent est
// jamais conservé.
func (s *Service) SaveLatest(criteria, format string, content io.Reader) (Metadata, error) {
	info, ok := formatInfo[format]
	if !ok {
		return Metadata{}, ErrUnsupportedFormat
	}

	// Nettoie un précédent rapport généré dans un AUTRE format, pour ne
	// jamais laisser deux fichiers "latest.*" coexister.
	for f, i := range formatInfo {
		if f != format {
			os.Remove(s.filePath(i.Extension))
		}
	}

	path := s.filePath(info.Extension)

	file, err := os.Create(path)
	if err != nil {
		return Metadata{}, err
	}

	size, err := io.Copy(file, content)
	if err != nil {
		file.Close()
		os.Remove(path)
		return Metadata{}, err
	}
	file.Close()

	meta := Metadata{
		GeneratedAt: time.Now(),
		Criteria:    criteria,
		SizeBytes:   size,
		Format:      format,
		ContentType: info.ContentType,
	}

	data, err := json.MarshalIndent(meta, "", "  ")
	if err != nil {
		return Metadata{}, err
	}

	if err := os.WriteFile(s.metadataPath(), data, 0o644); err != nil {
		return Metadata{}, err
	}

	return meta, nil
}

func (s *Service) GetLatestMetadata() (Metadata, error) {
	data, err := os.ReadFile(s.metadataPath())
	if err != nil {
		if os.IsNotExist(err) {
			return Metadata{}, ErrNoReport
		}
		return Metadata{}, err
	}

	var meta Metadata
	if err := json.Unmarshal(data, &meta); err != nil {
		return Metadata{}, err
	}

	return meta, nil
}

func (s *Service) OpenLatestFile() (*os.File, Metadata, error) {
	meta, err := s.GetLatestMetadata()
	if err != nil {
		return nil, Metadata{}, err
	}

	info, ok := formatInfo[meta.Format]
	if !ok {
		return nil, Metadata{}, ErrUnsupportedFormat
	}

	file, err := os.Open(s.filePath(info.Extension))
	if err != nil {
		if os.IsNotExist(err) {
			return nil, Metadata{}, ErrNoReport
		}
		return nil, Metadata{}, err
	}

	return file, meta, nil
}
