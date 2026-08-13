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
// disque : le PDF lui-même plus un petit fichier JSON de métadonnées. Pas
// besoin de table en base — même approche légère que les sauvegardes BDD.
type Service struct {
	dir string
}

type Metadata struct {
	GeneratedAt time.Time `json:"generated_at"`
	Criteria    string    `json:"criteria"`
	SizeBytes   int64     `json:"size_bytes"`
}

const pdfFilename = "latest.pdf"
const metadataFilename = "latest.json"

var ErrNoReport = errors.New("aucun rapport généré pour l'instant")

func NewService(dir string) (*Service, error) {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, err
	}

	return &Service{dir: dir}, nil
}

func (s *Service) pdfPath() string {
	return filepath.Join(s.dir, pdfFilename)
}

func (s *Service) metadataPath() string {
	return filepath.Join(s.dir, metadataFilename)
}

// SaveLatest écrase le rapport précédemment généré (s'il existe) par un
// nouveau — seul le rapport le plus récent est jamais conservé.
func (s *Service) SaveLatest(criteria string, content io.Reader) (Metadata, error) {
	pdfPath := s.pdfPath()

	file, err := os.Create(pdfPath)
	if err != nil {
		return Metadata{}, err
	}

	size, err := io.Copy(file, content)
	if err != nil {
		file.Close()
		os.Remove(pdfPath)
		return Metadata{}, err
	}
	file.Close()

	meta := Metadata{
		GeneratedAt: time.Now(),
		Criteria:    criteria,
		SizeBytes:   size,
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

func (s *Service) OpenLatestPdf() (*os.File, error) {
	file, err := os.Open(s.pdfPath())
	if err != nil {
		if os.IsNotExist(err) {
			return nil, ErrNoReport
		}
		return nil, err
	}

	return file, nil
}
