package repository

import (
	"context"
	"testing"
	"time"
)

// Sessions and roles decide who gets in and what they may touch. Getting either
// wrong is not a crash, it is someone else's account — so these run against a
// real database.
//
// Not covered here: the last-admin guard in withAdminGuard. Reaching it means
// being the only unblocked admin in the database, and the development database
// already has one; demoting it to set the scene would leave the machine without
// an admin if the test were interrupted, and BootstrapAdmin only steps in when
// the users table is completely empty. That path needs an isolated database.

func TestSessionRoundTrip(t *testing.T) {
	pool := billingPool(t)
	ctx := context.Background()
	repo := NewUserRepo(pool)
	userID := testUser(t, pool)

	token, err := repo.CreateSession(ctx, userID)
	if err != nil {
		t.Fatalf("create session: %v", err)
	}
	if token == "" {
		t.Fatal("empty session token")
	}

	got, err := repo.GetUserBySession(ctx, token)
	if err != nil || got == nil {
		t.Fatalf("session not resolvable: %v", err)
	}
	if got.ID != userID {
		t.Fatalf("session resolved to user %d, wanted %d", got.ID, userID)
	}

	// Signing out has to actually invalidate the token, not just forget it
	// client-side — a stolen cookie must stop working.
	if err := repo.DeleteSession(ctx, token); err != nil {
		t.Fatalf("delete session: %v", err)
	}
	if _, err := repo.GetUserBySession(ctx, token); err == nil {
		t.Error("a deleted session still resolves to a user")
	}
}

func TestDeleteSessionsSignsOutEverywhere(t *testing.T) {
	pool := billingPool(t)
	ctx := context.Background()
	repo := NewUserRepo(pool)
	userID := testUser(t, pool)

	var tokens []string
	for i := 0; i < 3; i++ {
		token, err := repo.CreateSession(ctx, userID)
		if err != nil {
			t.Fatalf("create session %d: %v", i, err)
		}
		tokens = append(tokens, token)
	}
	if err := repo.DeleteSessions(ctx, userID); err != nil {
		t.Fatalf("delete sessions: %v", err)
	}
	// Blocking a user or resetting their password is worthless if the devices they
	// are already signed in on keep working.
	for i, token := range tokens {
		if _, err := repo.GetUserBySession(ctx, token); err == nil {
			t.Errorf("session %d survived a sign-out-everywhere", i)
		}
	}
}

func TestSetAccessChangesRoleAndBlock(t *testing.T) {
	pool := billingPool(t)
	ctx := context.Background()
	repo := NewUserRepo(pool)
	userID := testUser(t, pool)

	if err := repo.SetAccess(ctx, userID, RoleAuthor, false); err != nil {
		t.Fatalf("promote to author: %v", err)
	}
	user, err := repo.GetByID(ctx, userID)
	if err != nil {
		t.Fatalf("read back: %v", err)
	}
	if user.Role != RoleAuthor {
		t.Fatalf("role is %q after promoting to author", user.Role)
	}
	if !user.CanAuthor() {
		t.Error("an author may not author")
	}
	if user.IsAdmin() {
		t.Error("an author counts as an admin")
	}

	if err := repo.SetAccess(ctx, userID, RoleStudent, true); err != nil {
		t.Fatalf("block: %v", err)
	}
	user, err = repo.GetByID(ctx, userID)
	if err != nil {
		t.Fatalf("read back: %v", err)
	}
	if !user.Blocked {
		t.Error("blocking did not take")
	}
	if user.CanAuthor() {
		t.Error("a demoted user kept authoring rights")
	}
}

func TestPasswordIsCheckedAgainstTheHash(t *testing.T) {
	pool := billingPool(t)
	ctx := context.Background()
	repo := NewUserRepo(pool)

	email := "pwd-test-" + time.Now().Format("150405.000000000") + "@example.test"
	user, err := repo.Create(ctx, email, "правильный-пароль", "Тест")
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	t.Cleanup(func() { pool.Exec(ctx, `DELETE FROM users WHERE id=$1`, user.ID) })

	if !repo.CheckPassword(user, "правильный-пароль") {
		t.Error("the correct password was rejected")
	}
	if repo.CheckPassword(user, "неправильный") {
		t.Error("a wrong password was accepted")
	}
	// The stored value must be a hash: a leaked dump should not be a password list.
	if user.PasswordHash == "правильный-пароль" {
		t.Fatal("the password is stored in clear text")
	}

	if err := repo.SetPassword(ctx, user.ID, "новый-пароль"); err != nil {
		t.Fatalf("set password: %v", err)
	}
	updated, err := repo.GetByEmail(ctx, email)
	if err != nil {
		t.Fatalf("read back: %v", err)
	}
	if !repo.CheckPassword(updated, "новый-пароль") {
		t.Error("the new password does not work")
	}
	if repo.CheckPassword(updated, "правильный-пароль") {
		t.Error("the old password still works after a change")
	}
}
