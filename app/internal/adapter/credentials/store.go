package credentials

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"syscall"

	port "agentcontrolplane/app/internal/port/credentials"
)

// NewStore öffnet einen lokalen Secret-Speicher mit getrenntem Masterschlüssel.
func NewStore(secretPath, keyPath string) (*Store, error) {
	if secretPath == "" || keyPath == "" || filepath.Clean(secretPath) == filepath.Clean(keyPath) {
		return nil, errors.New("separate secret and key paths required")
	}

	s := &Store{secretPath: filepath.Clean(secretPath), keyPath: filepath.Clean(keyPath)}
	if err := s.preparePaths(); err != nil {
		return nil, err
	}

	if err := s.openKey(); err != nil {
		return nil, err
	}

	if _, err := s.readAll(); err != nil {
		return nil, err
	}

	return s, nil
}

func (s *Store) preparePaths() error {
	if err := s.prepareDir(filepath.Dir(s.secretPath)); err != nil {
		return err
	}

	return s.prepareDir(filepath.Dir(s.keyPath))
}

func (s *Store) Save(ctx context.Context, name string, secret []byte) error {
	if err := s.validate(ctx, name); err != nil {
		return err
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	all, err := s.readAll()
	if err != nil {
		return err
	}

	all[name] = append([]byte(nil), secret...)
	return s.writeAll(all)
}

func (s *Store) Load(ctx context.Context, name string) ([]byte, error) {
	if err := s.validate(ctx, name); err != nil {
		return nil, err
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	all, err := s.readAll()
	if err != nil {
		return nil, err
	}

	secret, ok := all[name]
	if !ok {
		return nil, port.ErrNotFound
	}

	return append([]byte(nil), secret...), nil
}

func (s *Store) Delete(ctx context.Context, name string) error {
	if err := s.validate(ctx, name); err != nil {
		return err
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	all, err := s.readAll()
	if err != nil {
		return err
	}

	delete(all, name)
	return s.writeAll(all)
}

func (s *Store) validate(ctx context.Context, name string) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	if name == "" {
		return errors.New("credential name required")
	}

	return nil
}

func (s *Store) prepareDir(path string) error {
	if err := os.MkdirAll(path, 0700); err != nil {
		return fmt.Errorf("create credential directory: %w", err)
	}

	info, err := os.Lstat(path)
	if err != nil {
		return err
	}

	if !info.IsDir() || info.Mode().Perm() != 0700 {
		return errors.New("credential directory must have mode 0700")
	}

	return nil
}

func (s *Store) openKey() error {
	data, err := s.readSecure(s.keyPath)
	if err == nil {
		return s.acceptKey(data)
	}

	if !errors.Is(err, os.ErrNotExist) {
		return err
	}

	if err := s.requireNoSecret(); err != nil {
		return err
	}

	return s.createKey()
}

func (s *Store) requireNoSecret() error {
	_, err := os.Lstat(s.secretPath)
	if err == nil {
		return errors.New("secret data exists without master key")
	}

	if !errors.Is(err, os.ErrNotExist) {
		return err
	}

	return nil
}

func (s *Store) acceptKey(data []byte) error {
	if len(data) != 32 {
		return errors.New("invalid master key length")
	}

	s.key = data
	return nil
}

func (s *Store) createKey() error {
	key := make([]byte, 32)
	if _, err := io.ReadFull(rand.Reader, key); err != nil {
		return err
	}

	f, err := os.OpenFile(s.keyPath, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600)
	if err != nil {
		return fmt.Errorf("create master key: %w", err)
	}

	if err := s.writeKey(f, key); err != nil {
		return err
	}

	s.key = key
	return nil
}

func (s *Store) writeKey(f *os.File, key []byte) error {
	_, err := f.Write(key)
	if err == nil {
		err = f.Sync()
	}

	closeErr := f.Close()
	if err != nil || closeErr != nil {
		os.Remove(s.keyPath)
		return errors.Join(err, closeErr)
	}

	return s.syncDir(filepath.Dir(s.keyPath))
}

func (s *Store) readSecure(path string) ([]byte, error) {
	fd, err := syscall.Open(path, syscall.O_RDONLY|syscall.O_NOFOLLOW, 0)
	if err != nil {
		return nil, err
	}

	f := os.NewFile(uintptr(fd), path)
	defer f.Close()
	info, err := f.Stat()
	if err != nil {
		return nil, err
	}

	if !info.Mode().IsRegular() || info.Mode().Perm() != 0600 {
		return nil, errors.New("credential file must be regular with mode 0600")
	}

	return io.ReadAll(f)
}

func (s *Store) readAll() (map[string][]byte, error) {
	data, err := s.readSecure(s.secretPath)
	if errors.Is(err, os.ErrNotExist) {
		return make(map[string][]byte), nil
	}

	if err != nil {
		return nil, err
	}

	plain, err := s.decrypt(data)
	if err != nil {
		return nil, err
	}

	all := make(map[string][]byte)
	err = json.Unmarshal(plain, &all)
	return all, err
}

func (s *Store) decrypt(data []byte) ([]byte, error) {
	block, err := aes.NewCipher(s.key)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	if len(data) < gcm.NonceSize() {
		return nil, errors.New("invalid encrypted credentials")
	}

	return gcm.Open(nil, data[:gcm.NonceSize()], data[gcm.NonceSize():], nil)
}

func (s *Store) encrypt(plain []byte) ([]byte, error) {
	block, err := aes.NewCipher(s.key)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}

	return gcm.Seal(nonce, nonce, plain, nil), nil
}

func (s *Store) writeAll(all map[string][]byte) error {
	plain, err := json.Marshal(all)
	if err != nil {
		return err
	}

	data, err := s.encrypt(plain)
	if err != nil {
		return err
	}

	return s.atomicWrite(data)
}

func (s *Store) atomicWrite(data []byte) error {
	if err := s.checkSecret(); err != nil {
		return err
	}

	f, err := os.CreateTemp(filepath.Dir(s.secretPath), ".credentials-*")
	if err != nil {
		return err
	}

	defer os.Remove(f.Name())
	return s.commitTemp(f, data)
}

func (s *Store) commitTemp(f *os.File, data []byte) error {
	if err := s.writeTemp(f, data); err != nil {
		return err
	}

	if err := os.Rename(f.Name(), s.secretPath); err != nil {
		return err
	}

	return s.syncDir(filepath.Dir(s.secretPath))
}

func (s *Store) checkSecret() error {
	_, err := s.readSecure(s.secretPath)
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return err
	}

	return nil
}

func (s *Store) writeTemp(f *os.File, data []byte) error {
	if err := f.Chmod(0600); err != nil {
		f.Close()
		return err
	}

	_, err := f.Write(data)
	if err == nil {
		err = f.Sync()
	}

	return errors.Join(err, f.Close())
}

func (s *Store) syncDir(path string) error {
	d, err := os.Open(path)
	if err != nil {
		return err
	}

	defer d.Close()
	return d.Sync()
}
