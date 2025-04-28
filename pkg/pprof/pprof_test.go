package pprof_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/m1khal3v/gometheus/pkg/pprof"
)

func TestCPUCapture(t *testing.T) {
	// Создаем временный файл для теста
	tempDir := t.TempDir()
	filename := filepath.Join(tempDir, "cpu.prof")

	// Устанавливаем контекст с тайм-аутом
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	err := pprof.CPUCapture(ctx, filename, 1*time.Second)
	if err != nil {
		t.Fatalf("CPUCapture failed: %v", err)
	}

	// Проверяем, что файл был создан
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		t.Fatalf("Expected profile file %s to be created", filename)
	}
}

func TestCPUCapture_ContextCancelled(t *testing.T) {
	// Создаем временный файл для теста
	tempDir := t.TempDir()
	filename := filepath.Join(tempDir, "cpu_cancel.prof")

	// Устанавливаем контекст, который будет отменен
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Отменяем контекст сразу

	err := pprof.CPUCapture(ctx, filename, 1*time.Second)
	if err == nil {
		t.Fatalf("Expected error due to context cancellation, but got none")
	}
}

func TestCapture(t *testing.T) {
	// Создаем временный файл для теста
	tempDir := t.TempDir()
	filename := filepath.Join(tempDir, "heap.prof")

	err := pprof.Capture(pprof.Heap, filename)
	if err != nil {
		t.Fatalf("Capture failed: %v", err)
	}

	// Проверяем, что файл был создан
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		t.Fatalf("Expected heap profile file %s to be created", filename)
	}
}

func TestCapture_InvalidProfile(t *testing.T) {
	// Создаем временный файл для теста
	tempDir := t.TempDir()
	filename := filepath.Join(tempDir, "invalid.prof")

	err := pprof.Capture("invalid_profile", filename)
	if err == nil {
		t.Fatalf("Expected error for invalid profile, but got none")
	}

	// Убедимся, что файл не был создан
	if _, err := os.Stat(filename); err == nil {
		t.Fatalf("Profile file should not have been created")
	}
}

func TestCapture_FileCreationFailure(t *testing.T) {
	// Путь к файлу в недоступной директории
	filename := filepath.Join("/invalid_path", "heap.prof")

	err := pprof.Capture(pprof.Heap, filename)
	if err == nil {
		t.Fatalf("Expected error for invalid file path, but got none")
	}
}

func TestCPUCapture_FileCreationFailure(t *testing.T) {
	// Недопустимый путь для теста
	filename := filepath.Join("/invalid_path", "cpu.prof")

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	err := pprof.CPUCapture(ctx, filename, 1*time.Second)
	if err == nil {
		t.Fatalf("Expected error for invalid file path, but got none")
	}
}
