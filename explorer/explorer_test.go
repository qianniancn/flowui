package explorer

import (
	"context"
	"errors"
	"testing"
)

func TestFunctionsRequireFlowUICommandContext(t *testing.T) {
	if _, err := ChooseFile(context.Background()); !errors.Is(err, ErrUnavailable) {
		t.Fatalf("ChooseFile error = %v, want ErrUnavailable", err)
	}
	if _, err := ChooseFiles(context.Background()); !errors.Is(err, ErrUnavailable) {
		t.Fatalf("ChooseFiles error = %v, want ErrUnavailable", err)
	}
	if _, err := CreateFile(context.Background(), "report.txt"); !errors.Is(err, ErrUnavailable) {
		t.Fatalf("CreateFile error = %v, want ErrUnavailable", err)
	}
}
