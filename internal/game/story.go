package game

import (
	"fmt"
)

// StoryManager manages story progression and content
type StoryManager struct {
	chapters []StoryChapter
	unlocked map[int]bool // chapter ID -> unlocked
}

// NewStoryManager creates a new story manager
func NewStoryManager() *StoryManager {
	sm := &StoryManager{
		chapters: getStoryChapters(),
		unlocked: make(map[int]bool),
	}
	// Unlock first chapter by default
	sm.unlocked[1] = true
	return sm
}

// getStoryChapters returns all story chapters
func getStoryChapters() []StoryChapter {
	return []StoryChapter{
		{
			ID:           1,
			Title:        "The Beginning",
			Content:      "In a digital jungle, a young monkey discovers a keyboard. Random keystrokes echo through the void as our hero begins their journey into the world of programming. Every keystroke is a step toward enlightenment.",
			TriggerLevel: 1,
			Unlocks:      []string{},
			IsRead:       false,
		},
		{
			ID:           2,
			Title:        "Pattern Recognition",
			Content:      "After countless random taps, something changes. The monkey starts seeing patterns in the chaos. Letters form recognizable shapes, and the concept of 'language' begins to emerge. This is the first step toward true understanding.",
			TriggerLevel: 5,
			Unlocks:      []string{"vocabulary_boost"},
			IsRead:       false,
		},
		{
			ID:           3,
			Title:        "First Words",
			Content:      "\"Hello\" appears on screen! The monkey gasps (in a metaphorical sense). Words have power, meaning, and structure. Each word formed is a victory against randomness, a step toward meaningful communication.",
			TriggerLevel: 10,
			Unlocks:      []string{"word_formation"},
			IsRead:       false,
		},
		{
			ID:           4,
			Title:        "The Loop of Understanding",
			Content:      "The monkey discovers loops - the ability to repeat actions efficiently. With each loop, knowledge compounds exponentially. This is the moment when our hero realizes that automation is the key to transcendence.",
			TriggerLevel: 15,
			Unlocks:      []string{"basic_programming"},
			IsRead:       false,
		},
		{
			ID:           5,
			Title:        "Function Enlightenment",
			Content:      "Functions! The monkey learns to package knowledge into reusable blocks of wisdom. Each function is a tool, a building block for greater creations. The journey from random typing to intentional programming accelerates.",
			TriggerLevel: 20,
			Unlocks:      []string{"advanced_programming"},
			IsRead:       false,
		},
		{
			ID:           6,
			Title:        "The Algorithm Age",
			Content:      "Our hero no longer just writes code - they create algorithms. Efficient, elegant solutions to complex problems. The monkey understands that programming isn't just about making things work, it's about making them work beautifully.",
			TriggerLevel: 30,
			Unlocks:      []string{"algorithm_mastery"},
			IsRead:       false,
		},
		{
			ID:           7,
			Title:        "Database Consciousness",
			Content:      "The monkey discovers the power of persistent knowledge. Data that survives beyond the current session, information that accumulates over time. This is digital immortality, the ability to build something that lasts.",
			TriggerLevel: 40,
			Unlocks:      []string{"data_mastery"},
			IsRead:       false,
		},
		{
			ID:           8,
			Title:        "Network Awakening",
			Content:      "Individual achievement is wonderful, but the monkey learns that true power comes from connection. Other programs, other minds, other monkeys working together. The network becomes a collective consciousness, a shared digital ecosystem.",
			TriggerLevel: 50,
			Unlocks:      []string{"network_programming"},
			IsRead:       false,
		},
		{
			ID:           9,
			Title:        "AI Integration",
			Content:      "The ultimate revelation: the monkey can create thinking machines! AI automations that learn, adapt, and grow. Our hero has transcended from random keystrokes to creating intelligence itself. The student becomes the master.",
			TriggerLevel: 60,
			Unlocks:      []string{"ai_automation"},
			IsRead:       false,
		},
		{
			ID:           10,
			Title:        "Digital Transcendence",
			Content:      "The monkey has become more than a monkey - they are a digital architect, a creator of worlds, a master of code. From random keystrokes to structured thought, from simple loops to AI creation. The journey is complete, but the evolution continues. What new worlds will you create?",
			TriggerLevel: 75,
			Unlocks:      []string{"transcendence"},
			IsRead:       false,
		},
	}
}

// CheckTriggers checks if any story chapters should be unlocked based on game state
func (sm *StoryManager) CheckTriggers(state *GameState) []StoryChapter {
	var newChapters []StoryChapter

	for _, chapter := range sm.chapters {
		// Skip if already unlocked
		if sm.unlocked[chapter.ID] {
			continue
		}

		// Check level-based triggers
		if state.CurrentLevel >= chapter.TriggerLevel {
			sm.unlocked[chapter.ID] = true
			newChapters = append(newChapters, chapter)
			continue
		}

		// Check upgrade-based triggers (some chapters require specific upgrades)
		if sm.checkUpgradeTriggers(state, chapter) {
			sm.unlocked[chapter.ID] = true
			newChapters = append(newChapters, chapter)
		}
	}

	return newChapters
}

// checkUpgradeTriggers checks if a chapter's unlock conditions are met via upgrades
func (sm *StoryManager) checkUpgradeTriggers(state *GameState, chapter StoryChapter) bool {
	if len(chapter.Unlocks) == 0 {
		return false
	}

	// For each required upgrade type, check if player has sufficient level
	for _, unlockType := range chapter.Unlocks {
		switch unlockType {
		case "vocabulary_boost":
			if state.Words >= 5 {
				return true
			}
		case "word_formation":
			if state.Words >= 20 {
				return true
			}
		case "basic_programming":
			if state.Programs >= 5 {
				return true
			}
		case "advanced_programming":
			if state.Programs >= 15 {
				return true
			}
		case "algorithm_mastery":
			if state.Programs >= 30 {
				return true
			}
		case "data_mastery":
			if state.Programs >= 50 && state.Words >= 100 {
				return true
			}
		case "network_programming":
			if state.AIAutomations >= 5 {
				return true
			}
		case "ai_automation":
			if state.AIAutomations >= 10 {
				return true
			}
		case "transcendence":
			if state.AIAutomations >= 25 && state.CurrentLevel >= 50 {
				return true
			}
		}
	}

	return false
}

// GetUnlockedChapters returns all unlocked story chapters
func (sm *StoryManager) GetUnlockedChapters() []StoryChapter {
	var unlocked []StoryChapter
	for _, chapter := range sm.chapters {
		if sm.unlocked[chapter.ID] {
			unlocked = append(unlocked, chapter)
		}
	}
	return unlocked
}

// GetCurrentChapter returns the most recent unread chapter
func (sm *StoryManager) GetCurrentChapter() *StoryChapter {
	unlocked := sm.GetUnlockedChapters()

	// Find first unread chapter
	for i := len(unlocked) - 1; i >= 0; i-- {
		if !unlocked[i].IsRead {
			return &unlocked[i]
		}
	}

	// If all read, return last unlocked
	if len(unlocked) > 0 {
		return &unlocked[len(unlocked)-1]
	}

	// Fallback to first chapter
	return &sm.chapters[0]
}

// MarkChapterRead marks a chapter as read
func (sm *StoryManager) MarkChapterRead(chapterID int) {
	// Update internal tracking
	for i, chapter := range sm.chapters {
		if chapter.ID == chapterID {
			sm.chapters[i].IsRead = true
			break
		}
	}
}

// GetProgress returns story progress (0-100%)
func (sm *StoryManager) GetProgress() float64 {
	totalChapters := len(sm.chapters)
	if totalChapters == 0 {
		return 0
	}

	unlockedCount := 0
	for _, chapter := range sm.chapters {
		if sm.unlocked[chapter.ID] {
			unlockedCount++
		}
	}

	return float64(unlockedCount) / float64(totalChapters) * 100
}

// GetNextChapter returns the next chapter that can be unlocked
func (sm *StoryManager) GetNextChapter(state *GameState) *StoryChapter {
	for _, chapter := range sm.chapters {
		if sm.unlocked[chapter.ID] {
			continue
		}

		// Check if this chapter can be unlocked
		if state.CurrentLevel >= chapter.TriggerLevel || sm.checkUpgradeTriggers(state, chapter) {
			return &chapter
		}
	}

	return nil
}

// GetHint returns a hint about what to do to unlock the next chapter
func (sm *StoryManager) GetHint(state *GameState) string {
	nextChapter := sm.GetNextChapter(state)
	if nextChapter == nil {
		return "You've unlocked all chapters! The journey is complete."
	}

	if state.CurrentLevel < nextChapter.TriggerLevel {
		levelsNeeded := nextChapter.TriggerLevel - state.CurrentLevel
		return fmt.Sprintf("Reach level %d to unlock '%s' (%d levels to go)",
			nextChapter.TriggerLevel, nextChapter.Title, levelsNeeded)
	}

	// Check specific requirements based on unlocks
	for _, unlock := range nextChapter.Unlocks {
		switch unlock {
		case "vocabulary_boost":
			if state.Words < 5 {
				needed := 5 - state.Words
				return fmt.Sprintf("Form %d more words to unlock '%s'", needed, nextChapter.Title)
			}
		case "word_formation":
			if state.Words < 20 {
				needed := 20 - state.Words
				return fmt.Sprintf("Form %d more words to unlock '%s'", needed, nextChapter.Title)
			}
		case "basic_programming":
			if state.Programs < 5 {
				needed := 5 - state.Programs
				return fmt.Sprintf("Create %d more programs to unlock '%s'", needed, nextChapter.Title)
			}
		case "advanced_programming":
			if state.Programs < 15 {
				needed := 15 - state.Programs
				return fmt.Sprintf("Create %d more programs to unlock '%s'", needed, nextChapter.Title)
			}
		case "algorithm_mastery":
			if state.Programs < 30 {
				needed := 30 - state.Programs
				return fmt.Sprintf("Create %d more programs to unlock '%s'", needed, nextChapter.Title)
			}
		case "data_mastery":
			if state.Programs < 50 || state.Words < 100 {
				progNeeded := 50 - state.Programs
				wordNeeded := 100 - state.Words
				return fmt.Sprintf("Create %d more programs and form %d more words to unlock '%s'",
					progNeeded, wordNeeded, nextChapter.Title)
			}
		case "network_programming":
			if state.AIAutomations < 5 {
				needed := 5 - state.AIAutomations
				return fmt.Sprintf("Build %d more AI automations to unlock '%s'", needed, nextChapter.Title)
			}
		case "ai_automation":
			if state.AIAutomations < 10 {
				needed := 10 - state.AIAutomations
				return fmt.Sprintf("Build %d more AI automations to unlock '%s'", needed, nextChapter.Title)
			}
		case "transcendence":
			if state.AIAutomations < 25 || state.CurrentLevel < 50 {
				aiNeeded := 25 - state.AIAutomations
				levelNeeded := 50 - state.CurrentLevel
				return fmt.Sprintf("Build %d more AI automations and reach level %d to unlock '%s'",
					aiNeeded, levelNeeded, nextChapter.Title)
			}
		}
	}

	return fmt.Sprintf("Continue progressing to unlock '%s'", nextChapter.Title)
}
