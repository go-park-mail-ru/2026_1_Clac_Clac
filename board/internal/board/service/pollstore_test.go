package service

import (
	"sync"
	"testing"

	"github.com/go-park-mail-ru/2026_1_Clac_Clac/board/internal/board/common"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestPollStore_Create(t *testing.T) {
	boardLink := uuid.New()
	adminLink := uuid.New()
	invitees := []uuid.UUID{uuid.New(), uuid.New()}

	t.Run("Success", func(t *testing.T) {
		ps := NewPollStore()
		err := ps.Create(boardLink, adminLink, []CardInfo{}, invitees)
		assert.NoError(t, err)

		poll, ok := ps.GetActivePoll(boardLink)
		assert.True(t, ok)
		assert.Equal(t, adminLink, poll.AdminLink)
		assert.Equal(t, 0, poll.CurrentIdx)
		assert.Equal(t, invitees, poll.Invitees)
	})

	t.Run("Error_AlreadyExists", func(t *testing.T) {
		ps := NewPollStore()
		_ = ps.Create(boardLink, adminLink, []CardInfo{}, invitees)
		err := ps.Create(boardLink, uuid.New(), []CardInfo{}, invitees)
		assert.ErrorIs(t, err, common.ErrPollAlreadyExists)
	})
}

func TestPollStore_Delete(t *testing.T) {
	boardLink := uuid.New()
	adminLink := uuid.New()
	otherLink := uuid.New()

	t.Run("Success", func(t *testing.T) {
		ps := NewPollStore()
		_ = ps.Create(boardLink, adminLink, []CardInfo{}, nil)
		err := ps.Delete(boardLink, adminLink)
		assert.NoError(t, err)

		_, ok := ps.GetActivePoll(boardLink)
		assert.False(t, ok)
	})

	t.Run("Error_NotFound", func(t *testing.T) {
		ps := NewPollStore()
		err := ps.Delete(boardLink, adminLink)
		assert.ErrorIs(t, err, common.ErrPollNotFound)
	})

	t.Run("Error_NotAdmin", func(t *testing.T) {
		ps := NewPollStore()
		_ = ps.Create(boardLink, adminLink, []CardInfo{}, nil)
		err := ps.Delete(boardLink, otherLink)
		assert.ErrorIs(t, err, common.ErrNotPollAdmin)
	})
}

func TestPollStore_NextCard(t *testing.T) {
	boardLink := uuid.New()
	adminLink := uuid.New()
	otherLink := uuid.New()
	card1 := uuid.New()
	card2 := uuid.New()

	t.Run("Success_Advance", func(t *testing.T) {
		ps := NewPollStore()
		_ = ps.Create(boardLink, adminLink, []CardInfo{{Link: card1}, {Link: card2}}, nil)
		poll, err := ps.NextCard(boardLink, adminLink)
		assert.NoError(t, err)
		assert.Equal(t, 1, poll.CurrentIdx)

		_, ok := ps.GetActivePoll(boardLink)
		assert.True(t, ok)
	})

	t.Run("Success_LastCardDeletesPoll", func(t *testing.T) {
		ps := NewPollStore()
		_ = ps.Create(boardLink, adminLink, []CardInfo{{Link: card1}}, nil)
		_, err := ps.NextCard(boardLink, adminLink)
		assert.ErrorIs(t, err, common.ErrPollNoMoreCards)

		_, ok := ps.GetActivePoll(boardLink)
		assert.False(t, ok)
	})

	t.Run("Error_NotFound", func(t *testing.T) {
		ps := NewPollStore()
		_, err := ps.NextCard(boardLink, adminLink)
		assert.ErrorIs(t, err, common.ErrPollNotFound)
	})

	t.Run("Error_NotAdmin", func(t *testing.T) {
		ps := NewPollStore()
		_ = ps.Create(boardLink, adminLink, []CardInfo{{Link: card1}, {Link: card2}}, nil)
		_, err := ps.NextCard(boardLink, otherLink)
		assert.ErrorIs(t, err, common.ErrNotPollAdmin)
	})
}

func TestPollStore_Vote(t *testing.T) {
	boardLink := uuid.New()
	adminLink := uuid.New()
	invitedUser := uuid.New()
	nonInvitedUser := uuid.New()
	card := uuid.New()

	t.Run("Success", func(t *testing.T) {
		ps := NewPollStore()
		_ = ps.Create(boardLink, adminLink, []CardInfo{{Link: card}}, []uuid.UUID{invitedUser})
		err := ps.Vote(boardLink, invitedUser, 5)
		assert.NoError(t, err)

		poll, ok := ps.GetActivePoll(boardLink)
		assert.True(t, ok)
		assert.Equal(t, 5, *poll.Tasks[0].Votes[invitedUser])
	})

	t.Run("Success_Revote", func(t *testing.T) {
		ps := NewPollStore()
		_ = ps.Create(boardLink, adminLink, []CardInfo{{Link: card}}, []uuid.UUID{invitedUser})
		_ = ps.Vote(boardLink, invitedUser, 5)
		err := ps.Vote(boardLink, invitedUser, 8)
		assert.NoError(t, err)

		poll, _ := ps.GetActivePoll(boardLink)
		assert.Equal(t, 8, *poll.Tasks[0].Votes[invitedUser])
	})

	t.Run("Error_NotFound", func(t *testing.T) {
		ps := NewPollStore()
		err := ps.Vote(boardLink, invitedUser, 5)
		assert.ErrorIs(t, err, common.ErrPollNotFound)
	})

	t.Run("Error_NotInvited", func(t *testing.T) {
		ps := NewPollStore()
		_ = ps.Create(boardLink, adminLink, []CardInfo{{Link: card}}, []uuid.UUID{invitedUser})
		err := ps.Vote(boardLink, nonInvitedUser, 5)
		assert.ErrorIs(t, err, common.ErrUserNotInvited)
	})
}

func TestPollStore_GetActivePoll(t *testing.T) {
	boardLink := uuid.New()
	adminLink := uuid.New()

	t.Run("Exists", func(t *testing.T) {
		ps := NewPollStore()
		_ = ps.Create(boardLink, adminLink, []CardInfo{}, nil)
		poll, ok := ps.GetActivePoll(boardLink)
		assert.True(t, ok)
		assert.NotNil(t, poll)
	})

	t.Run("NotExists", func(t *testing.T) {
		ps := NewPollStore()
		poll, ok := ps.GetActivePoll(uuid.New())
		assert.False(t, ok)
		assert.Nil(t, poll)
	})
}

func TestPollStore_IsPollAdmin(t *testing.T) {
	boardLink := uuid.New()
	adminLink := uuid.New()
	otherLink := uuid.New()

	ps := NewPollStore()
	_ = ps.Create(boardLink, adminLink, []CardInfo{}, nil)

	assert.True(t, ps.IsPollAdmin(boardLink, adminLink))
	assert.False(t, ps.IsPollAdmin(boardLink, otherLink))
	assert.False(t, ps.IsPollAdmin(uuid.New(), adminLink))
}

func TestPollStore_ConcurrentVote(t *testing.T) {
	boardLink := uuid.New()
	adminLink := uuid.New()
	card := uuid.New()
	invitees := make([]uuid.UUID, 10)
	for i := range invitees {
		invitees[i] = uuid.New()
	}

	ps := NewPollStore()
	_ = ps.Create(boardLink, adminLink, []CardInfo{{Link: card}}, invitees)

	var wg sync.WaitGroup
	for _, user := range invitees {
		wg.Add(1)
		go func(uid uuid.UUID) {
			defer wg.Done()
			_ = ps.Vote(boardLink, uid, 3)
		}(user)
	}
	wg.Wait()

	poll, ok := ps.GetActivePoll(boardLink)
	assert.True(t, ok)
	assert.Equal(t, len(invitees), len(poll.Tasks[0].Votes))
}
