package beanstalk

import "time"

// Tube represents tube Name on the server connected to by Conn.
// It has methods for commands that operate on a single tube.
type Tube struct {
	Conn *Conn
	Name string
}

// Put puts a job into tube `t` with priority pri and TTR ttr, and returns
// the id of the newly-created job. If delay is nonzero, the server will
// wait the given amount of time after returning to the client and before
// putting the job into the ready queue.
func (t *Tube) Put(body []byte, pri uint32, delay, ttr time.Duration) (id uint64, err error) {
	r, err := t.Conn.cmd(t, nil, body, "put", pri, dur(delay), dur(ttr))
	if err != nil {
		return 0, err
	}
	_, err = t.Conn.readResp(r, false, "INSERTED %d", &id)
	if err != nil {
		return 0, err
	}
	return id, nil
}

// PeekReady gets a copy of the job at the front of t's ready queue.
func (t *Tube) PeekReady() (id uint64, body []byte, err error) {
	r, err := t.Conn.cmd(t, nil, nil, "peek-ready")
	if err != nil {
		return 0, nil, err
	}
	body, err = t.Conn.readResp(r, true, "FOUND %d", &id)
	if err != nil {
		return 0, nil, err
	}
	return id, body, nil
}

// PeekDelayed gets a copy of the delayed job that is next to be
// put in t's ready queue.
func (t *Tube) PeekDelayed() (id uint64, body []byte, err error) {
	r, err := t.Conn.cmd(t, nil, nil, "peek-delayed")
	if err != nil {
		return 0, nil, err
	}
	body, err = t.Conn.readResp(r, true, "FOUND %d", &id)
	if err != nil {
		return 0, nil, err
	}
	return id, body, nil
}

// PeekBuried gets a copy of the job in the holding area that would
// be kicked next by Kick.
func (t *Tube) PeekBuried() (id uint64, body []byte, err error) {
	r, err := t.Conn.cmd(t, nil, nil, "peek-buried")
	if err != nil {
		return 0, nil, err
	}
	body, err = t.Conn.readResp(r, true, "FOUND %d", &id)
	if err != nil {
		return 0, nil, err
	}
	return id, body, nil
}

// Kick takes up to bound jobs from the holding area and moves them into
// the ready queue, then returns the number of jobs moved. Jobs will be
// taken in the order in which they were last buried.
func (t *Tube) Kick(bound int) (n int, err error) {
	r, err := t.Conn.cmd(t, nil, nil, "kick", bound)
	if err != nil {
		return 0, err
	}
	_, err = t.Conn.readResp(r, false, "KICKED %d", &n)
	if err != nil {
		return 0, err
	}
	return n, nil
}

// KickJob moves the given job from the holding area into the ready
// queue.
func (t *Tube) KickJob(id uint64) error {
	r, err := t.Conn.cmd(t, nil, nil, "kick-job", id)
	if err != nil {
		return err
	}
	_, err = t.Conn.readResp(r, false, "KICKED")
	return err
}

// FlushReady deletes all jobs residing in the "ready" queue.
func (t *Tube) FlushReady() error {
	return t.flush(t.PeekReady)
}

// FlushDelayed deletes all jobs residing in the "delayed" queue.
func (t *Tube) FlushDelayed() error {
	return t.flush(t.PeekDelayed)
}

// FlushBuried deletes all jobs residing in the holding area.
func (t *Tube) FlushBuried() error {
	return t.flush(t.PeekBuried)
}

type peekFunc func() (uint64, []byte, error)

func (t *Tube) flush(peek peekFunc) error {
	for {
		id, _, err := peek()
		if err != nil {
			if id == 0 {
				return nil
			}
			return err
		}
		if err = t.Conn.Delete(id); err != nil {
			return err
		}
	}
}

type flushFunc func() error

// Flush flushes tube `t` in its entirety.
//
// It is meant as a convenience for calling FlushReady, FlushDelayed, and
// FlushBuried sequentially. If any of those functions fails with an error,
// Flush returns that error.
func (t *Tube) Flush() error {
	for _, flush := range []flushFunc{t.FlushReady, t.FlushDelayed, t.FlushBuried} {
		if err := flush(); err != nil {
			return err
		}
	}
	return nil
}

// Stats retrieves statistics about tube `t`.
func (t *Tube) Stats() (map[string]string, error) {
	r, err := t.Conn.cmd(nil, nil, nil, "stats-tube", t.Name)
	if err != nil {
		return nil, err
	}
	body, err := t.Conn.readResp(r, true, "OK")
	return parseDict(body), err
}

// Pause pauses new reservations in t for time d.
func (t *Tube) Pause(d time.Duration) error {
	r, err := t.Conn.cmd(nil, nil, nil, "pause-tube", t.Name, dur(d))
	if err != nil {
		return err
	}
	_, err = t.Conn.readResp(r, false, "PAUSED")
	if err != nil {
		return err
	}
	return nil
}
