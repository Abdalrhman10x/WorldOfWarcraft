package store

import (
	"errors"
	"testing"
	"time"

	"WorldOfWarcraft/internal/domain"
)

func TestCreate(t *testing.T) {
	s := NewMemory()
	c, err := s.Create("thrall", domain.Orc, domain.Shaman)
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if c.Name != "thrall" {
		t.Errorf("Name = %q, want %q", c.Name, "thrall")
	}
	if c.Level != 1 {
		t.Errorf("Level = %d, want 1", c.Level)
	}
	if c.MaxHealth <= 0 {
		t.Errorf("MaxHealth = %d, want > 0", c.MaxHealth)
	}
}

func TestCreateNormalizesNameBeforeKeying(t *testing.T) {
	s := NewMemory()
	if _, err := s.Create("  Thrall  ", domain.Orc, domain.Shaman); err != nil {
		t.Fatalf("Create: %v", err)
	}

	if _, err := s.Get("thrall"); err != nil {
		t.Errorf("Get(thrall) after creating %q: %v", "  Thrall  ", err)
	}
}

func TestCreateRejectsDuplicateRegardlessOfCase(t *testing.T) {
	s := NewMemory()
	if _, err := s.Create("Thrall", domain.Orc, domain.Shaman); err != nil {
		t.Fatalf("first Create: %v", err)
	}

	_, err := s.Create("thrall", domain.Orc, domain.Shaman)
	if !errors.Is(err, ErrAlreadyExists) {
		t.Fatalf("err = %v, want ErrAlreadyExists", err)
	}
	if got := len(s.List()); got != 1 {
		t.Fatalf("stored %d characters, want 1: case variants must not split", got)
	}
}

func TestCreateRejectsInvalidCharacter(t *testing.T) {
	s := NewMemory()
	_, err := s.Create("sylvanas", domain.Orc, domain.Druid)
	if !errors.Is(err, domain.ErrInvalidClassCombo) {
		t.Errorf("err = %v, want ErrInvalidClassCombo", err)
	}
	if got := len(s.List()); got != 0 {
		t.Errorf("stored %d characters after failed Create, want 0", got)
	}
}

func TestGetNotFound(t *testing.T) {
	s := NewMemory()
	if _, err := s.Get("nobody"); !errors.Is(err, ErrNotFound) {
		t.Errorf("err = %v, want ErrNotFound", err)
	}
}

func TestWithMutatesAndPersists(t *testing.T) {
	s := NewMemory()
	if _, err := s.Create("thrall", domain.Orc, domain.Shaman); err != nil {
		t.Fatalf("Create: %v", err)
	}

	before, err := s.Get("thrall")
	if err != nil {
		t.Fatalf("Get: %v", err)
	}

	err = s.With("thrall", func(c *domain.Character) error {
		_, err := c.AddXp(before.MaxXp)
		return err
	})
	if err != nil {
		t.Fatalf("With: %v", err)
	}

	after, err := s.Get("thrall")
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if after.Level != 2 {
		t.Errorf("Level = %d, want 2", after.Level)
	}
}

func TestWithPropagatesCallbackError(t *testing.T) {
	s := NewMemory()
	if _, err := s.Create("thrall", domain.Orc, domain.Shaman); err != nil {
		t.Fatalf("Create: %v", err)
	}

	boom := errors.New("boom")
	err := s.With("thrall", func(c *domain.Character) error {
		_, addErr := c.AddXp(c.MaxXp)
		if addErr != nil {
			return addErr
		}
		return boom
	})
	if !errors.Is(err, boom) {
		t.Errorf("err = %v, want boom", err)
	}

	after, getErr := s.Get("thrall")
	if getErr != nil {
		t.Fatalf("Get: %v", getErr)
	}
	if after.Level != 2 {
		t.Errorf("Level = %d, want 2: callback mutations must persist despite the returned error", after.Level)
	}
}

func TestWithCallbackMustNotReenterStore(t *testing.T) {
	s := NewMemory()
	if _, err := s.Create("thrall", domain.Orc, domain.Shaman); err != nil {
		t.Fatalf("Create: %v", err)
	}

	// With holds the write lock for the whole callback, and sync.RWMutex is
	// not reentrant, so calling Get from inside it deadlocks forever. The
	// callback already receives the character; there is no reason to reenter.
	finished := make(chan struct{})
	go func() {
		defer close(finished)
		_ = s.With("thrall", func(c *domain.Character) error {
			if _, err := c.AddXp(c.MaxXp); err != nil {
				return err
			}
			return nil
		})
	}()

	select {
	case <-finished:
	case <-time.After(2 * time.Second):
		t.Fatal("With did not complete: the callback must not call back into the store")
	}
}

func TestWithNotFound(t *testing.T) {
	s := NewMemory()
	called := false
	err := s.With("nobody", func(*domain.Character) error {
		called = true
		return nil
	})
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("err = %v, want ErrNotFound", err)
	}
	if called {
		t.Error("callback ran for a missing character")
	}
}

func TestListIsSortedAndNonNil(t *testing.T) {
	s := NewMemory()
	if got := s.List(); got == nil {
		t.Fatal("List() = nil, want empty slice (would marshal as null)")
	}

	for _, n := range []string{"zulu", "alpha", "mike"} {
		if _, err := s.Create(domain.Name(n), domain.Orc, domain.Shaman); err != nil {
			t.Fatalf("Create(%q): %v", n, err)
		}
	}

	got := s.List()
	for i := 1; i < len(got); i++ {
		if got[i-1].Name > got[i].Name {
			t.Fatalf("List not sorted: %q before %q", got[i-1].Name, got[i].Name)
		}
	}
}

func TestDelete(t *testing.T) {
	s := NewMemory()
	if _, err := s.Create("thrall", domain.Orc, domain.Shaman); err != nil {
		t.Fatalf("Create: %v", err)
	}
	if err := s.Delete("thrall"); err != nil {
		t.Fatalf("Delete: %v", err)
	}
	if _, err := s.Get("thrall"); !errors.Is(err, ErrNotFound) {
		t.Errorf("Get after Delete err = %v, want ErrNotFound", err)
	}
	if err := s.Delete("thrall"); !errors.Is(err, ErrNotFound) {
		t.Errorf("second Delete err = %v, want ErrNotFound", err)
	}
}

func TestConcurrentAccessIsRaceFree(t *testing.T) {
	s := NewMemory()
	const workers = 8

	done := make(chan struct{})
	for i := 0; i < workers; i++ {
		go func(i int) {
			defer func() { done <- struct{}{} }()
			name := domain.Name(string(rune('a' + i)))
			if _, err := s.Create(name, domain.Orc, domain.Shaman); err != nil {
				t.Errorf("Create: %v", err)
				return
			}
			if err := s.With(name, func(c *domain.Character) error {
				_, err := c.AddXp(10)
				return err
			}); err != nil {
				t.Errorf("With: %v", err)
			}
			s.List()
			s.Get(name)
		}(i)
	}
	for i := 0; i < workers; i++ {
		<-done
	}

	if got := len(s.List()); got != workers {
		t.Errorf("stored %d characters, want %d", got, workers)
	}
}
