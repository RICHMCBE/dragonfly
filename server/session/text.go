package session

import (
	"time"

	"github.com/df-mc/dragonfly/server/player/chat"
	"github.com/df-mc/dragonfly/server/player/scoreboard"
	"github.com/sandertv/gophertunnel/minecraft/protocol"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
	"golang.org/x/text/language"
)

// SendMessage ...
func (s *Session) SendMessage(message string) {
	s.writePacket(&packet.Text{
		TextType: packet.TextTypeRaw,
		Message:  message,
	})
}

// SendTranslation sends a translation localised for a specific language.Tag.
func (s *Session) SendTranslation(t chat.Translation, l language.Tag, a []any) {
	tr := t.F(a...)
	s.writePacket(&packet.Text{
		TextType:         packet.TextTypeTranslation,
		NeedsTranslation: true,
		Message:          tr.Resolve(l),
		Parameters:       tr.Params(l),
	})
}

// SendTip ...
func (s *Session) SendTip(message string) {
	s.writePacket(&packet.Text{
		TextType: packet.TextTypeTip,
		Message:  message,
	})
}

// SendAnnouncement ...
func (s *Session) SendAnnouncement(message string) {
	s.writePacket(&packet.Text{
		TextType: packet.TextTypeAnnouncement,
		Message:  message,
	})
}

// SendPopup ...
func (s *Session) SendPopup(message string) {
	s.writePacket(&packet.Text{
		TextType: packet.TextTypePopup,
		Message:  message,
	})
}

// SendJukeboxPopup ...
func (s *Session) SendJukeboxPopup(message string) {
	s.writePacket(&packet.Text{
		TextType: packet.TextTypeJukeboxPopup,
		Message:  message,
	})
}

// SendToast ...
func (s *Session) SendToast(title, message string) {
	s.writePacket(&packet.ToastRequest{
		Title:   title,
		Message: message,
	})
}

// SendScoreboard ...
func (s *Session) SendScoreboard(sb *scoreboard.Scoreboard) {
	if s == Nop || s.conf.DisableScoreboard {
		return
	}
	currentName, currentLines := *s.currentScoreboard.Load(), *s.currentLines.Load()

	if currentName != sb.ObjectiveName() || s.scoreboardDisplayName != sb.Name() {
		s.removeScoreboardObjective()
		pk := &packet.SetDisplayObjective{
			DisplaySlot:   "sidebar",
			ObjectiveName: sb.ObjectiveName(),
			DisplayName:   sb.Name(),
			CriteriaName:  "dummy",
		}
		if sb.Descending() {
			pk.SortOrder = packet.ScoreboardSortOrderDescending
		} else {
			pk.SortOrder = packet.ScoreboardSortOrderAscending
		}
		s.writePacket(pk)
		name, lines := sb.ObjectiveName(), append([]string(nil), sb.Lines()...)
		s.scoreboardDisplayName = sb.Name()
		s.currentScoreboard.Store(&name)
		s.currentLines.Store(&lines)
	} else {
		// Remove all current lines from the scoreboard. We can't replace them without removing them.
		pk := &packet.SetScore{}
		for i := range currentLines {
			pk.Entries = append(pk.Entries, protocol.ScoreboardEntry{
				IdentityType:  protocol.ScoreboardIdentityRemove,
				EntryID:       int64(i),
				ObjectiveName: currentName,
				Score:         int32(i),
			})
		}
		if len(pk.Entries) > 0 {
			s.writePacket(pk)
		}
	}
	pk := &packet.SetScore{}
	for k, line := range sb.Lines() {
		if len(line) == 0 {
			line = "§" + colours[k]
		}
		pk.Entries = append(pk.Entries, protocol.ScoreboardEntry{
			EntryID:       int64(k),
			ObjectiveName: sb.ObjectiveName(),
			Score:         int32(k),
			IdentityType:  protocol.ScoreboardIdentityFakePlayer,
			DisplayName:   line,
		})
	}
	if len(pk.Entries) > 0 {
		s.writePacket(pk)
	}
	// Keep removal state in sync even when the objective title is unchanged.
	// InfoPlugin may hide or split rows as tags change, independently of HP.
	lines := append([]string(nil), sb.Lines()...)
	s.currentLines.Store(&lines)
	s.sendHealthHUDRows()
}

// colours holds a list of colour codes to be filled out for empty lines in a scoreboard.
var colours = [15]string{"1", "2", "3", "4", "5", "6", "7", "8", "9", "a", "b", "c", "d", "e", "f"}

// RemoveScoreboard ...
func (s *Session) removeScoreboardObjective() {
	if name := *s.currentScoreboard.Load(); name != "" {
		s.writePacket(&packet.RemoveObjective{ObjectiveName: name})
	}
	s.healthHUDSent, s.infoHUDSent = false, false
	s.scoreboardDisplayName = ""
	var name string
	var lines []string
	s.currentScoreboard.Store(&name)
	s.currentLines.Store(&lines)
}

// SendBossBar sends a boss bar to the player with the text passed and the health percentage of the bar.
// SendBossBar removes any boss bar that might be active before sending the new one.
func (s *Session) SendBossBar(text string, colour uint8, healthPercentage float64) {
	s.RemoveBossBar()
	s.writePacket(&packet.BossEvent{
		BossEntityUniqueID: selfEntityRuntimeID,
		EventType:          packet.BossEventShow,
		BossBarTitle:       text,
		HealthPercentage:   float32(healthPercentage),
		Colour:             colour,
	})
}

// RemoveBossBar removes any boss bar currently active on the player's screen.
func (s *Session) RemoveBossBar() {
	s.writePacket(&packet.BossEvent{
		BossEntityUniqueID: selfEntityRuntimeID,
		EventType:          packet.BossEventHide,
	})
}

const tickLength = time.Second / 20

// SetTitleDurations ...
func (s *Session) SetTitleDurations(fadeInDuration, remainDuration, fadeOutDuration time.Duration) {
	s.writePacket(&packet.SetTitle{
		ActionType:      packet.TitleActionSetDurations,
		FadeInDuration:  int32(fadeInDuration / tickLength),
		RemainDuration:  int32(remainDuration / tickLength),
		FadeOutDuration: int32(fadeOutDuration / tickLength),
	})
}

// SendTitle ...
func (s *Session) SendTitle(text string) {
	s.writePacket(&packet.SetTitle{ActionType: packet.TitleActionSetTitle, Text: text})
}

// SendSubtitle ...
func (s *Session) SendSubtitle(text string) {
	s.writePacket(&packet.SetTitle{ActionType: packet.TitleActionSetSubtitle, Text: text})
}

// SendActionBarMessage ...
func (s *Session) SendActionBarMessage(text string) {
	s.writePacket(&packet.SetTitle{ActionType: packet.TitleActionSetActionBar, Text: text})
}

// RemoveScoreboard removes ordinary rows while retaining the health HUD.
func (s *Session) RemoveScoreboard() {
	if s == Nop || s.conf.DisableScoreboard {
		return
	}
	s.removeScoreboardObjective()
	s.ensureHealthHUDObjective()
	s.sendHealthHUDRows()
}

// SendHealthHUD reserves two sidebar rows (and SendInfoHUD a third). Normal sidebar users should
// keep at most 12 rows so that the combined display fits Bedrock's 15-row sidebar.
// Updates use positive IDs outside the ordinary 0..14 row range.
func (s *Session) SendHealthHUD(bar, text string) {
	if s == Nop || s.conf.DisableScoreboard {
		return
	}
	s.healthHUDMu.Lock()
	lines := [3]string{bar, text}
	if old := s.healthHUDLines.Load(); old != nil {
		lines[2] = old[2]
	}
	s.healthHUDLines.Store(&lines)
	s.healthHUDMu.Unlock()
	s.ensureHealthHUDObjective()
	s.sendHealthHUDRows()
}

// SendInfoHUD sets a third reserved sidebar row next to the health HUD, for a resource pack HUD
// element that reads it by a prefix of its own (such as the server season). An empty text
// removes the row.
func (s *Session) SendInfoHUD(text string) {
	if s == Nop || s.conf.DisableScoreboard {
		return
	}
	s.healthHUDMu.Lock()
	var lines [3]string
	if old := s.healthHUDLines.Load(); old != nil {
		lines = *old
	}
	lines[2] = text
	s.healthHUDLines.Store(&lines)
	s.healthHUDMu.Unlock()
	s.ensureHealthHUDObjective()
	s.sendHealthHUDRows()
}

func (s *Session) ensureHealthHUDObjective() {
	if s.healthHUDLines.Load() == nil || *s.currentScoreboard.Load() != "" {
		return
	}
	name := "InfoPlugin"
	s.writePacket(&packet.SetDisplayObjective{DisplaySlot: "sidebar", ObjectiveName: name, DisplayName: "§r", CriteriaName: "dummy"})
	s.scoreboardDisplayName = "§r"
	s.currentScoreboard.Store(&name)
}

func (s *Session) sendHealthHUDRows() {
	lines := s.healthHUDLines.Load()
	if lines == nil {
		return
	}
	name := *s.currentScoreboard.Load()
	if name == "" {
		return
	}
	entries := make([]protocol.ScoreboardEntry, 0, 3)
	for i, line := range lines {
		// The health rows are only sent once the health HUD was set, the info row while it has
		// text.
		if line == "" && (i == 2 || lines[0] == "") {
			continue
		}
		// Scores only determine row ordering; HP is carried in the name payload.
		score := int32(14 + i)
		entries = append(entries, protocol.ScoreboardEntry{EntryID: int64(1001 + i), ObjectiveName: name, Score: score, IdentityType: protocol.ScoreboardIdentityFakePlayer, DisplayName: line})
	}
	// Removing before adding allows clients to replace changed fake-player names.
	remove := &packet.SetScore{}
	for i := range lines {
		if i < 2 && s.healthHUDSent || i == 2 && s.infoHUDSent {
			remove.Entries = append(remove.Entries, protocol.ScoreboardEntry{EntryID: int64(1001 + i), ObjectiveName: name, Score: int32(14 + i), IdentityType: protocol.ScoreboardIdentityRemove})
		}
	}
	if len(remove.Entries) > 0 {
		s.writePacket(remove)
	}
	if len(entries) > 0 {
		s.writePacket(&packet.SetScore{Entries: entries})
	}
	s.healthHUDSent = lines[0] != ""
	s.infoHUDSent = lines[2] != ""
}
