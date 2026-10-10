package world

// AfterTicks schedules an owner callback on the global 20 Hz world loop. Unlike
// block/random ticking, scheduled plugin work also runs in worlds without viewers.
// Callbacks are ephemeral, matching the original plugin scheduler lifetime.
func (tx *Tx) AfterTicks(delay int64, f func(*Tx)) {
	if f == nil {
		return
	}
	if delay < 1 {
		delay = 1
	}
	if tx.w.pluginTasks == nil {
		tx.w.pluginTasks = map[int64][]func(*Tx){}
	}
	due := tx.w.pluginTick + delay
	tx.w.pluginTasks[due] = append(tx.w.pluginTasks[due], f)
}
func (w *World) runPluginTasks(tx *Tx) {
	w.pluginTick++
	tasks := w.pluginTasks[w.pluginTick]
	delete(w.pluginTasks, w.pluginTick)
	for _, f := range tasks {
		f(tx)
	}
}
