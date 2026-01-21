package game

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewStoryManager(t *testing.T) {
	sm := NewStoryManager()

	assert.NotNil(t, sm)
	assert.NotNil(t, sm.chapters)
	assert.NotNil(t, sm.unlocked)
	assert.True(t, len(sm.chapters) > 0, "Should have story chapters")
	assert.True(t, sm.unlocked[1], "First chapter should be unlocked by default")
}

func TestStoryChapters(t *testing.T) {
	chapters := getStoryChapters()

	assert.Equal(t, 10, len(chapters), "Should have 10 story chapters")

	// Check first chapter
	firstChapter := chapters[0]
	assert.Equal(t, 1, firstChapter.ID)
	assert.Equal(t, "The Beginning", firstChapter.Title)
	assert.Equal(t, 1, firstChapter.TriggerLevel)
	assert.Contains(t, firstChapter.Content, "keyboard")

	// Check last chapter
	lastChapter := chapters[len(chapters)-1]
	assert.Equal(t, 10, lastChapter.ID)
	assert.Equal(t, "Digital Transcendence", lastChapter.Title)
	assert.Equal(t, 75, lastChapter.TriggerLevel)
	assert.Contains(t, lastChapter.Content, "journey")
}

func TestCheckTriggers_LevelBased(t *testing.T) {
	sm := NewStoryManager()
	state := NewGameState("test")

	// Initially only chapter 1 should be unlocked
	assert.Equal(t, 1, len(sm.GetUnlockedChapters()))

	// Level up to trigger chapter 2
	state.CurrentLevel = 5
	newChapters := sm.CheckTriggers(state)
	assert.Equal(t, 1, len(newChapters), "Should unlock one new chapter")
	assert.Equal(t, 2, newChapters[0].ID)
	assert.Equal(t, 2, len(sm.GetUnlockedChapters()))

	// Level up to trigger chapter 3
	state.CurrentLevel = 10
	newChapters = sm.CheckTriggers(state)
	assert.Equal(t, 1, len(newChapters), "Should unlock one new chapter")
	assert.Equal(t, 3, newChapters[0].ID)
	assert.Equal(t, 3, len(sm.GetUnlockedChapters()))
}

func TestCheckTriggers_UpgradeBased(t *testing.T) {
	sm := NewStoryManager()
	state := NewGameState("test")

	// Trigger some chapters by resources instead of just level
	state.Words = 25         // Should trigger word formation chapter
	state.Programs = 20      // Should trigger programming chapters
	state.AIAutomations = 15 // Should trigger AI chapters

	newChapters := sm.CheckTriggers(state)
	assert.True(t, len(newChapters) >= 3, "Should unlock multiple resource-based chapters")
}

func TestGetUnlockedChapters(t *testing.T) {
	sm := NewStoryManager()

	// Initially only first chapter
	unlocked := sm.GetUnlockedChapters()
	assert.Equal(t, 1, len(unlocked))
	assert.Equal(t, 1, unlocked[0].ID)

	// Unlock more chapters
	sm.unlocked[2] = true
	sm.unlocked[3] = true

	unlocked = sm.GetUnlockedChapters()
	assert.Equal(t, 3, len(unlocked))
	assert.Equal(t, 1, unlocked[0].ID)
	assert.Equal(t, 2, unlocked[1].ID)
	assert.Equal(t, 3, unlocked[2].ID)
}

func TestGetCurrentChapter(t *testing.T) {
	sm := NewStoryManager()

	// Should return first chapter initially
	current := sm.GetCurrentChapter()
	assert.NotNil(t, current)
	assert.Equal(t, 1, current.ID)

	// Mark as read and check
	sm.MarkChapterRead(1)
	current = sm.GetCurrentChapter()
	assert.True(t, current.IsRead)

	// Unlock more chapters
	sm.unlocked[2] = true
	current = sm.GetCurrentChapter()
	assert.Equal(t, 2, current.ID) // Should return first unread chapter
}

func TestMarkChapterRead(t *testing.T) {
	sm := NewStoryManager()

	// Initially first chapter is unread
	chapter := sm.chapters[0]
	assert.False(t, chapter.IsRead)

	// Mark as read
	sm.MarkChapterRead(1)
	assert.True(t, sm.chapters[0].IsRead)
}

func TestGetProgress(t *testing.T) {
	sm := NewStoryManager()

	// Initially 10% (1 out of 10 chapters)
	progress := sm.GetProgress()
	assert.Equal(t, 10.0, progress)

	// Unlock more chapters
	sm.unlocked[2] = true
	sm.unlocked[3] = true

	progress = sm.GetProgress()
	assert.Equal(t, 30.0, progress)

	// All chapters unlocked
	for i := 1; i <= 10; i++ {
		sm.unlocked[i] = true
	}

	progress = sm.GetProgress()
	assert.Equal(t, 100.0, progress)
}

func TestGetNextChapter(t *testing.T) {
	sm := NewStoryManager()
	state := NewGameState("test")
	state.StoryManager = sm

	// Initially at level 1, no chapter can be unlocked (chapter 2 needs level 5)
	next := sm.GetNextChapter(state)
	assert.Nil(t, next, "No chapter should be unlockable at level 1")

	// Level up to unlock chapter 2
	state.CurrentLevel = 5
	newChapters := sm.CheckTriggers(state)
	assert.Equal(t, 1, len(newChapters), "Should unlock chapter 2")

	// Next should be chapter 3 (level 10) since chapter 2 is now unlocked
	next = sm.GetNextChapter(state)
	if next != nil {
		assert.Equal(t, 3, next.ID)
		assert.Equal(t, 10, next.TriggerLevel)
	}
}

func TestGetHint(t *testing.T) {
	sm := NewStoryManager()
	state := NewGameState("test")
	state.StoryManager = sm

	// Initially need to reach level 5
	hint := sm.GetHint(state)
	if hint != "You've unlocked all chapters! The journey is complete." {
		assert.Contains(t, hint, "level 5")
	}

	// Level up
	state.CurrentLevel = 5
	sm.CheckTriggers(state)

	// Should now hint about next unlockable chapter
	hint = sm.GetHint(state)
	if hint != "You've unlocked all chapters! The journey is complete." {
		t.Logf("Hint: %s", hint)
	}

	// Test resource-based hints by setting high level to enable resource triggers
	state.CurrentLevel = 15 // High enough for all level-based triggers
	state.Words = 3
	hint = sm.GetHint(state)
	if hint != "You've unlocked all chapters! The journey is complete." {
		t.Logf("Resource hint: %s", hint)
	}
}

func TestGameStateStoryIntegration(t *testing.T) {
	state := NewGameState("test")

	assert.NotNil(t, state.StoryManager, "Story manager should be initialized")

	// Test initial state
	unlocked := state.StoryManager.GetUnlockedChapters()
	assert.Equal(t, 1, len(unlocked))

	// Test trigger checking through game state
	state.CurrentLevel = 10
	newChapters := state.CheckStoryTriggers()
	assert.True(t, len(newChapters) >= 2, "Should unlock chapters based on level")

	// Test that notifications are added
	assert.True(t, len(state.Notifications) > 0)
	assert.Contains(t, state.Notifications[len(state.Notifications)-1], "New Story Chapter")
}

func TestStoryResourceTriggers(t *testing.T) {
	tests := []struct {
		name             string
		setupState       func(*GameState)
		expectedUnlocked []int
	}{
		{
			name: "Word formation trigger",
			setupState: func(state *GameState) {
				state.Words = 25
			},
			expectedUnlocked: []int{3}, // Should unlock word formation chapter
		},
		{
			name: "Programming trigger",
			setupState: func(state *GameState) {
				state.Programs = 20     // Need 15+ for advanced programming
				state.CurrentLevel = 15 // Need to be high enough level
			},
			expectedUnlocked: []int{4, 5}, // Should unlock programming chapters
		},
		{
			name: "AI automation trigger",
			setupState: func(state *GameState) {
				state.AIAutomations = 12
			},
			expectedUnlocked: []int{9}, // Should unlock AI chapter
		},
		{
			name: "Transcendence trigger",
			setupState: func(state *GameState) {
				state.CurrentLevel = 60
				state.AIAutomations = 30
			},
			expectedUnlocked: []int{10}, // Should unlock final chapter
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sm := NewStoryManager()
			state := NewGameState("test")
			state.StoryManager = sm

			tt.setupState(state)
			newChapters := sm.CheckTriggers(state)

			assert.GreaterOrEqual(t, len(newChapters), len(tt.expectedUnlocked))

			for _, expectedID := range tt.expectedUnlocked {
				found := false
				for _, chapter := range newChapters {
					if chapter.ID == expectedID {
						found = true
						break
					}
				}
				assert.True(t, found, "Should unlock chapter %d", expectedID)
			}
		})
	}
}

func TestStoryChapterContent(t *testing.T) {
	chapters := getStoryChapters()

	// Test that all chapters have meaningful content
	for _, chapter := range chapters {
		assert.NotEmpty(t, chapter.Title, "Chapter %d should have a title", chapter.ID)
		assert.NotEmpty(t, chapter.Content, "Chapter %d should have content", chapter.ID)
		assert.Greater(t, chapter.TriggerLevel, 0, "Chapter %d should have a trigger level", chapter.ID)
		assert.GreaterOrEqual(t, chapter.TriggerLevel, chapter.ID-1, "Chapter %d trigger should be reasonable", chapter.ID)

		// Check for key story themes
		if chapter.ID <= 3 {
			assert.Contains(t, chapter.Content, "monkey", "Early chapters should mention monkey")
		}
		if chapter.ID >= 9 { // Only last chapters definitely mention AI
			assert.Contains(t, chapter.Content, "AI", "Later chapters should mention AI")
		}
	}
}
