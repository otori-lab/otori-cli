package templates

import (
	"embed"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
)

//go:embed data/*
var templatesFS embed.FS

// Extract extrait les templates du type donné vers le dossier destination
// templateType: "classic" ou "ia"
// destDir: chemin du dossier profil (ex: ~/.otori/profiles/mon-profil)
func Extract(templateType string, destDir string) error {
	srcPath := "data/" + templateType

	// Vérifier que le template existe
	if _, err := templatesFS.ReadDir(srcPath); err != nil {
		return fmt.Errorf("template type '%s' not found: %w", templateType, err)
	}

	return fs.WalkDir(templatesFS, srcPath, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		// Calculer le chemin relatif
		relPath, err := filepath.Rel(srcPath, path)
		if err != nil {
			return err
		}

		// Ignorer la racine
		if relPath == "." {
			return nil
		}

		destPath := filepath.Join(destDir, relPath)

		if d.IsDir() {
			return os.MkdirAll(destPath, 0755)
		}

		// Lire le contenu du fichier embarqué
		content, err := templatesFS.ReadFile(path)
		if err != nil {
			return fmt.Errorf("error reading embedded file %s: %w", path, err)
		}

		// Écrire le fichier
		return os.WriteFile(destPath, content, 0644)
	})
}

// GetAvailableTypes retourne la liste des types de templates disponibles
func GetAvailableTypes() []string {
	return []string{"classic", "ia"}
}

// HasType vérifie si un type de template existe
func HasType(templateType string) bool {
	srcPath := "data/" + templateType
	_, err := templatesFS.ReadDir(srcPath)
	return err == nil
}
