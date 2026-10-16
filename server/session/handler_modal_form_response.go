package session

import (
	"encoding/json"
	"fmt"
	"github.com/df-mc/dragonfly/server/player/form"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
	"sync"
	"sync/atomic"
	"time"
)

// ModalFormResponseHandler handles the ModalFormResponse packet.
type ModalFormResponseHandler struct {
	mu        sync.Mutex
	forms     map[uint32]form.Form
	currentID atomic.Uint32
}

// Handle ...
func (h *ModalFormResponseHandler) Handle(p packet.Packet, s *Session, tx *world.Tx, c Controllable) error {
	pk := p.(*packet.ModalFormResponse)

	h.mu.Lock()
	f, ok := h.forms[pk.FormID]
	delete(h.forms, pk.FormID)
	h.mu.Unlock()

	resp, exists := pk.ResponseData.Value()
	if reason, cancelled := pk.CancelReason.Value(); ok && !exists && cancelled && reason == packet.ModalFormCancelReasonUserBusy {
		if retryer, supported := f.(form.BusyRetryer); supported {
			if retry, again := retryer.RetryAfterBusy(); again {
				data, err := json.Marshal(retry)
				if err != nil {
					return err
				}
				go func() {
					timer := time.NewTimer(350 * time.Millisecond)
					defer timer.Stop()
					select {
					case <-s.closeBackground:
						return
					case <-timer.C:
					}
					h.mu.Lock()
					defer h.mu.Unlock()
					// A newer form supersedes this retry, so never reopen an old menu.
					if h.currentID.Load() != pk.FormID {
						return
					}
					h.forms[pk.FormID] = retry
					s.writePacket(&packet.ModalFormRequest{FormID: pk.FormID, FormData: data})
				}()
				return nil
			}
		}
	}
	if !ok && !exists {
		// Sometimes the client seems to send a second response with no data, which would cause the player to be kicked
		// by the server. This should patch that.
		return nil
	}
	if !exists || len(resp) == 0 {
		// The form was cancelled: The cross in the top right corner was clicked.
		resp = nil
	}
	if !ok {
		return fmt.Errorf("no form with ID %v currently opened", pk.FormID)
	}
	if err := f.SubmitJSON(resp, c, tx); err != nil {
		return fmt.Errorf("error submitting form data: %w", err)
	}
	return nil
}
