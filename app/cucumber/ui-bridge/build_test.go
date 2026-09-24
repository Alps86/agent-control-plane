package uibridge

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
)

func (s *Suite) verifyCopyTree(relative string) error {
	root := filepath.Join("../../../ui/web", relative)
	entries, err := os.ReadDir(root)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		child := filepath.Join(relative, entry.Name())
		if err := s.verifyCopy(child, entry.IsDir()); err != nil {
			return err
		}
	}

	return nil
}

func (s *Suite) verifyCopy(relative string, directory bool) error {
	if directory {
		return s.verifyCopyTree(relative)
	}

	source, err := os.ReadFile(filepath.Join("../../../ui/web", relative))
	if err != nil {
		return err
	}

	copy, err := os.ReadFile(filepath.Join("../../../ui/bridge/dist", relative))
	if err != nil || !bytes.Equal(source, copy) {
		return fmt.Errorf("build copy differs: %s: %v", relative, err)
	}

	return nil
}

func (s *Suite) verifyEmbeddedAssets() error {
	entries, err := os.ReadDir("../../../ui/web/dist/assets")
	if err != nil {
		return err
	}

	for _, entry := range entries {
		if err := s.verifyEmbeddedAsset(entry.Name()); err != nil {
			return err
		}
	}

	return nil
}

func (s *Suite) verifyEmbeddedAsset(name string) error {
	expected, err := os.ReadFile(filepath.Join("../../../ui/web/dist/assets", name))
	if err != nil {
		return err
	}

	actual, status, err := s.fetch("/assets/" + name)
	if err != nil || status != 200 || !bytes.Equal(expected, []byte(actual)) {
		return fmt.Errorf("embedded asset differs: %s HTTP %d: %v", name, status, err)
	}

	return nil
}
