package session

import (
	"fmt"
	"github.com/df-mc/dragonfly/server/player/scoreboard"
	"github.com/sandertv/gophertunnel/minecraft/protocol"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
	"testing"
)

func TestHealthGaugeScoresOnlyOrderRows(t *testing.T) {
	s := &Session{packets: make(chan packet.Packet, 100), closeBackground: make(chan struct{})}
	name := ""
	lines := []string{}
	s.currentScoreboard.Store(&name)
	s.currentLines.Store(&lines)
	for _, percent := range []int{100, 75, 25, 0, 50, 100} {
		s.SendHealthHUD(fmt.Sprintf("§C§S§H§B%d", percent), "hp")
		found := false
		for len(s.packets) > 0 {
			if scores, ok := (<-s.packets).(*packet.SetScore); ok {
				for _, e := range scores.Entries {
					if e.EntryID == 1001 && e.IdentityType == protocol.ScoreboardIdentityFakePlayer {
						if e.Score != 14 {
							t.Fatalf("gauge score=%d want fixed ordering score 14", e.Score)
						}
						found = true
					}
				}
			}
		}
		if !found {
			t.Fatal("gauge row missing")
		}
	}
}

func TestDisabledScoreboardSendsNoPackets(t *testing.T) {
	s := &Session{conf: Config{DisableScoreboard: true}, packets: make(chan packet.Packet, 10), closeBackground: make(chan struct{})}
	s.SendHealthHUD("75", "75/100")
	s.SendScoreboard(scoreboard.New("status"))
	s.RemoveScoreboard()
	if len(s.packets) != 0 {
		t.Fatal("disabled scoreboard sent packets")
	}
}

func TestInfoObjectiveAndInitialHealthLifecycle(t *testing.T) {
	s := &Session{packets: make(chan packet.Packet, 100), closeBackground: make(chan struct{})}
	name := ""
	lines := []string{}
	s.currentScoreboard.Store(&name)
	s.currentLines.Store(&lines)
	s.SendHealthHUD("bar", "hp")
	for len(s.packets) > 0 {
		switch pk := (<-s.packets).(type) {
		case *packet.RemoveObjective:
			t.Fatal("removed an objective before creation")
		case *packet.SetScore:
			for _, e := range pk.Entries {
				if e.IdentityType == protocol.ScoreboardIdentityRemove || e.EntryID <= 0 {
					t.Fatal("initial health must only create positive IDs")
				}
			}
		}
	}
	for _, title := range []string{"리치팜", "새 제목"} {
		b := scoreboard.New(title)
		b.SetObjectiveName("InfoPlugin")
		b.WriteString("info")
		s.SendScoreboard(b)
		if *s.currentScoreboard.Load() != "InfoPlugin" {
			t.Fatal("title changed objective ID")
		}
		for len(s.packets) > 0 {
			if p, ok := (<-s.packets).(*packet.SetDisplayObjective); ok && (p.ObjectiveName != "InfoPlugin" || p.DisplayName != title) {
				t.Fatal("objective/display mismatch")
			}
		}
	}
}

func TestHealthScoreboardSurvivesOrdinaryScoreboardChanges(t *testing.T) {
	s := &Session{packets: make(chan packet.Packet, 100), closeBackground: make(chan struct{})}
	name := ""
	lines := []string{}
	s.currentScoreboard.Store(&name)
	s.currentLines.Store(&lines)
	check := func(ordinary bool) {
		t.Helper()
		health, normal := 0, false
		for len(s.packets) > 0 {
			pk := <-s.packets
			if _, ok := pk.(*packet.SetTitle); ok {
				t.Fatal("health used title channel")
			}
			if scores, ok := pk.(*packet.SetScore); ok {
				for _, entry := range scores.Entries {
					if entry.IdentityType != protocol.ScoreboardIdentityFakePlayer {
						continue
					}
					if entry.EntryID >= 1001 {
						health++
					}
					if entry.DisplayName == "regular" {
						normal = true
					}
					if entry.ObjectiveName != *s.currentScoreboard.Load() {
						t.Fatal("HUD used stale objective")
					}
				}
			}
		}
		if health != 2 || normal != ordinary {
			t.Fatalf("health=%d ordinary=%v", health, normal)
		}
	}
	s.SendHealthHUD("§C§S§H§B75", "§C§S§H§T75/100")
	check(false)
	board := scoreboard.New("status")
	board.RemovePadding()
	board.WriteString("regular")
	s.SendScoreboard(board)
	check(true)
	s.SendScoreboard(board)
	check(true)
	board = scoreboard.New("renamed status")
	board.RemovePadding()
	board.WriteString("regular")
	s.SendScoreboard(board)
	check(true)
	s.SendHealthHUD("§C§S§H§B85", "§C§S§H§T85/100")
	check(false)
	if len(*s.currentLines.Load()) != 1 {
		t.Fatal("health replaced ordinary lines")
	}
	s.RemoveScoreboard()
	check(false)
	if len(*s.currentLines.Load()) != 0 || *s.currentScoreboard.Load() == "" {
		t.Fatal("removing sidebar removed HUD")
	}
}

func TestHealthHUDWithChangingInfoRowCount(t *testing.T) {
	s := &Session{packets: make(chan packet.Packet, 100), closeBackground: make(chan struct{})}
	name := ""
	lines := []string{}
	s.currentScoreboard.Store(&name)
	s.currentLines.Store(&lines)
	s.SendHealthHUD("§C§S§H§B75", "§C§S§H§T75/100")
	drain := func() {
		for len(s.packets) > 0 {
			<-s.packets
		}
	}
	drain()
	for _, count := range []int{1, 5, 0, 3, 1} {
		before := len(*s.currentLines.Load())
		b := scoreboard.New("리치팜")
		b.RemovePadding()
		for i := 0; i < count; i++ {
			b.WriteString("info")
		}
		s.SendScoreboard(b)
		removed, health := 0, 0
		for len(s.packets) > 0 {
			pk := <-s.packets
			if scores, ok := pk.(*packet.SetScore); ok {
				for _, e := range scores.Entries {
					if e.EntryID < 1001 && e.IdentityType == protocol.ScoreboardIdentityRemove {
						removed++
					}
					if e.EntryID >= 1001 && e.IdentityType == protocol.ScoreboardIdentityFakePlayer {
						health++
					}
				}
			}
		}
		if before > 0 && removed != before {
			t.Fatalf("stale rows: previous %d removed %d", before, removed)
		}
		if health != 2 || len(*s.currentLines.Load()) != count {
			t.Fatalf("health %d rows %d expected %d", health, len(*s.currentLines.Load()), count)
		}
	}
}
