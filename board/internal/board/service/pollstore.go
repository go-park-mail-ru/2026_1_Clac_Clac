package service

import (
	"slices"
	"sync"

	"github.com/go-park-mail-ru/2026_1_Clac_Clac/board/internal/board/common"
	"github.com/google/uuid"
)

type PollStore struct {
	mu    sync.RWMutex
	polls map[string]*Poll
}

type Poll struct {
	BoardLink  uuid.UUID
	AdminLink  uuid.UUID
	Tasks      []PollTask
	CurrentIdx int
	Invitees   []uuid.UUID
}

type PollTask struct {
	CardLink uuid.UUID
	Title    string
	Votes    map[uuid.UUID]*int
}

func NewPollStore() *PollStore {
	return &PollStore{
		polls: make(map[string]*Poll),
	}
}

func (ps *PollStore) Create(boardLink, adminLink uuid.UUID, cards []uuid.UUID, invitees []uuid.UUID) error {
	ps.mu.Lock()
	defer ps.mu.Unlock()

	key := boardLink.String()
	if _, ok := ps.polls[key]; ok {
		return common.ErrPollAlreadyExists
	}

	tasks := make([]PollTask, len(cards))
	for i, c := range cards {
		tasks[i] = PollTask{
			CardLink: c,
			Title:    "",
			Votes:    make(map[uuid.UUID]*int),
		}
	}

	ps.polls[key] = &Poll{
		BoardLink:  boardLink,
		AdminLink:  adminLink,
		Tasks:      tasks,
		CurrentIdx: 0,
		Invitees:   invitees,
	}

	return nil
}

func (ps *PollStore) Delete(boardLink, userLink uuid.UUID) error {
	ps.mu.Lock()
	defer ps.mu.Unlock()

	key := boardLink.String()
	poll, ok := ps.polls[key]
	if !ok {
		return common.ErrPollNotFound
	}

	if poll.AdminLink != userLink {
		return common.ErrNotPollAdmin
	}

	delete(ps.polls, key)
	return nil
}

func (ps *PollStore) NextCard(boardLink, userLink uuid.UUID) (*Poll, error) {
	ps.mu.Lock()
	defer ps.mu.Unlock()

	key := boardLink.String()
	poll, ok := ps.polls[key]
	if !ok {
		return nil, common.ErrPollNotFound
	}

	if poll.AdminLink != userLink {
		return nil, common.ErrNotPollAdmin
	}

	poll.CurrentIdx++

	if poll.CurrentIdx >= len(poll.Tasks) {
		delete(ps.polls, key)
		return nil, common.ErrPollNoMoreCards
	}

	return poll, nil
}

func (ps *PollStore) Vote(boardLink, userLink uuid.UUID, points int) error {
	ps.mu.Lock()
	defer ps.mu.Unlock()

	key := boardLink.String()
	poll, ok := ps.polls[key]
	if !ok {
		return common.ErrPollNotFound
	}

	if !slices.Contains(poll.Invitees, userLink) {
		return common.ErrUserNotInvited
	}

	poll.Tasks[poll.CurrentIdx].Votes[userLink] = &points
	return nil
}

func (ps *PollStore) GetActivePoll(boardLink uuid.UUID) (*Poll, bool) {
	ps.mu.RLock()
	defer ps.mu.RUnlock()

	poll, ok := ps.polls[boardLink.String()]
	return poll, ok
}

func (ps *PollStore) IsPollAdmin(boardLink, userLink uuid.UUID) bool {
	ps.mu.RLock()
	defer ps.mu.RUnlock()

	poll, ok := ps.polls[boardLink.String()]
	if !ok {
		return false
	}

	return poll.AdminLink == userLink
}
